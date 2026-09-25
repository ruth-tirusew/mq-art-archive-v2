package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	"github.com/mq/api/internal/port/inbound"
)

type DigestHandler struct {
	digest          inbound.DigestService
	telegramBotUser string
}

func NewDigestHandler(digest inbound.DigestService, telegramBotUsername string) *DigestHandler {
	return &DigestHandler{digest: digest, telegramBotUser: telegramBotUsername}
}

// CreateTelegramLink issues a one-time t.me deep link that links the caller's account to
// whichever Telegram chat opens it. Requires the bot to be configured (TELEGRAM_BOT_USERNAME).
func (h *DigestHandler) CreateTelegramLink(c *gin.Context) {
	if h.telegramBotUser == "" {
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{Error: "telegram bot not configured"})
		return
	}
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	token, err := h.digest.CreateTelegramLink(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.TelegramLinkResponse{
		DeepLink: fmt.Sprintf("https://t.me/%s?start=%s", h.telegramBotUser, token),
	})
}

// Unsubscribe handles a click from a digest email's unsubscribe link. It's public (no
// auth) since the token itself is the credential, and it always reports success — see
// Service.Unsubscribe for why an invalid token isn't distinguished from "already unsubscribed".
func (h *DigestHandler) Unsubscribe(c *gin.Context) {
	token := c.Query("token")
	if err := h.digest.Unsubscribe(c.Request.Context(), token); err != nil {
		writeError(c, err)
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, "<!doctype html><html><body style=\"font-family:sans-serif;text-align:center;padding:4rem 1rem\">"+
		"<p>You've been unsubscribed from the Artiv digest.</p></body></html>")
}
