package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	"github.com/mq/api/internal/domain/events"
)

func (h *EventHandler) ListAdmin(c *gin.Context) {
	filter := events.ListFilter{Limit: 100}
	if raw := c.Query("status"); raw != "" && raw != "all" {
		s := events.EventStatus(raw)
		filter.Status = &s
	} else if raw == "" {
		// Default admin list: pending (previous behavior)
		s := events.EventStatusPending
		filter.Status = &s
	}
	if lim, err := strconv.Atoi(c.Query("limit")); err == nil && lim > 0 {
		filter.Limit = lim
	}
	if off, err := strconv.Atoi(c.Query("offset")); err == nil && off > 0 {
		filter.Offset = off
	}

	items, err := h.events.List(c.Request.Context(), filter)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToEventResponses(items))
}

func (h *EventHandler) ListPending(c *gin.Context) {
	items, err := h.events.ListPending(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToEventResponses(items))
}

func (h *EventHandler) GetByID(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}

	event, err := h.events.GetByID(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToEventResponse(*event))
}

func (h *EventHandler) CreateAdmin(c *gin.Context) {
	var req dto.AdminEventWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	event, err := h.events.AdminCreate(c.Request.Context(), req.ToWrite())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.ToEventResponse(*event))
}

func (h *EventHandler) UpdateAdmin(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	var req dto.AdminEventWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	event, err := h.events.AdminUpdateContent(c.Request.Context(), id, req.ToWrite())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToEventResponse(*event))
}

func (h *EventHandler) DeleteAdmin(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	if err := h.events.AdminDelete(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *EventHandler) Review(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}

	var req dto.ReviewEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	status := events.EventStatus(req.Status)
	if status != events.EventStatusApproved && status != events.EventStatusRejected {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "status must be approved or rejected"})
		return
	}

	reviewerID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	event, err := h.events.Review(c.Request.Context(), id, reviewerID, status, req.Notes)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToEventResponse(*event))
}

func (h *EventHandler) Sync(c *gin.Context) {
	count, err := h.events.Sync(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SyncEventsResponse{Upserted: count})
}
