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
	ReviewNotes     string
	Status          ApprovalStatus
	ReviewedBy      *uuid.UUID
	ReviewedAt      *time.Time
	// QuestionnaireRevisionID is the revision the applicant answered, so a reviewer sees their
	// answers beside the wording they responded to. Nil for applications made before the
	// questionnaire existed.
	QuestionnaireRevisionID *uuid.UUID
	// QuestionnaireAnswers is a frozen copy of the answers submitted with the application. It is
	// opaque here: the questionnaire context writes and reads it. Nil when none was captured.
	QuestionnaireAnswers []byte
	// Works are the pieces of work the applicant submitted with the application, frozen at that
	// moment. Empty for applications made before works were required.
	Works []ApplicantWork
	// InviteID is set when the applicant arrived through an invite link.
	InviteID  *uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}
