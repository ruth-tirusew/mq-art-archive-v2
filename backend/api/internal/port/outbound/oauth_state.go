package outbound

import "github.com/google/uuid"

type OAuthStateSigner interface {
	Sign(returnTo string) (string, error)
	Verify(state string) (returnTo string, err error)
	// SignLink / VerifyLink carry a state for a signed-in user connecting a provider to their
	// own account. Link states never verify as sign-in states and vice versa.
	SignLink(returnTo string, userID uuid.UUID) (string, error)
	VerifyLink(state string) (returnTo string, userID uuid.UUID, err error)
}
