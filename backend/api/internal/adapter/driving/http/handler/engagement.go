package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

type highlightResponse struct {
	ID         uuid.UUID `json:"id"`
	QuotedText string    `json:"quoted_text"`
	Start      int       `json:"start"`
	End        int       `json:"end"`
	Matched    bool      `json:"matched"`
	CreatedAt  time.Time `json:"created_at"`
}

func toHighlightResponse(h inbound.ResolvedHighlight) highlightResponse {
	return highlightResponse{ID: h.ID, QuotedText: h.QuotedText, Start: h.Start, End: h.End, Matched: h.Matched, CreatedAt: h.CreatedAt}
}

type createHighlightRequest struct {
	QuotedText string `json:"quoted_text" binding:"required"`
	Start      int    `json:"start"`
	End        int    `json:"end" binding:"required"`
}

type popularHighlightResponse struct {
	Start int `json:"start"`
	End   int `json:"end"`
	Count int `json:"count"`
}

func (h *EngagementHandler) CreateHighlight(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	articleID, err := uuidParam(c, "articleId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}
	var req createHighlightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	highlight, err := h.engagement.CreateHighlight(c.Request.Context(), userID, articleID, req.QuotedText, req.Start, req.End)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toHighlightResponse(*highlight))
}

func (h *EngagementHandler) DeleteHighlight(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	highlightID, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.engagement.DeleteHighlight(c.Request.Context(), userID, highlightID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *EngagementHandler) ListMyHighlights(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	articleID, err := uuidParam(c, "articleId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}
	highlights, err := h.engagement.ListMyHighlights(c.Request.Context(), userID, articleID)
	if err != nil {
		writeError(c, err)
		return
	}
	out := make([]highlightResponse, len(highlights))
	for i, hl := range highlights {
		out[i] = toHighlightResponse(hl)
	}
	c.JSON(http.StatusOK, out)
}

// GetPopularHighlight is public: seeing what other readers found worth highlighting
// doesn't require an account.
func (h *EngagementHandler) GetPopularHighlight(c *gin.Context) {
	articleID, err := uuidParam(c, "articleId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}
	popular, err := h.engagement.PopularHighlight(c.Request.Context(), articleID)
	if err != nil {
		writeError(c, err)
		return
	}
	if popular == nil {
		c.JSON(http.StatusOK, nil)
		return
	}
	c.JSON(http.StatusOK, popularHighlightResponse{Start: popular.Start, End: popular.End, Count: popular.Count})
}

type commentResponse struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	AuthorName string    `json:"author_name"`
	Body       string    `json:"body"`
	IsGeneral  bool      `json:"is_general"`
	QuotedText string    `json:"quoted_text,omitempty"`
	Start      int       `json:"start"`
	End        int       `json:"end"`
	Matched    bool      `json:"matched"`
	CreatedAt  time.Time `json:"created_at"`
}

func toCommentResponse(c inbound.ResolvedComment) commentResponse {
	return commentResponse{
		ID: c.ID, UserID: c.UserID, AuthorName: c.AuthorName, Body: c.Body, IsGeneral: c.IsGeneral,
		QuotedText: c.QuotedText, Start: c.Start, End: c.End, Matched: c.Matched, CreatedAt: c.CreatedAt,
	}
}

type createCommentRequest struct {
	Body       string `json:"body" binding:"required"`
	IsGeneral  bool   `json:"is_general"`
	QuotedText string `json:"quoted_text"`
	Start      int    `json:"start"`
	End        int    `json:"end"`
}

func (h *EngagementHandler) CreateComment(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	articleID, err := uuidParam(c, "articleId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}
	var req createCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	comment, err := h.engagement.CreateComment(c.Request.Context(), userID, articleID, req.Body, req.QuotedText, req.Start, req.End, req.IsGeneral)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toCommentResponse(*comment))
}

func (h *EngagementHandler) DeleteComment(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	commentID, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.engagement.DeleteComment(c.Request.Context(), userID, commentID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListComments is public: reading the conversation on a wiki article doesn't require an
// account, only posting to it does.
func (h *EngagementHandler) ListComments(c *gin.Context) {
	articleID, err := uuidParam(c, "articleId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}
	comments, err := h.engagement.ListComments(c.Request.Context(), articleID)
	if err != nil {
		writeError(c, err)
		return
	}
	out := make([]commentResponse, len(comments))
	for i, cm := range comments {
		out[i] = toCommentResponse(cm)
	}
	c.JSON(http.StatusOK, out)
}
