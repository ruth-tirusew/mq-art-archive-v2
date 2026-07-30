package outbound

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/identity"
)

type UserManagementRepository interface {
	List(ctx context.Context, role *identity.Role, limit, offset int) ([]identity.User, int, error)
	CountByRole(ctx context.Context, role identity.Role) (int, error)
}

type EmailVerifiedUserRepository interface {
	MarkEmailVerified(ctx context.Context, id uuid.UUID, at time.Time) error
}
