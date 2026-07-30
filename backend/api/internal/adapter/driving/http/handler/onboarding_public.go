package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	domain "github.com/mq/api/internal/domain/onboarding"
)

func (h *OnboardingHandler) Submit(c *gin.Context) {
	applicantID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req dto.SubmitApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	app, err := h.onboarding.SubmitWithHandle(
		c.Request.Context(),
		applicantID,
		domain.ApplicantType(req.ApplicantType),
		req.DisplayName,
		req.RequestedHandle,
		req.Notes,
	)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.ToOnboardingApplicationResponse(*app))
}

func (h *OnboardingHandler) GetMyApplication(c *gin.Context) {
	applicantID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	app, err := h.onboarding.GetMyApplication(c.Request.Context(), applicantID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToOnboardingApplicationResponse(*app))
}

func (h *OnboardingHandler) CheckHandleAvailable(c *gin.Context) {
	available, err := h.onboarding.CheckHandleAvailable(c.Request.Context(), c.Param("handle"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"handle": c.Param("handle"), "available": available})
}
