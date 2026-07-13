package outbound

type OAuthStateSigner interface {
	Sign(returnTo string) (string, error)
	Verify(state string) (returnTo string, err error)
}
