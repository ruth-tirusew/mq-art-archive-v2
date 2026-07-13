package identity

import (
	"context"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

type Service struct {
	users outbound.UserRepository
}

func NewService(users outbound.UserRepository) inbound.IdentityService {
	return &Service{users: users}
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.users.GetByID(ctx, id)
}
