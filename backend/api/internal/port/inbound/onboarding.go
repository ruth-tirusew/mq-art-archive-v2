package inbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/onboarding"
)

type OnboardingService interface {
	ListPending(ctx context.Context) ([]onboarding.OnboardingApplication, error)
	GetByID(ctx context.Context, id uuid.UUID) (*onboarding.OnboardingApplication, error)
	GetMyApplication(ctx context.Context, applicantID uuid.UUID) (*onboarding.OnboardingApplication, error)
	Submit(ctx context.Context, applicantID uuid.UUID, applicantType onboarding.ApplicantType, displayName, notes string) (*onboarding.OnboardingApplication, error)
	SubmitWithHandle(ctx context.Context, applicantID uuid.UUID, applicantType onboarding.ApplicantType, displayName, requestedHandle, notes string) (*onboarding.OnboardingApplication, error)
	CheckHandleAvailable(ctx context.Context, handle string) (bool, error)
	Review(ctx context.Context, id uuid.UUID, reviewerID uuid.UUID, status onboarding.ApprovalStatus, notes string) (*onboarding.OnboardingApplication, error)
}
