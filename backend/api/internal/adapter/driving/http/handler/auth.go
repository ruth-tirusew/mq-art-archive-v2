package handler

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/port/inbound"
)

type AuthHandler struct {
	auth       inbound.AuthService
	cookieName string
}

func NewAuthHandler(auth inbound.AuthService, cookieName string) *AuthHandler {
	return &AuthHandler{auth: auth, cookieName: cookieName}
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	returnTo := c.Query("return_to")
	if returnTo == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "return_to is required"})
		return
	}

	authURL, state, err := h.auth.BeginOAuth(c.Request.Context(), "google", returnTo)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	// Remember which browser started this sign-in. The callback only proceeds in that same
	// browser, so a link an attacker crafts (their own code and state) cannot sign a victim in
	// to the attacker's account.
	h.setOAuthStateCookie(c, state)
	c.Redirect(http.StatusFound, authURL)
}

// GoogleLink starts connecting Google to the signed-in user's account.
func (h *AuthHandler) GoogleLink(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	returnTo := c.Query("return_to")
	if returnTo == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "return_to is required"})
		return
	}
	authURL, err := h.auth.BeginOAuthLink(c.Request.Context(), userID, "google", returnTo)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.Redirect(http.StatusFound, authURL)
}

func (h *AuthHandler) UnlinkGoogle(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	err = h.auth.UnlinkOAuth(c.Request.Context(), userID, "google")
	switch {
	case err == nil:
		c.Status(http.StatusNoContent)
	case errors.Is(err, apperrors.ErrConflict):
		c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "set a password before disconnecting Google, or you would lose access"})
	case errors.Is(err, apperrors.ErrNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "google is not linked"})
	default:
		writeAuthError(c, err)
	}
}

// redirectWithParams sends the browser back to a (validated) return URL with result flags.
func redirectWithParams(c *gin.Context, returnTo string, params map[string]string) {
	u, err := url.Parse(returnTo)
	if err != nil {
		c.Redirect(http.StatusFound, returnTo)
		return
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	c.Redirect(http.StatusFound, u.String())
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "missing code or state"})
		return
	}

	if h.auth.IsOAuthLinkState(state) {
		h.completeGoogleLink(c, code, state)
		return
	}

	bound, cookieErr := c.Cookie(oauthStateCookie)
	h.clearOAuthStateCookie(c)
	if cookieErr != nil || subtle.ConstantTimeCompare([]byte(bound), []byte(state)) != 1 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "sign-in expired or was started in another browser; please try again"})
		return
	}

	token, _, returnTo, err := h.auth.CompleteOAuth(c.Request.Context(), "google", code, state)
	if err != nil {
		if errors.Is(err, apperrors.ErrConflict) && returnTo != "" {
			// An account with this email exists: send them back to sign in the normal way.
			redirectWithParams(c, returnTo, map[string]string{"error": "account_exists"})
			return
		}
		writeAuthError(c, err)
		return
	}

	h.setAuthCookie(c, token)
	c.Redirect(http.StatusFound, returnTo)
}

func (h *AuthHandler) completeGoogleLink(c *gin.Context, code, state string) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "sign in before linking an account"})
		return
	}
	returnTo, err := h.auth.CompleteOAuthLink(c.Request.Context(), "google", code, state, userID)
	if returnTo == "" {
		writeAuthError(c, err)
		return
	}
	switch {
	case err == nil:
		redirectWithParams(c, returnTo, map[string]string{"google": "linked"})
	case errors.Is(err, apperrors.ErrConflict):
		redirectWithParams(c, returnTo, map[string]string{"google": "error", "reason": "already_linked"})
	case errors.Is(err, apperrors.ErrUnauthorized):
		redirectWithParams(c, returnTo, map[string]string{"google": "error", "reason": "not_verified"})
	default:
		redirectWithParams(c, returnTo, map[string]string{"google": "error", "reason": "failed"})
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.CredentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request body"})
		return
	}

	token, user, err := h.auth.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	h.setAuthCookie(c, token)
	c.JSON(http.StatusCreated, toUserResponse(user))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.CredentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request body"})
		return
	}

	token, user, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	h.setAuthCookie(c, token)
	c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *AuthHandler) Logout(c *gin.Context) {
	h.clearAuthCookie(c)
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request body"})
		return
	}
	if err := h.auth.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		writeAuthError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request body"})
		return
	}
	if err := h.auth.ResetPassword(c.Request.Context(), req.Token, req.Password); err != nil {
		writeAuthError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request body"})
		return
	}
	if err := h.auth.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		writeAuthError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) ResendEmailVerification(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	if err := h.auth.ResendEmailVerification(c.Request.Context(), userID); err != nil {
		writeAuthError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	user, err := h.auth.GetMe(c.Request.Context(), userID)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *AuthHandler) UpdateMyProfile(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request body"})
		return
	}
	user, err := h.auth.UpdateMyProfile(c.Request.Context(), userID, req.DisplayName, req.AvatarURL)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *AuthHandler) ChangeEmail(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	var req dto.ChangeEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request body"})
		return
	}
	user, err := h.auth.ChangeEmail(c.Request.Context(), userID, req.Email, req.CurrentPassword)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request body"})
		return
	}
	if err := h.auth.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		writeAuthError(c, err)
		return
	}
	// Changing the password revoked every older session, including this one. Keep the person who
	// just changed it signed in on this device; every other device has to sign in again.
	if token, err := h.auth.ReissueSession(c.Request.Context(), userID); err == nil {
		h.setAuthCookie(c, token)
	}
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) GetNotifications(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	prefs, err := h.auth.GetNotificationPreferences(c.Request.Context(), userID)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, toNotificationResponse(prefs))
}

