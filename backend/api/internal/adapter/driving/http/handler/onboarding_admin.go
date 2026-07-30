package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	"github.com/mq/api/internal/domain/onboarding"
)

func (h *OnboardingHandler) ListPending(c *gin.Context) {
	apps, err := h.onboarding.ListPending(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
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

	status := onboarding.ApprovalStatus(req.Status)
	if status != onboarding.ApprovalStatusApproved && status != onboarding.ApprovalStatusRejected {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "status must be approved or rejected"})
		return
	}

	reviewerID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	app, err := h.onboarding.Review(c.Request.Context(), id, reviewerID, status, req.Notes)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToOnboardingApplicationResponse(*app))
}
