package handler

import (
	"errors"
	"net/http"
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

	authURL, _, err := h.auth.BeginOAuth(c.Request.Context(), "google", returnTo)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.Redirect(http.StatusFound, authURL)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "missing code or state"})
		return
	}

	token, _, returnTo, err := h.auth.CompleteOAuth(c.Request.Context(), "google", code, state)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	h.setAuthCookie(c, token)
	c.Redirect(http.StatusFound, returnTo)
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

func (h *AuthHandler) setAuthCookie(c *gin.Context, token string) {
	secure := strings.EqualFold(os.Getenv("AUTH_COOKIE_SECURE"), "true")
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(h.cookieName, token, 0, "/", "", secure, true)
}

func (h *AuthHandler) clearAuthCookie(c *gin.Context) {
	secure := strings.EqualFold(os.Getenv("AUTH_COOKIE_SECURE"), "true")
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(h.cookieName, "", -1, "/", "", secure, true)
}

func writeAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperrors.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, apperrors.ErrForbidden):
		c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, apperrors.ErrConflict):
		c.JSON(http.StatusConflict, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, apperrors.ErrValidation):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, apperrors.ErrInvalidToken):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	default:
		writeError(c, err)
	}
}
