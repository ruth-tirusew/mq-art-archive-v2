package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	"github.com/mq/api/internal/port/inbound"
)

type EngagementHandler struct{ engagement inbound.EngagementService }

func NewEngagementHandler(engagement inbound.EngagementService) *EngagementHandler {
	return &EngagementHandler{engagement: engagement}
}

type favoriteStatusResponse struct {
	Favorited bool `json:"favorited"`
	Count     int  `json:"count"`
}

// ToggleFavorite adds or removes the caller's favorite on the article, whoever they are —
// favoriting isn't restricted to any role, unlike the wiki submission/review endpoints.
func (h *EngagementHandler) ToggleFavorite(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	articleID, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	favorited, count, err := h.engagement.ToggleFavorite(c.Request.Context(), userID, articleID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, favoriteStatusResponse{Favorited: favorited, Count: count})
}

// GetFavoriteStatus is public: an unauthenticated viewer still sees the total count, just
// with favorited always false.
func (h *EngagementHandler) GetFavoriteStatus(c *gin.Context) {
	articleID, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	viewerID, _ := requestauth.UserIDFromContext(c) // zero value (uuid.Nil) if unauthenticated
	favorited, count, err := h.engagement.FavoriteStatus(c.Request.Context(), viewerID, articleID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, favoriteStatusResponse{Favorited: favorited, Count: count})
}

func (h *EngagementHandler) ListMyFavorites(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	articles, err := h.engagement.ListMyFavorites(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArticleResponses(articles))
}
