package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/onboarding"
)

// List serves the admin queue. With no ?status= it is the pending review queue, as before; approved,
// rejected and all give the decision history.
func (h *OnboardingHandler) List(c *gin.Context) {
	status, ok := onboarding.ParseStatusFilter(c.Query("status"))
	if !ok {
		writeError(c, fmt.Errorf("%w: status must be pending, approved, rejected or all", apperrors.ErrValidation))
		return
	}
	apps, total, err := h.onboarding.List(c.Request.Context(), onboarding.ListFilter{
		Status: status, Limit: queryLimit(c, 50), Offset: queryOffset(c),
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.Header("X-Total-Count", strconv.Itoa(total))
	c.JSON(http.StatusOK, dto.ToOnboardingApplicationResponses(apps))
}

func (h *OnboardingHandler) GetByID(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}

	app, err := h.onboarding.GetByID(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToOnboardingApplicationResponse(*app))
}

func (h *OnboardingHandler) Review(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}

	var req dto.ReviewApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	reviewerID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	app, err := h.onboarding.Review(c.Request.Context(), id, reviewerID, onboarding.ApprovalStatus(req.Status), req.Notes)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToOnboardingApplicationResponse(*app))
}
