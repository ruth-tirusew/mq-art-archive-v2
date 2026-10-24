package outbound

import (
	"context"

	"github.com/google/uuid"

	"github.com/mq/api/internal/domain/identity"
)

type OAuthAccountRepository interface {
	GetByProviderSubject(ctx context.Context, provider, providerUserID string) (*identity.OAuthAccount, error)
	Create(ctx context.Context, account identity.OAuthAccount) error
	GetByUserProvider(ctx context.Context, userID uuid.UUID, provider string) (*identity.OAuthAccount, error)
	DeleteByUserProvider(ctx context.Context, userID uuid.UUID, provider string) error
}
