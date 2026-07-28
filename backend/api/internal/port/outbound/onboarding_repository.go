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
	Save(ctx context.Context, app onboarding.OnboardingApplication) (*onboarding.OnboardingApplication, error)
}
