package onboarding

import (
	"context"
	"time"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

type Service struct {
	applications outbound.OnboardingRepository
}

func NewService(applications outbound.OnboardingRepository) inbound.OnboardingService {
	return &Service{applications: applications}
}

func (s *Service) ListPending(ctx context.Context) ([]domain.OnboardingApplication, error) {
	return s.applications.ListPending(ctx)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
	return s.applications.GetByID(ctx, id)
}

func (s *Service) Review(ctx context.Context, id uuid.UUID, reviewerID uuid.UUID, status domain.ApprovalStatus, notes string) (*domain.OnboardingApplication, error) {
	app, err := s.applications.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	app.Status = status
	app.Notes = notes
	app.ReviewedBy = &reviewerID
	app.ReviewedAt = &now
	app.UpdatedAt = now

	return s.applications.Save(ctx, *app)
}
