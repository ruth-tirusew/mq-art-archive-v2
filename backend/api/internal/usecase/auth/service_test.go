package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	drivenauth "github.com/mq/api/internal/adapter/driven/auth"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/port/inbound"
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
	byID           map[uuid.UUID]identity.User
	byEmail        map[string]identity.User
	passwordHashes map[string]string
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

func (r *stubUserRepo) GetAuthByEmail(_ context.Context, email string) (*identity.User, *string, error) {
	u, ok := r.byEmail[email]
	if !ok {
		return nil, nil, apperrors.ErrNotFound
	}
	hash, hasHash := r.passwordHashes[email]
	if !hasHash {
		return &u, nil, nil
	}
	h := hash
	u.HasPassword = true
	return &u, &h, nil
}

func (r *stubUserRepo) GetAuthByID(_ context.Context, id uuid.UUID) (*identity.User, *string, error) {
	u, ok := r.byID[id]
	if !ok {
		return nil, nil, apperrors.ErrNotFound
	}
	hash, hasHash := r.passwordHashes[u.Email]
	if !hasHash {
		return &u, nil, nil
	}
	h := hash
	u.HasPassword = true
	return &u, &h, nil
}

func (r *stubUserRepo) List(_ context.Context, role *identity.Role, limit, offset int) ([]identity.User, int, error) {
	users := make([]identity.User, 0, len(r.byID))
	for _, user := range r.byID {
		if role == nil || user.Role == *role {
			users = append(users, user)
		}
	}
	return users, len(users), nil
}

func (r *stubUserRepo) CountByRole(_ context.Context, role identity.Role) (int, error) {
	count := 0
	for _, user := range r.byID {
		if user.Role == role {
			count++
		}
	}
	return count, nil
}

func (r *stubUserRepo) Create(_ context.Context, user identity.User) error {
	r.byID[user.ID] = user
	r.byEmail[user.Email] = user
	return nil
}

func (r *stubUserRepo) CreateWithPassword(_ context.Context, user identity.User, passwordHash string) error {
	r.byID[user.ID] = user
	r.byEmail[user.Email] = user
	if r.passwordHashes == nil {
		r.passwordHashes = map[string]string{}
	}
	r.passwordHashes[user.Email] = passwordHash
	return nil
}

func (r *stubUserRepo) UpdatePassword(_ context.Context, id uuid.UUID, passwordHash string) error {
	u, ok := r.byID[id]
	if !ok {
		return apperrors.ErrNotFound
	}
	if r.passwordHashes == nil {
		r.passwordHashes = map[string]string{}
	}
	r.passwordHashes[u.Email] = passwordHash
	return nil
}

func (r *stubUserRepo) UpdateEmail(_ context.Context, id uuid.UUID, email string) error {
	u, ok := r.byID[id]
	if !ok {
		return apperrors.ErrNotFound
	}
	delete(r.byEmail, u.Email)
	if r.passwordHashes != nil {
		if hash, ok := r.passwordHashes[u.Email]; ok {
			delete(r.passwordHashes, u.Email)
			r.passwordHashes[email] = hash
		}
	}
	u.Email = email
	r.byID[id] = u
	r.byEmail[email] = u
	return nil
}

func (r *stubUserRepo) UpdateProfile(_ context.Context, id uuid.UUID, displayName, avatarURL string) error {
	u, ok := r.byID[id]
	if !ok {
		return apperrors.ErrNotFound
	}
	u.DisplayName = displayName
	u.AvatarURL = avatarURL
	r.byID[id] = u
	r.byEmail[u.Email] = u
	return nil
}

func (r *stubUserRepo) UpdateRole(_ context.Context, id uuid.UUID, role identity.Role) error {
	u, ok := r.byID[id]
	if !ok {
		return apperrors.ErrNotFound
	}
	u.Role = role
	r.byID[id] = u
	r.byEmail[u.Email] = u
	return nil
}

func (r *stubUserRepo) PromoteToArtist(ctx context.Context, id uuid.UUID) error {
	return r.UpdateRole(ctx, id, identity.RoleArtist)
}

