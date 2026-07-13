package institution

import (
	"context"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/institution"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

type Service struct {
	institutions outbound.InstitutionRepository
}

func NewService(institutions outbound.InstitutionRepository) inbound.InstitutionService {
	return &Service{institutions: institutions}
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*domain.Institution, error) {
	return s.institutions.GetBySlug(ctx, slug)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Institution, error) {
	return s.institutions.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, inst domain.Institution) (*domain.Institution, error) {
	return s.institutions.Save(ctx, inst)
}
