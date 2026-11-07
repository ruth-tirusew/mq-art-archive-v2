package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/onboarding"
)

type OnboardingRepository interface {
	ListPending(ctx context.Context) ([]onboarding.OnboardingApplication, error)
	// List returns one page and the total number of matches. The pending queue is oldest first;
	// decided applications are newest decision first.
	List(ctx context.Context, filter onboarding.ListFilter) ([]onboarding.OnboardingApplication, int, error)
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