func (r *stubUserRepo) PromoteToInstitution(ctx context.Context, id uuid.UUID) error {
	return r.UpdateRole(ctx, id, identity.RoleInstitution)
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

type stubPasswordHasher struct{}

func (stubPasswordHasher) Hash(password string) (string, error) {
	return "hash:" + password, nil
}

func (stubPasswordHasher) Check(hash, password string) bool {
	return hash == "hash:"+password
}

type stubMailer struct{ sent int }

func (m *stubMailer) Send(context.Context, string, string, string) error {
	m.sent++
	return nil
}

type stubVerificationRepo struct {
	created  int
	consumed string
}

func (r *stubVerificationRepo) Create(context.Context, outbound.EmailVerificationToken) error {
	r.created++
	return nil
}
func (r *stubVerificationRepo) Consume(_ context.Context, hash string, _ time.Time) (uuid.UUID, error) {
	r.consumed = hash
	return uuid.New(), nil
}

func newAuthService(
	users outbound.UserRepository,
	oauthAccounts outbound.OAuthAccountRepository,
	state outbound.OAuthStateSigner,
	providers []outbound.OAuthProvider,
	origins []string,
) inbound.AuthService {
	return authuc.NewService(
		users,
		oauthAccounts,
		stubTokenIssuer{},
		stubPasswordHasher{},
		state,
		providers,
		origins,
		nil,
		nil,
		nil,
		"http://localhost:5173",
	)
}

func TestService_RegisterCreatesEmailVerification(t *testing.T) {
	users := &stubUserRepo{byID: map[uuid.UUID]identity.User{}, byEmail: map[string]identity.User{}}
	mailer := &stubMailer{}
	verification := &stubVerificationRepo{}
	svc := authuc.NewService(users, &stubOAuthAccountRepo{}, stubTokenIssuer{}, stubPasswordHasher{}, nil, nil, nil,
		mailer, nil, nil, "http://localhost:5173", verification)

	if _, _, err := svc.Register(t.Context(), "new@example.com", "password1"); err != nil {
		t.Fatalf("register: %v", err)
	}
	assist.Equal(t, 1, verification.created)
	assist.Equal(t, 1, mailer.sent)
}

func TestService_UpdateUserRoleProtectsLastAdmin(t *testing.T) {
	adminID := uuid.New()
	admin := identity.User{ID: adminID, Email: "admin@example.com", Role: identity.RoleAdmin}
	users := &stubUserRepo{
		byID:    map[uuid.UUID]identity.User{adminID: admin},
		byEmail: map[string]identity.User{admin.Email: admin},
	}
	svc := newAuthService(users, &stubOAuthAccountRepo{}, nil, nil, nil)
	if _, err := svc.UpdateUserRole(t.Context(), uuid.New(), adminID, identity.RolePublic); err == nil {
		t.Fatal("expected last-admin demotion to be rejected")
	}
}

func TestService_CompleteOAuth_createsUser(t *testing.T) {
	stateSvc := drivenauth.NewOAuthStateService("secret", 10*time.Minute)
	state, err := stateSvc.Sign("http://localhost:5173/auth/callback")
	assist.NoError(t, err)

	users := &stubUserRepo{byID: map[uuid.UUID]identity.User{}, byEmail: map[string]identity.User{}}
	oauthAccounts := &stubOAuthAccountRepo{}
	svc := newAuthService(
		users,
		oauthAccounts,
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

	svc := newAuthService(
		&stubUserRepo{byID: map[uuid.UUID]identity.User{}, byEmail: map[string]identity.User{}},
		&stubOAuthAccountRepo{},
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
	svc := newAuthService(
		&stubUserRepo{byID: map[uuid.UUID]identity.User{}, byEmail: map[string]identity.User{}},
		&stubOAuthAccountRepo{},
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

	svc := newAuthService(users, oauthAccounts, stateSvc,
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
	svc := newAuthService(
		&stubUserRepo{byID: map[uuid.UUID]identity.User{}, byEmail: map[string]identity.User{}},
		&stubOAuthAccountRepo{},
		drivenauth.NewOAuthStateService("secret", 10*time.Minute),
		[]outbound.OAuthProvider{stubOAuthProvider{name: "google"}},
		[]string{"http://localhost:5173"},
	)

	_, _, _, err := svc.CompleteOAuth(t.Context(), "google", "code", "bad-state")
	assist.Error(t, err)
	assist.ErrorIs(t, err, apperrors.ErrInvalidToken)
}

func TestService_Register_success(t *testing.T) {
	users := &stubUserRepo{byID: map[uuid.UUID]identity.User{}, byEmail: map[string]identity.User{}}
	svc := newAuthService(users, &stubOAuthAccountRepo{}, drivenauth.NewOAuthStateService("secret", 10*time.Minute), nil, nil)

	token, user, err := svc.Register(t.Context(), "  New@Example.com ", "password1")
	assist.NoError(t, err)
	assist.Equal(t, "access-token", token)
	assist.Equal(t, "new@example.com", user.Email)
	assist.Equal(t, identity.RolePublic, user.Role)
	assist.Equal(t, "hash:password1", users.passwordHashes["new@example.com"])
}

func TestService_Register_duplicateEmail(t *testing.T) {
	existingID := uuid.New()
	users := &stubUserRepo{
		byID:    map[uuid.UUID]identity.User{existingID: {ID: existingID, Email: "user@example.com"}},
		byEmail: map[string]identity.User{"user@example.com": {ID: existingID, Email: "user@example.com"}},
	}
	svc := newAuthService(users, &stubOAuthAccountRepo{}, drivenauth.NewOAuthStateService("secret", 10*time.Minute), nil, nil)

	_, _, err := svc.Register(t.Context(), "user@example.com", "password1")
	assist.ErrorIs(t, err, apperrors.ErrConflict)
}

func TestService_Register_shortPassword(t *testing.T) {
	users := &stubUserRepo{byID: map[uuid.UUID]identity.User{}, byEmail: map[string]identity.User{}}
	svc := newAuthService(users, &stubOAuthAccountRepo{}, drivenauth.NewOAuthStateService("secret", 10*time.Minute), nil, nil)

	_, _, err := svc.Register(t.Context(), "user@example.com", "short")
	assist.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestService_Login_success(t *testing.T) {
	existingID := uuid.New()
	users := &stubUserRepo{
		byID:           map[uuid.UUID]identity.User{existingID: {ID: existingID, Email: "user@example.com", Role: identity.RolePublic}},
		byEmail:        map[string]identity.User{"user@example.com": {ID: existingID, Email: "user@example.com", Role: identity.RolePublic}},
		passwordHashes: map[string]string{"user@example.com": "hash:password1"},
	}
	svc := newAuthService(users, &stubOAuthAccountRepo{}, drivenauth.NewOAuthStateService("secret", 10*time.Minute), nil, nil)

	token, user, err := svc.Login(t.Context(), "user@example.com", "password1")
	assist.NoError(t, err)
	assist.Equal(t, "access-token", token)
	assist.Equal(t, existingID, user.ID)
}

func TestService_Login_wrongPassword(t *testing.T) {
	existingID := uuid.New()
	users := &stubUserRepo{
		byID:           map[uuid.UUID]identity.User{existingID: {ID: existingID, Email: "user@example.com"}},
		byEmail:        map[string]identity.User{"user@example.com": {ID: existingID, Email: "user@example.com"}},
		passwordHashes: map[string]string{"user@example.com": "hash:password1"},
	}
	svc := newAuthService(users, &stubOAuthAccountRepo{}, drivenauth.NewOAuthStateService("secret", 10*time.Minute), nil, nil)

	_, _, err := svc.Login(t.Context(), "user@example.com", "wrong-password")
	assist.ErrorIs(t, err, apperrors.ErrUnauthorized)
}

func TestService_Login_googleOnlyUser(t *testing.T) {
	existingID := uuid.New()
	users := &stubUserRepo{
		byID:    map[uuid.UUID]identity.User{existingID: {ID: existingID, Email: "user@example.com"}},
		byEmail: map[string]identity.User{"user@example.com": {ID: existingID, Email: "user@example.com"}},
	}
	svc := newAuthService(users, &stubOAuthAccountRepo{}, drivenauth.NewOAuthStateService("secret", 10*time.Minute), nil, nil)

	_, _, err := svc.Login(t.Context(), "user@example.com", "password1")
	assist.ErrorIs(t, err, apperrors.ErrUnauthorized)
}

func TestService_CompleteOAuth_linksPasswordUser(t *testing.T) {
	stateSvc := drivenauth.NewOAuthStateService("secret", 10*time.Minute)
	state, _ := stateSvc.Sign("http://localhost:5173/auth/callback")

	existingID := uuid.New()
	users := &stubUserRepo{
		byID:           map[uuid.UUID]identity.User{existingID: {ID: existingID, Email: "user@example.com", Role: identity.RolePublic}},
		byEmail:        map[string]identity.User{"user@example.com": {ID: existingID, Email: "user@example.com", Role: identity.RolePublic}},
		passwordHashes: map[string]string{"user@example.com": "hash:password1"},
	}
	oauthAccounts := &stubOAuthAccountRepo{}

	svc := newAuthService(users, oauthAccounts, stateSvc,
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