func (h *AuthHandler) UpdateNotifications(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	var req dto.NotificationPreferencesResponse
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request body"})
		return
	}
	prefs, err := h.auth.UpdateNotificationPreferences(c.Request.Context(), userID, identity.NotificationPreferences{
		EmailOnNewApplication:   req.EmailOnNewApplication,
		EmailOnEventSyncSummary: req.EmailOnEventSyncSummary,
		NewsletterEnabled:       req.NewsletterEnabled,
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, toNotificationResponse(prefs))
}

func toUserResponse(user *identity.User) dto.UserResponse {
	return dto.UserResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		Role:          string(user.Role),
		DisplayName:   user.DisplayName,
		AvatarURL:     user.AvatarURL,
		HasPassword:   user.HasPassword,
		GoogleLinked:  user.GoogleLinked,
		EmailVerified: user.EmailVerifiedAt != nil,
	}
}

func toNotificationResponse(prefs *identity.NotificationPreferences) dto.NotificationPreferencesResponse {
	return dto.NotificationPreferencesResponse{
		EmailOnNewApplication:   prefs.EmailOnNewApplication,
		EmailOnEventSyncSummary: prefs.EmailOnEventSyncSummary,
		NewsletterEnabled:       prefs.NewsletterEnabled,
	}
}

const oauthStateCookie = "mq_oauth_state"

// The state cookie is only ever read on a top-level GET navigation back from Google, which sends
// Lax cookies, so it does not need SameSite=None. Scoped to the OAuth routes, short-lived.
func (h *AuthHandler) setOAuthStateCookie(c *gin.Context, state string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oauthStateCookie, state, 600, "/api/v1/auth/google", "", authCookieSecure(http.SameSiteLaxMode), true)
}

func (h *AuthHandler) clearOAuthStateCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oauthStateCookie, "", -1, "/api/v1/auth/google", "", authCookieSecure(http.SameSiteLaxMode), true)
}

func (h *AuthHandler) setAuthCookie(c *gin.Context, token string) {
	sameSite := authCookieSameSite()
	secure := authCookieSecure(sameSite)
	c.SetSameSite(sameSite)
	c.SetCookie(h.cookieName, token, 0, "/", "", secure, true)
}

func (h *AuthHandler) clearAuthCookie(c *gin.Context) {
	sameSite := authCookieSameSite()
	secure := authCookieSecure(sameSite)
	c.SetSameSite(sameSite)
	c.SetCookie(h.cookieName, "", -1, "/", "", secure, true)
}

func authCookieSameSite() http.SameSite {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("AUTH_COOKIE_SAME_SITE"))) {
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	case "lax":
		return http.SameSiteLaxMode
	default:
		if strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production") {
			return http.SameSiteNoneMode
		}
		return http.SameSiteLaxMode
	}
}

func authCookieSecure(sameSite http.SameSite) bool {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("AUTH_COOKIE_SECURE")), "true") {
		return true
	}
	if sameSite == http.SameSiteNoneMode {
		return true
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production") {
		return true
	}
	return false
}

func writeAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperrors.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: apperrors.ErrUnauthorized.Error()})
	case errors.Is(err, apperrors.ErrForbidden):
		c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: apperrors.ErrForbidden.Error()})
	case errors.Is(err, apperrors.ErrConflict):
		c.JSON(http.StatusConflict, dto.ErrorResponse{Error: apperrors.ErrConflict.Error()})
	case errors.Is(err, apperrors.ErrValidation):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, apperrors.ErrInvalidToken):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: apperrors.ErrInvalidToken.Error()})
	default:
		writeError(c, err)
	}
}
