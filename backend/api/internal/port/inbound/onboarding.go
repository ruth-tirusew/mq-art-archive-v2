package inbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/onboarding"
)

type OnboardingService interface {
	ListPending(ctx context.Context) ([]onboarding.OnboardingApplication, error)
	GetByID(ctx context.Context, id uuid.UUID) (*onboarding.OnboardingApplication, error)
	Review(ctx context.Context, id uuid.UUID, reviewerID uuid.UUID, status onboarding.ApprovalStatus, notes string) (*onboarding.OnboardingApplication, error)
}
