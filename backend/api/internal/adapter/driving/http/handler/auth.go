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

func (h *AuthHandler) Logout(c *gin.Context) {
	h.clearAuthCookie(c)
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	role, err := requestauth.UserRoleFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	email, _ := c.Get(requestauth.ContextUserEmail)
	emailStr, _ := email.(string)

	c.JSON(http.StatusOK, dto.UserResponse{
		ID:    userID.String(),
		Email: emailStr,
		Role:  string(role),
	})
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
	case errors.Is(err, apperrors.ErrInvalidToken):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	default:
		writeError(c, err)
	}
}
