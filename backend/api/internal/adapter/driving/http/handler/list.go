package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/domain/profile"
	"github.com/mq/api/internal/port/inbound"
)

func queryLimit(c *gin.Context, defaultLimit int) int {
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultLimit
}

func queryOffset(c *gin.Context) int {
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return n
		}
	}
	return 0
}

func queryBool(c *gin.Context, key string) *bool {
	v := c.Query(key)
	if v == "" {
		return nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return nil
	}
	return &b
}

func queryInt(c *gin.Context, key string) *int {
	v := c.Query(key)
	if v == "" {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil
	}
	return &n
}

func profileListFilter(c *gin.Context) profile.ListFilter {
	return profile.ListFilter{
		Query:    c.Query("q"),
		Limit:    queryLimit(c, 50),
		Offset:   queryOffset(c),
		Featured: queryBool(c, "featured"),
	}
}

func artListFilter(c *gin.Context) art.ListFilter {
	return art.ListFilter{
		City:                c.Query("city"),
		Medium:              c.Query("medium"),
		Year:                queryInt(c, "year"),
		Style:               c.Query("style"),
		FeaturedAcquisition: queryBool(c, "featured"),
		Query:               c.Query("q"),
		Limit:               queryLimit(c, 50),
		Offset:              queryOffset(c),
	}
}

func articleListFilter(c *gin.Context) content.ListFilter {
	return content.ListFilter{
		Category: c.Query("category"),
		Query:    c.Query("q"),
		Limit:    queryLimit(c, 50),
		Offset:   queryOffset(c),
	}
}

func eventListFilter(c *gin.Context) events.ListFilter {
	status := events.EventStatusApproved
	filter := events.ListFilter{
		Status:    &status,
		EventType: c.Query("type"),
		Query:     c.Query("q"),
		Limit:     queryLimit(c, 50),
		Offset:    queryOffset(c),
	}
	if c.Query("upcoming") == "false" {
		filter.UpcomingOnly = false
	} else {
		filter.UpcomingOnly = true
	}
	return filter
}

func (h *ProfileHandler) List(c *gin.Context) {
	profiles, err := h.profiles.ListApproved(c.Request.Context(), profileListFilter(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtistProfileResponses(profiles))
}

func (h *ProfileHandler) GetByHandle(c *gin.Context) {
	profile, err := h.profiles.GetArtistByHandle(c.Request.Context(), c.Param("handle"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtistProfileResponse(*profile))
}

func (h *ArtHandler) List(c *gin.Context) {
	posts, err := h.art.ListPublished(c.Request.Context(), artListFilter(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtPostWithArtistResponses(posts))
}

type EventHandler struct {
	events inbound.EventsService
}

func NewEventHandler(events inbound.EventsService) *EventHandler {
	return &EventHandler{events: events}
}

func (h *EventHandler) List(c *gin.Context) {
	items, err := h.events.List(c.Request.Context(), eventListFilter(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToEventResponses(items))
}

func (h *EventHandler) GetBySlug(c *gin.Context) {
	event, err := h.events.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToEventResponse(*event))
}
