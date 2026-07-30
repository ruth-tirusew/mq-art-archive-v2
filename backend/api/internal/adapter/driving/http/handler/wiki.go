package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
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

func (h *WikiHandler) Approve(c *gin.Context) { h.review(c, true) }
func (h *WikiHandler) Reject(c *gin.Context)  { h.review(c, false) }

func (h *WikiHandler) review(c *gin.Context, approve bool) {
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
	if approve {
		item, err = h.wiki.Approve(c.Request.Context(), id, reviewer, req.Notes)
	} else {
		item, err = h.wiki.Reject(c.Request.Context(), id, reviewer, req.Notes)
	}
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}
