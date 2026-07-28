package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/identity"
)

type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*identity.User, error)
	GetByEmail(ctx context.Context, email string) (*identity.User, error)
	// GetAuthByEmail returns the user and optional password hash (nil if Google-only).
	GetAuthByEmail(ctx context.Context, email string) (*identity.User, *string, error)
	// GetAuthByID returns the user and optional password hash.
	GetAuthByID(ctx context.Context, id uuid.UUID) (*identity.User, *string, error)
	Create(ctx context.Context, user identity.User) error
	CreateWithPassword(ctx context.Context, user identity.User, passwordHash string) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	UpdateEmail(ctx context.Context, id uuid.UUID, email string) error
	UpdateProfile(ctx context.Context, id uuid.UUID, displayName, avatarURL string) error
	UpdateRole(ctx context.Context, id uuid.UUID, role identity.Role) error
	PromoteToArtist(ctx context.Context, id uuid.UUID) error
	PromoteToInstitution(ctx context.Context, id uuid.UUID) error
}
