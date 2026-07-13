package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/identity"
)

type TokenClaims struct {
	UserID uuid.UUID
	Role   identity.Role
}

type TokenIssuer interface {
	Issue(ctx context.Context, user *identity.User) (string, error)
}

type TokenVerifier interface {
	Verify(ctx context.Context, token string) (TokenClaims, error)
}
