package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	domain "github.com/mq/api/internal/domain/media"
	"github.com/mq/api/internal/port/inbound"
)

type MediaHandler struct{ media inbound.MediaService }

func NewMediaHandler(media inbound.MediaService) *MediaHandler { return &MediaHandler{media: media} }

func (h *MediaHandler) Sign(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	signed, err := h.media.SignForUser(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, signed)
}

type completeMediaRequest struct {
	PublicID     string `json:"public_id" binding:"required"`
	SecureURL    string `json:"secure_url" binding:"required"`
	ResourceType string `json:"resource_type" binding:"required"`
	Format       string `json:"format" binding:"required"`
	Width        int    `json:"width" binding:"required"`
	Height       int    `json:"height" binding:"required"`
	Bytes        int    `json:"bytes" binding:"required"`
}

func (h *MediaHandler) Complete(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req completeMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	asset, err := h.media.CompleteUpload(c.Request.Context(), userID, domain.Completion{
		PublicID: req.PublicID, SecureURL: req.SecureURL, ResourceType: req.ResourceType,
		Format: req.Format, Width: req.Width, Height: req.Height, Bytes: req.Bytes,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, asset)
}
