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
	ReviewNotes     string     `json:"review_notes,omitempty"`
	Status          string     `json:"status"`
	ReviewedBy      *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	// QuestionnaireRevisionID is the revision the applicant answered; InviteID is set for invited applicants.
	QuestionnaireRevisionID *uuid.UUID `json:"questionnaire_revision_id,omitempty"`
	InviteID                *uuid.UUID `json:"invite_id,omitempty"`
	// Works are the pieces of work submitted with the application, as they were at submission.
	Works []ApplicantWorkResponse `json:"works,omitempty"`
	// Portrait is the applicant's portrait as submitted. Absent on applications from before
	// portraits were required.
	Portrait  *ApplicantPortraitResponse `json:"portrait,omitempty"`
	CreatedAt time.Time                  `json:"created_at"`
	UpdatedAt time.Time                  `json:"updated_at"`
}

func ToOnboardingApplicationResponse(app domain.OnboardingApplication) OnboardingApplicationResponse {
	return OnboardingApplicationResponse{
		ID:                      app.ID,
		ApplicantID:             app.ApplicantID,
		ApplicantType:           string(app.ApplicantType),
		DisplayName:             app.DisplayName,
		RequestedHandle:         app.RequestedHandle,
		Notes:                   app.Notes,
		ReviewNotes:             app.ReviewNotes,
		Status:                  string(app.Status),
		ReviewedBy:              app.ReviewedBy,
		ReviewedAt:              app.ReviewedAt,
		QuestionnaireRevisionID: app.QuestionnaireRevisionID,
		InviteID:                app.InviteID,
		Works:                   ToApplicantWorkResponses(app.Works),
		Portrait:                ToApplicantPortraitResponse(app.Portrait),
		CreatedAt:               app.CreatedAt,
		UpdatedAt:               app.UpdatedAt,
	}
}

func ToOnboardingApplicationResponses(apps []domain.OnboardingApplication) []OnboardingApplicationResponse {
	out := make([]OnboardingApplicationResponse, len(apps))
	for i, app := range apps {
		out[i] = ToOnboardingApplicationResponse(app)
	}
	return out
}

type AddWorkRequest struct {
	MediaAssetID uuid.UUID `json:"media_asset_id" binding:"required"`
	Title        string    `json:"title" binding:"required,max=200"`
	Medium       string    `json:"medium" binding:"max=100"`
	Year         *int      `json:"year"`
}

type ApplicantPortraitResponse struct {
	MediaAssetID uuid.UUID `json:"media_asset_id"`
	URL          string    `json:"url"`
	Blurhash     string    `json:"blurhash,omitempty"`
}

func ToApplicantPortraitResponse(p *domain.ApplicantPortrait) *ApplicantPortraitResponse {
	if p == nil {
		return nil
	}
	return &ApplicantPortraitResponse{MediaAssetID: p.MediaAssetID, URL: p.URL, Blurhash: p.Blurhash}
}

type SetPortraitRequest struct {
	MediaAssetID uuid.UUID `json:"media_asset_id" binding:"required"`
}

type ApplicantWorkResponse struct {
	ID           uuid.UUID `json:"id"`
	MediaAssetID uuid.UUID `json:"media_asset_id"`
	URL          string    `json:"url"`
	Blurhash     string    `json:"blurhash,omitempty"`
	Title        string    `json:"title"`
	Medium       string    `json:"medium,omitempty"`
	Year         *int      `json:"year,omitempty"`
	Position     int       `json:"position"`
}

func ToApplicantWorkResponse(w domain.ApplicantWork) ApplicantWorkResponse {
	return ApplicantWorkResponse{
		ID: w.ID, MediaAssetID: w.MediaAssetID, URL: w.URL, Blurhash: w.Blurhash,
		Title: w.Title, Medium: w.Medium, Year: w.Year, Position: w.Position,
	}
}

func ToApplicantWorkResponses(works []domain.ApplicantWork) []ApplicantWorkResponse {
	if len(works) == 0 {
		return nil
	}
	out := make([]ApplicantWorkResponse, len(works))
	for i, w := range works {
		out[i] = ToApplicantWorkResponse(w)
	}
	return out
}
