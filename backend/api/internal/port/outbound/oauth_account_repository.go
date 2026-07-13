package outbound

import (
	"context"

	"github.com/mq/api/internal/domain/identity"
)

type OAuthAccountRepository interface {
	GetByProviderSubject(ctx context.Context, provider, providerUserID string) (*identity.OAuthAccount, error)
	Create(ctx context.Context, account identity.OAuthAccount) error
}
