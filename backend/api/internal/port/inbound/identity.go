package inbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/identity"
)

type IdentityService interface {
	GetUser(ctx context.Context, id uuid.UUID) (*identity.User, error)
}
