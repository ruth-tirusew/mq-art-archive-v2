package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	drivenauth "github.com/mq/api/internal/adapter/driven/auth"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/port/outbound"
	"github.com/mq/api/internal/testutil/assist"
	authuc "github.com/mq/api/internal/usecase/auth"
)

type stubOAuthProvider struct {
	name     string
	userInfo *outbound.OAuthUserInfo
}

func (s stubOAuthProvider) Name() string { return s.name }
func (s stubOAuthProvider) AuthCodeURL(state string) string {
	return "https://example.com/auth?state=" + state
}
func (s stubOAuthProvider) Exchange(_ context.Context, _ string) (*outbound.OAuthUserInfo, error) {
	return s.userInfo, nil
}

type stubUserRepo struct {
	byID    map[uuid.UUID]identity.User
	byEmail map[string]identity.User
}

func (r *stubUserRepo) GetByID(_ context.Context, id uuid.UUID) (*identity.User, error) {
	u, ok := r.byID[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return &u, nil
}

func (r *stubUserRepo) GetByEmail(_ context.Context, email string) (*identity.User, error) {
	u, ok := r.byEmail[email]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return &u, nil
}

func (r *stubUserRepo) Create(_ context.Context, user identity.User) error {
	r.byID[user.ID] = user
	r.byEmail[user.Email] = user
	return nil
}

type stubOAuthAccountRepo struct {
	accounts []identity.OAuthAccount
}

func (r *stubOAuthAccountRepo) GetByProviderSubject(_ context.Context, provider, providerUserID string) (*identity.OAuthAccount, error) {
	for _, a := range r.accounts {
		if a.Provider == provider && a.ProviderUserID == providerUserID {
			return &a, nil
		}
	}
	return nil, apperrors.ErrNotFound
}

func (r *stubOAuthAccountRepo) Create(_ context.Context, account identity.OAuthAccount) error {
	r.accounts = append(r.accounts, account)
	return nil
}

type stubTokenIssuer struct{}

func (stubTokenIssuer) Issue(_ context.Context, _ *identity.User) (string, error) {
	return "access-token", nil
}

func TestService_CompleteOAuth_createsUser(t *testing.T) {
	stateSvc := drivenauth.NewOAuthStateService("secret", 10*time.Minute)
	state, err := stateSvc.Sign("http://localhost:5173/auth/callback")
	assist.NoError(t, err)

	users := &stubUserRepo{byID: map[uuid.UUID]identity.User{}, byEmail: map[string]identity.User{}}
	oauthAccounts := &stubOAuthAccountRepo{}
	svc := authuc.NewService(
		users,
		oauthAccounts,
		stubTokenIssuer{},
		stateSvc,
		[]outbound.OAuthProvider{stubOAuthProvider{
			name: "google",
			userInfo: &outbound.OAuthUserInfo{
				ProviderUserID: "google-sub",
				Email:          "new@example.com",
				EmailVerified:  true,
			},
		}},
		[]string{"http://localhost:5173"},
	)

	token, user, returnTo, err := svc.CompleteOAuth(t.Context(), "google", "code", state)
	assist.NoError(t, err)
	assist.Equal(t, "access-token", token)
	assist.Equal(t, identity.RolePublic, user.Role)
	assist.Equal(t, "http://localhost:5173/auth/callback", returnTo)
	assist.Equal(t, 1, len(oauthAccounts.accounts))
}

func TestService_CompleteOAuth_rejectsUnverifiedEmail(t *testing.T) {
	stateSvc := drivenauth.NewOAuthStateService("secret", 10*time.Minute)
	state, _ := stateSvc.Sign("http://localhost:5173/auth/callback")

	svc := authuc.NewService(
		&stubUserRepo{byID: map[uuid.UUID]identity.User{}, byEmail: map[string]identity.User{}},
		&stubOAuthAccountRepo{},
		stubTokenIssuer{},
		stateSvc,
		[]outbound.OAuthProvider{stubOAuthProvider{
			name: "google",
			userInfo: &outbound.OAuthUserInfo{
				ProviderUserID: "google-sub",
				Email:          "new@example.com",
				EmailVerified:  false,
			},
		}},
		[]string{"http://localhost:5173"},
	)

	_, _, _, err := svc.CompleteOAuth(t.Context(), "google", "code", state)
	assist.Error(t, err)
	assist.ErrorIs(t, err, apperrors.ErrUnauthorized)
}

func TestService_BeginOAuth_rejectsReturnTo(t *testing.T) {
	stateSvc := drivenauth.NewOAuthStateService("secret", 10*time.Minute)
	svc := authuc.NewService(
		&stubUserRepo{byID: map[uuid.UUID]identity.User{}, byEmail: map[string]identity.User{}},
		&stubOAuthAccountRepo{},
		stubTokenIssuer{},
		stateSvc,
		[]outbound.OAuthProvider{stubOAuthProvider{name: "google"}},
		[]string{"http://localhost:5173"},
	)

	_, _, err := svc.BeginOAuth(t.Context(), "google", "http://evil.example/auth/callback")
	assist.Error(t, err)
	assist.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestService_CompleteOAuth_linksExistingEmail(t *testing.T) {
	stateSvc := drivenauth.NewOAuthStateService("secret", 10*time.Minute)
	state, _ := stateSvc.Sign("http://localhost:5173/auth/callback")

	existingID := uuid.New()
	users := &stubUserRepo{
		byID:    map[uuid.UUID]identity.User{existingID: {ID: existingID, Email: "user@example.com", Role: identity.RoleArtist}},
		byEmail: map[string]identity.User{"user@example.com": {ID: existingID, Email: "user@example.com", Role: identity.RoleArtist}},
	}
	oauthAccounts := &stubOAuthAccountRepo{}

	svc := authuc.NewService(users, oauthAccounts, stubTokenIssuer{}, stateSvc,
		[]outbound.OAuthProvider{stubOAuthProvider{
			name: "google",
			userInfo: &outbound.OAuthUserInfo{
				ProviderUserID: "google-sub",
				Email:          "user@example.com",
				EmailVerified:  true,
			},
		}},
		[]string{"http://localhost:5173"},
	)

	_, user, _, err := svc.CompleteOAuth(t.Context(), "google", "code", state)
	assist.NoError(t, err)
	assist.Equal(t, existingID, user.ID)
	assist.Equal(t, 1, len(oauthAccounts.accounts))
}

func TestService_CompleteOAuth_invalidState(t *testing.T) {
	svc := authuc.NewService(
		&stubUserRepo{byID: map[uuid.UUID]identity.User{}, byEmail: map[string]identity.User{}},
		&stubOAuthAccountRepo{},
		stubTokenIssuer{},
		drivenauth.NewOAuthStateService("secret", 10*time.Minute),
		[]outbound.OAuthProvider{stubOAuthProvider{name: "google"}},
		[]string{"http://localhost:5173"},
	)

	_, _, _, err := svc.CompleteOAuth(t.Context(), "google", "code", "bad-state")
	assist.Error(t, err)
	assist.ErrorIs(t, err, apperrors.ErrInvalidToken)
}
