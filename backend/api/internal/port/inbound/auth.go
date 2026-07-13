package inbound

import (
	"context"

	"github.com/mq/api/internal/domain/identity"
)

type AuthService interface {
	BeginOAuth(ctx context.Context, provider, returnTo string) (authURL string, state string, err error)
	CompleteOAuth(ctx context.Context, provider, code, state string) (accessToken string, user *identity.User, returnTo string, err error)
}
