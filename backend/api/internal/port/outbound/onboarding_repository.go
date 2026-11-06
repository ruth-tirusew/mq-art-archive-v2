package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/onboarding"
)

type OnboardingRepository interface {
	ListPending(ctx context.Context) ([]onboarding.OnboardingApplication, error)
	GetByID(ctx context.Context, id uuid.UUID) (*onboarding.OnboardingApplication, error)
	GetLatestByApplicantID(ctx context.Context, applicantID uuid.UUID) (*onboarding.OnboardingApplication, error)
	// Save inserts or updates an application. ErrConflict if it would break a uniqueness rule
	// (a handle already requested, or a second pending application for the same applicant).
	Save(ctx context.Context, app onboarding.OnboardingApplication) (*onboarding.OnboardingApplication, error)
	// Transition writes the review fields of app only if the stored status is still `from`, as a
	// single conditional UPDATE, so of two concurrent decisions only one wins. ErrConflict if the
	// status has moved on, ErrNotFound if the application does not exist.
	Transition(ctx context.Context, app onboarding.OnboardingApplication, from onboarding.ApprovalStatus) (*onboarding.OnboardingApplication, error)
}
