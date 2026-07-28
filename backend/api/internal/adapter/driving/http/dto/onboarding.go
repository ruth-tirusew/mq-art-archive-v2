package dto

import (
	"time"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/onboarding"
)

type ReviewApplicationRequest struct {
	Status string `json:"status" binding:"required"`
	Notes  string `json:"notes"`
}

type SubmitApplicationRequest struct {
	ApplicantType   string `json:"applicant_type" binding:"required"`
	DisplayName     string `json:"display_name" binding:"required"`
	RequestedHandle string `json:"requested_handle"`
	Notes           string `json:"notes"`
}

type OnboardingApplicationResponse struct {
	ID              uuid.UUID  `json:"id"`
	ApplicantID     uuid.UUID  `json:"applicant_id"`
	ApplicantType   string     `json:"applicant_type"`
	DisplayName     string     `json:"display_name"`
	RequestedHandle string     `json:"requested_handle,omitempty"`
	Notes           string     `json:"notes,omitempty"`
	Status          string     `json:"status"`
	ReviewedBy      *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func ToOnboardingApplicationResponse(app domain.OnboardingApplication) OnboardingApplicationResponse {
	return OnboardingApplicationResponse{
		ID:              app.ID,
		ApplicantID:     app.ApplicantID,
		ApplicantType:   string(app.ApplicantType),
		DisplayName:     app.DisplayName,
		RequestedHandle: app.RequestedHandle,
		Notes:           app.Notes,
		Status:          string(app.Status),
		ReviewedBy:      app.ReviewedBy,
		ReviewedAt:      app.ReviewedAt,
		CreatedAt:       app.CreatedAt,
		UpdatedAt:       app.UpdatedAt,
	}
}

func ToOnboardingApplicationResponses(apps []domain.OnboardingApplication) []OnboardingApplicationResponse {
	out := make([]OnboardingApplicationResponse, len(apps))
	for i, app := range apps {
		out[i] = ToOnboardingApplicationResponse(app)
	}
	return out
}
