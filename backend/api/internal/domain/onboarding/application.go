package onboarding

import (
	"time"

	"github.com/google/uuid"
)

type ApplicantType string

const (
	ApplicantTypeArtist      ApplicantType = "artist"
	ApplicantTypeInstitution ApplicantType = "institution"
)

type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "pending"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusRejected ApprovalStatus = "rejected"
)

type OnboardingApplication struct {
	ID              uuid.UUID
	ApplicantID     uuid.UUID
	ApplicantType   ApplicantType
	DisplayName     string
	RequestedHandle string
	Notes           string
	Status          ApprovalStatus
	ReviewedBy      *uuid.UUID
	ReviewedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
