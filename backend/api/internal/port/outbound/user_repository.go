package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/identity"
)

type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*identity.User, error)
	GetByEmail(ctx context.Context, email string) (*identity.User, error)
	Create(ctx context.Context, user identity.User) error
}
