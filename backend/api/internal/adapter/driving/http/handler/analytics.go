package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	"github.com/mq/api/internal/port/inbound"
)

type AnalyticsHandler struct{ analytics inbound.AnalyticsService }

func NewAnalyticsHandler(service inbound.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analytics: service}
}

type viewRequest struct {
	EntityType string    `json:"entity_type" binding:"required"`
	EntityID   uuid.UUID `json:"entity_id" binding:"required"`
}

func (h *AnalyticsHandler) Record(c *gin.Context) {
	visitorID, err := c.Cookie("mq_vid")
	if err != nil || visitorID == "" {
		var raw [24]byte
		if _, err := rand.Read(raw[:]); err != nil {
			writeError(c, err)
			return
		}
		visitorID = hex.EncodeToString(raw[:])
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie("mq_vid", visitorID, 365*24*60*60, "/", "", c.Request.TLS != nil, true)
	}
	var req viewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	recorded, err := h.analytics.RecordView(c.Request.Context(), visitorID, req.EntityType, req.EntityID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"recorded": recorded})
}

func (h *AnalyticsHandler) Query(c *gin.Context) {
	var entityID *uuid.UUID
	if raw := c.Query("entity_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entity_id"})
			return
		}
		entityID = &parsed
	}
	from, to, ok := analyticsRange(c)
	if !ok {
		return
	}
	views, err := h.analytics.Query(c.Request.Context(), c.Query("entity_type"), entityID, from, to)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, views)
}

func (h *AnalyticsHandler) MeOverview(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	from, to, ok := analyticsRange(c)
	if !ok {
		return
	}
	views, err := h.analytics.MeOverview(c.Request.Context(), userID, from, to)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, views)
}

func analyticsRange(c *gin.Context) (time.Time, time.Time, bool) {
	to := time.Now().UTC()
	from := to.AddDate(0, -1, 0)
	if raw := c.Query("from"); raw != "" {
		parsed, err := time.Parse("2006-01-02", raw)
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid from"})
			return time.Time{}, time.Time{}, false
		}
		from = parsed
	}
	if raw := c.Query("to"); raw != "" {
		parsed, err := time.Parse("2006-01-02", raw)
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid to"})
			return time.Time{}, time.Time{}, false
		}
		to = parsed
	}
	return from, to, true
}
