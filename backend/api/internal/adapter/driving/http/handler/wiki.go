package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	"github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/domain/wiki"
	"github.com/mq/api/internal/port/inbound"
)

type WikiHandler struct{ wiki inbound.WikiSubmissionService }

func NewWikiHandler(wiki inbound.WikiSubmissionService) *WikiHandler { return &WikiHandler{wiki: wiki} }

type wikiSubmissionRequest struct {
	ArticleID *uuid.UUID `json:"article_id"`
	Title     string     `json:"title" binding:"required"`
	Body      string     `json:"body" binding:"required"`
}

type wikiReviewRequest struct {
	Notes string `json:"notes"`
}

func (h *WikiHandler) Submit(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req wikiSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	submission, err := h.wiki.Submit(c.Request.Context(), userID, req.ArticleID, req.Title, req.Body)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, submission)
}

func (h *WikiHandler) ListMine(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	items, err := h.wiki.ListMine(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *WikiHandler) ListPending(c *gin.Context) {
	items, err := h.wiki.ListPending(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

// ListOwnPending is the artist-facing counterpart to ListPending: submissions awaiting
// review against an article the caller authors.
func (h *WikiHandler) ListOwnPending(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	items, err := h.wiki.ListOwnPending(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

type wikiSubmissionReviewResponse struct {
	Submission *wiki.Submission     `json:"submission"`
	Article    *dto.ArticleResponse `json:"article,omitempty"`
}

// Get and GetOwn return a submission alongside the current state of its target article
// (omitted for a new-article submission), so a caller can render a diff between the
// proposal and what's live. GetOwn is the artist-facing counterpart, restricted to
// submissions against an article the caller owns.
func (h *WikiHandler) Get(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	submission, article, err := h.wiki.GetForReview(c.Request.Context(), id)
	h.writeReviewResponse(c, submission, article, err)
}

func (h *WikiHandler) GetOwn(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	submission, article, err := h.wiki.GetOwnForReview(c.Request.Context(), id, userID)
	h.writeReviewResponse(c, submission, article, err)
}

func (h *WikiHandler) writeReviewResponse(c *gin.Context, submission *wiki.Submission, article *content.Article, err error) {
	if err != nil {
		writeError(c, err)
		return
	}
	resp := wikiSubmissionReviewResponse{Submission: submission}
	if article != nil {
		converted := dto.ToArticleResponse(*article)
		resp.Article = &converted
	}
	c.JSON(http.StatusOK, resp)
}

func (h *WikiHandler) Approve(c *gin.Context) { h.review(c, true, false) }
func (h *WikiHandler) Reject(c *gin.Context)  { h.review(c, false, false) }

// ApproveOwn and RejectOwn are the artist-facing counterparts to Approve/Reject: they
// review a submission against an article the caller owns, rather than requiring the
// admin role. Ownership is enforced by the service layer (apperrors.ErrForbidden if the
// caller doesn't own the submission's target article).
func (h *WikiHandler) ApproveOwn(c *gin.Context) { h.review(c, true, true) }
func (h *WikiHandler) RejectOwn(c *gin.Context)  { h.review(c, false, true) }

func (h *WikiHandler) review(c *gin.Context, approve, ownOnly bool) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	reviewer, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req wikiReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var item any
	switch {
	case approve && ownOnly:
		item, err = h.wiki.ApproveOwn(c.Request.Context(), id, reviewer, req.Notes)
	case approve:
		item, err = h.wiki.Approve(c.Request.Context(), id, reviewer, req.Notes)
	case ownOnly:
		item, err = h.wiki.RejectOwn(c.Request.Context(), id, reviewer, req.Notes)
	default:
		item, err = h.wiki.Reject(c.Request.Context(), id, reviewer, req.Notes)
	}
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}
