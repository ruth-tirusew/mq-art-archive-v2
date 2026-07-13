package outbound

import "context"

type OAuthUserInfo struct {
	ProviderUserID string
	Email          string
	EmailVerified  bool
}

type OAuthProvider interface {
	Name() string
	AuthCodeURL(state string) string
	Exchange(ctx context.Context, code string) (*OAuthUserInfo, error)
}
