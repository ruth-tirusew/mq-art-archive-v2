package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

const minPasswordLength = 8
const resetTokenTTL = time.Hour
const emailVerificationTTL = 24 * time.Hour

type Service struct {
	users          outbound.UserRepository
	userManagement outbound.UserManagementRepository
	verifiedUsers  outbound.EmailVerifiedUserRepository
	oauthAccounts  outbound.OAuthAccountRepository
	tokens         outbound.TokenIssuer
	passwords      outbound.PasswordHasher
	state          outbound.OAuthStateSigner
	providers      map[string]outbound.OAuthProvider
	allowedReturn  map[string]struct{}
	mailer         outbound.Mailer
	resets         outbound.PasswordResetRepository
	notifications  outbound.NotificationPreferencesRepository
	webAppURL      string
	verifications  outbound.EmailVerificationRepository

	dummyOnce sync.Once
	dummyHash string
}

func NewService(
	users outbound.UserRepository,
	oauthAccounts outbound.OAuthAccountRepository,
	tokens outbound.TokenIssuer,
	passwords outbound.PasswordHasher,
	state outbound.OAuthStateSigner,
	providers []outbound.OAuthProvider,
	allowedReturnOrigins []string,
	mailer outbound.Mailer,
	resets outbound.PasswordResetRepository,
	notifications outbound.NotificationPreferencesRepository,
	webAppURL string,
	verification ...outbound.EmailVerificationRepository,
) inbound.AuthService {
	providerMap := make(map[string]outbound.OAuthProvider, len(providers))
	for _, p := range providers {
		providerMap[p.Name()] = p
	}

	allowed := make(map[string]struct{}, len(allowedReturnOrigins))
	for _, origin := range allowedReturnOrigins {
		allowed[strings.TrimSpace(origin)] = struct{}{}
	}

	service := &Service{
		users:         users,
		oauthAccounts: oauthAccounts,
		tokens:        tokens,
		passwords:     passwords,
		state:         state,
		providers:     providerMap,
		allowedReturn: allowed,
		mailer:        mailer,
		resets:        resets,
		notifications: notifications,
		webAppURL:     strings.TrimRight(webAppURL, "/"),
	}
	service.userManagement, _ = users.(outbound.UserManagementRepository)
	service.verifiedUsers, _ = users.(outbound.EmailVerifiedUserRepository)
	if len(verification) > 0 {
		service.verifications = verification[0]
	}
	return service
}

func (s *Service) BeginOAuth(_ context.Context, provider, returnTo string) (string, string, error) {
	p, ok := s.providers[provider]
	if !ok {
		return "", "", fmt.Errorf("unknown provider: %s", provider)
	}
	if err := s.validateReturnTo(returnTo); err != nil {
		return "", "", err
	}

	state, err := s.state.Sign(returnTo)
	if err != nil {
		return "", "", err
	}
	return p.AuthCodeURL(state), state, nil
}

func (s *Service) CompleteOAuth(ctx context.Context, provider, code, state string) (string, *identity.User, string, error) {
	p, ok := s.providers[provider]
	if !ok {
		return "", nil, "", fmt.Errorf("unknown provider: %s", provider)
	}

	returnTo, err := s.state.Verify(state)
	if err != nil {
		return "", nil, "", err
	}
	if err := s.validateReturnTo(returnTo); err != nil {
		return "", nil, "", err
	}

	info, err := p.Exchange(ctx, code)
	if err != nil {
		return "", nil, "", err
	}
	if !info.EmailVerified {
		return "", nil, "", apperrors.ErrUnauthorized
	}

	user, err := s.findOrCreateUser(ctx, provider, info)
	if err != nil {
		// ErrConflict means an account with this email already exists. The caller still needs
		// returnTo to explain what to do next ("sign in, then link Google in settings").
		if errors.Is(err, apperrors.ErrConflict) {
			return "", nil, returnTo, err
		}
		return "", nil, "", err
	}
	if info.EmailVerified && user.EmailVerifiedAt == nil {
		now := time.Now().UTC()
		if s.verifiedUsers != nil {
			if err := s.verifiedUsers.MarkEmailVerified(ctx, user.ID, now); err != nil {
				return "", nil, "", err
			}
		}
		user.EmailVerifiedAt = &now
	}

	token, err := s.tokens.Issue(ctx, user)
	if err != nil {
		return "", nil, "", err
	}
	return token, user, returnTo, nil
}

func (s *Service) BeginOAuthLink(_ context.Context, userID uuid.UUID, provider, returnTo string) (string, error) {
	p, ok := s.providers[provider]
	if !ok {
		return "", fmt.Errorf("unknown provider: %s", provider)
	}
	if err := s.validateReturnTo(returnTo); err != nil {
		return "", err
	}
	state, err := s.state.SignLink(returnTo, userID)
	if err != nil {
		return "", err
	}
	return p.AuthCodeURL(state), nil
}

func (s *Service) IsOAuthLinkState(state string) bool {
	_, _, err := s.state.VerifyLink(state)
	return err == nil
}

// CompleteOAuthLink attaches the provider account to actorID. The state must have been issued
// to that same user, so a link callback replayed in someone else's browser is rejected.
func (s *Service) CompleteOAuthLink(ctx context.Context, provider, code, state string, actorID uuid.UUID) (string, error) {
	p, ok := s.providers[provider]
	if !ok {
		return "", fmt.Errorf("unknown provider: %s", provider)
	}
	returnTo, stateUser, err := s.state.VerifyLink(state)
	if err != nil {
		return "", err
	}
	if err := s.validateReturnTo(returnTo); err != nil {
		return "", err
	}
	if stateUser != actorID {
		return returnTo, apperrors.ErrUnauthorized
	}

	info, err := p.Exchange(ctx, code)
	if err != nil {
		return returnTo, err
	}
	if !info.EmailVerified {
		return returnTo, apperrors.ErrUnauthorized
	}

	if existing, err := s.oauthAccounts.GetByProviderSubject(ctx, provider, info.ProviderUserID); err == nil {
		if existing.UserID == actorID {
			return returnTo, nil // already linked to this user; nothing to do
		}
		return returnTo, apperrors.ErrConflict // belongs to a different Artiv account
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return returnTo, err
	}
	if _, err := s.oauthAccounts.GetByUserProvider(ctx, actorID, provider); err == nil {
		return returnTo, apperrors.ErrConflict // this user already has a different one linked
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return returnTo, err
	}

	if err := s.oauthAccounts.Create(ctx, identity.OAuthAccount{
		UserID:         actorID,
		Provider:       provider,
		ProviderUserID: info.ProviderUserID,
		Email:          info.Email,
	}); err != nil {
		return returnTo, err
	}

	// If the provider's verified email is the account's own email, that also proves ownership.
	if s.verifiedUsers != nil {
		if user, err := s.users.GetByID(ctx, actorID); err == nil && user.EmailVerifiedAt == nil && strings.EqualFold(user.Email, info.Email) {
			_ = s.verifiedUsers.MarkEmailVerified(ctx, actorID, time.Now().UTC())
		}
	}
	return returnTo, nil
}

// UnlinkOAuth disconnects a provider. It requires a password so the user cannot lock themselves
// out of the only way back into the account.
func (s *Service) UnlinkOAuth(ctx context.Context, userID uuid.UUID, provider string) error {
	_, passwordHash, err := s.users.GetAuthByID(ctx, userID)
	if err != nil {
		return err
	}
	if passwordHash == nil || *passwordHash == "" {
		return apperrors.ErrConflict
	}
	return s.oauthAccounts.DeleteByUserProvider(ctx, userID, provider)
}

func (s *Service) Register(ctx context.Context, email, password string) (string, *identity.User, error) {
	normalized, err := normalizeEmail(email)
	if err != nil {
		return "", nil, err
	}
	if err := validatePassword(password); err != nil {
		return "", nil, err
	}

	_, err = s.users.GetByEmail(ctx, normalized)
	if err == nil {
		return "", nil, apperrors.ErrConflict
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		return "", nil, err
	}

	hash, err := s.passwords.Hash(password)
	if err != nil {
		return "", nil, err
	}

	user := identity.User{
		ID:    uuid.New(),
		Email: normalized,
		Role:  identity.RolePublic,
	}
	if err := s.users.CreateWithPassword(ctx, user, hash); err != nil {
		return "", nil, err
	}
	if err := s.sendEmailVerification(ctx, user); err != nil {
		return "", nil, err
	}

	token, err := s.tokens.Issue(ctx, &user)
	if err != nil {
		return "", nil, err
	}
	user.HasPassword = true
	return token, &user, nil
}

// burnPasswordCheck spends one password-hash comparison against a throwaway hash. It is used on
// login paths that have no real hash to check so they take about as long as a real attempt.
func (s *Service) burnPasswordCheck(password string) {
	s.dummyOnce.Do(func() {
		s.dummyHash, _ = s.passwords.Hash("artiv-timing-equaliser")
	})
	if s.dummyHash != "" {
		_ = s.passwords.Check(s.dummyHash, password)
	}
}

func (s *Service) Login(ctx context.Context, email, password string) (string, *identity.User, error) {
	normalized, err := normalizeEmail(email)
	if err != nil {
		return "", nil, apperrors.ErrUnauthorized
	}

	user, passwordHash, err := s.users.GetAuthByEmail(ctx, normalized)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			// Same work and same answer as a wrong password, so response time and status
			// don't reveal whether the email has an account.
			s.burnPasswordCheck(password)
			return "", nil, apperrors.ErrUnauthorized
		}
		return "", nil, err
	}
	if passwordHash == nil || *passwordHash == "" {
		s.burnPasswordCheck(password)
		return "", nil, apperrors.ErrUnauthorized
	}
	if !s.passwords.Check(*passwordHash, password) {
		return "", nil, apperrors.ErrUnauthorized
	}

	token, err := s.tokens.Issue(ctx, user)
	if err != nil {
		return "", nil, err
	}
	user.HasPassword = true
	return token, user, nil
}

func (s *Service) findOrCreateUser(ctx context.Context, provider string, info *outbound.OAuthUserInfo) (*identity.User, error) {
	account, err := s.oauthAccounts.GetByProviderSubject(ctx, provider, info.ProviderUserID)
	if err == nil {
		return s.users.GetByID(ctx, account.UserID)
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}

	// An account with this email already exists (password-registered, or another login). Do not
	// attach this provider to it automatically: anyone can register a password account with
	// someone else's email address, and linking by email would hand that account to the real
	// owner while leaving the registrant's password working. The owner links explicitly from
	// settings while signed in instead.
	if _, err := s.users.GetByEmail(ctx, info.Email); err == nil {
		return nil, apperrors.ErrConflict
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}

	newUser := identity.User{
		ID:    uuid.New(),
		Email: info.Email,
		Role:  identity.RolePublic,
	}
	if err := s.users.Create(ctx, newUser); err != nil {
		return nil, err
	}
	if err := s.oauthAccounts.Create(ctx, identity.OAuthAccount{
		UserID:         newUser.ID,
		Provider:       provider,
		ProviderUserID: info.ProviderUserID,
		Email:          info.Email,
	}); err != nil {
		return nil, err
	}
	return &newUser, nil
}

func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	normalized, err := normalizeEmail(email)
	if err != nil {
		// Still return nil to avoid enumeration via validation differences on malformed emails.
		return nil
	}
	user, err := s.users.GetByEmail(ctx, normalized)
	if err != nil {
		return nil
	}
	if s.resets == nil || s.mailer == nil {
		return nil
	}

	raw, err := randomToken(32)
	if err != nil {
		return err
	}
	hash := hashToken(raw)
	now := time.Now().UTC()
	token := outbound.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: now.Add(resetTokenTTL),
		CreatedAt: now,
	}
	if err := s.resets.Create(ctx, token); err != nil {
		return err
	}

	resetURL := s.webAppURL + "/reset-password?token=" + url.QueryEscape(raw)
	body := "Reset your Artiv password using this link (expires in 1 hour):\n\n" + resetURL + "\n"
	_ = s.mailer.Send(ctx, user.Email, "Reset your Artiv password", body)
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	if s.resets == nil {
		return apperrors.ErrNotImplemented
	}
	record, err := s.resets.GetByHash(ctx, hashToken(token))
	if err != nil {
		return apperrors.ErrInvalidToken
	}
	if record.UsedAt != nil || time.Now().UTC().After(record.ExpiresAt) {
		return apperrors.ErrInvalidToken
	}
	hash, err := s.passwords.Hash(newPassword)
	if err != nil {
		return err
	}
	if err := s.users.UpdatePassword(ctx, record.UserID, hash); err != nil {
		return err
	}
	return s.resets.MarkUsed(ctx, record.ID)
}

func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	if s.verifications == nil || strings.TrimSpace(token) == "" {
		return apperrors.ErrInvalidToken
	}
	if _, err := s.verifications.Consume(ctx, hashToken(token), time.Now().UTC()); err != nil {
		return apperrors.ErrInvalidToken
	}
	return nil
}

func (s *Service) ResendEmailVerification(ctx context.Context, userID uuid.UUID) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.EmailVerifiedAt != nil {
		return nil
	}
	return s.sendEmailVerification(ctx, *user)
}

func (s *Service) sendEmailVerification(ctx context.Context, user identity.User) error {
	if s.verifications == nil || s.mailer == nil {
		return nil
	}
	raw, err := randomToken(32)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if err := s.verifications.Create(ctx, outbound.EmailVerificationToken{
		ID: uuid.New(), UserID: user.ID, TokenHash: hashToken(raw),
		ExpiresAt: now.Add(emailVerificationTTL), CreatedAt: now,
	}); err != nil {
		return err
	}
	link := s.webAppURL + "/verify-email?token=" + url.QueryEscape(raw)
	return s.mailer.Send(ctx, user.Email, "Verify your Artiv email", "Verify your email using this link:\n\n"+link+"\n")
}

func (s *Service) GetMe(ctx context.Context, userID uuid.UUID) (*identity.User, error) {
	user, passwordHash, err := s.users.GetAuthByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.HasPassword = passwordHash != nil && *passwordHash != ""
	if _, err := s.oauthAccounts.GetByUserProvider(ctx, userID, "google"); err == nil {
		user.GoogleLinked = true
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}
	return user, nil
}

func (s *Service) UpdateMyProfile(ctx context.Context, userID uuid.UUID, displayName, avatarURL string) (*identity.User, error) {
	if err := s.users.UpdateProfile(ctx, userID, strings.TrimSpace(displayName), strings.TrimSpace(avatarURL)); err != nil {
		return nil, err
	}
	return s.GetMe(ctx, userID)
}

func (s *Service) ChangeEmail(ctx context.Context, userID uuid.UUID, newEmail, currentPassword string) (*identity.User, error) {
	normalized, err := normalizeEmail(newEmail)
	if err != nil {
		return nil, err
	}
	user, passwordHash, err := s.users.GetAuthByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if passwordHash == nil || *passwordHash == "" {
		return nil, apperrors.ErrUnauthorized
	}
	if !s.passwords.Check(*passwordHash, currentPassword) {
		return nil, apperrors.ErrUnauthorized
	}
	if normalized == user.Email {
		user.HasPassword = true
		return user, nil
	}
	existing, err := s.users.GetByEmail(ctx, normalized)
	if err == nil && existing.ID != userID {
		return nil, apperrors.ErrConflict
	}
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}
	if err := s.users.UpdateEmail(ctx, userID, normalized); err != nil {
		return nil, err
	}
	return s.GetMe(ctx, userID)
}

func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	_, passwordHash, err := s.users.GetAuthByID(ctx, userID)
	if err != nil {
		return err
	}
	if passwordHash == nil || *passwordHash == "" {
		return apperrors.ErrUnauthorized
	}
	if !s.passwords.Check(*passwordHash, currentPassword) {
		return apperrors.ErrUnauthorized
	}
	hash, err := s.passwords.Hash(newPassword)
	if err != nil {
		return err
	}
	return s.users.UpdatePassword(ctx, userID, hash)
}

func (s *Service) GetNotificationPreferences(ctx context.Context, userID uuid.UUID) (*identity.NotificationPreferences, error) {
	if s.notifications == nil {
		prefs := identity.DefaultNotificationPreferences(userID)
		return &prefs, nil
	}
	prefs, err := s.notifications.GetByUserID(ctx, userID)
	if err == nil {
		return prefs, nil
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}
	defaults := identity.DefaultNotificationPreferences(userID)
	if err := s.notifications.Upsert(ctx, defaults); err != nil {
		return nil, err
	}
	return &defaults, nil
}

// UpdateNotificationPreferences updates the email/newsletter toggles this endpoint exposes.
// TelegramChatID is intentionally not part of the incoming prefs value here — it's managed
// only by the Telegram bot's link flow — so it's carried over from the existing row rather
// than accepted from the caller, to avoid silently unlinking it on every unrelated save.
func (s *Service) UpdateNotificationPreferences(ctx context.Context, userID uuid.UUID, prefs identity.NotificationPreferences) (*identity.NotificationPreferences, error) {
	if s.notifications == nil {
		return nil, apperrors.ErrNotImplemented
	}
	current, err := s.notifications.GetByUserID(ctx, userID)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}
	if current != nil {
		prefs.TelegramChatID = current.TelegramChatID
	}
	prefs.UserID = userID
	prefs.UpdatedAt = time.Now().UTC()
	if err := s.notifications.Upsert(ctx, prefs); err != nil {
		return nil, err
	}
	return s.notifications.GetByUserID(ctx, userID)
}

func (s *Service) ListUsers(ctx context.Context, role *identity.Role, limit, offset int) ([]identity.User, int, error) {
	if s.userManagement == nil {
		return nil, 0, apperrors.ErrNotImplemented
	}
	return s.userManagement.List(ctx, role, limit, offset)
}

func (s *Service) UpdateUserRole(ctx context.Context, actorID, userID uuid.UUID, role identity.Role) (*identity.User, error) {
	switch role {
	case identity.RolePublic, identity.RoleArtist, identity.RoleInstitution, identity.RoleContributor, identity.RoleAdmin:
	default:
		return nil, apperrors.ErrValidation
	}
	current, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if actorID == userID && current.Role == identity.RoleAdmin && role != identity.RoleAdmin {
		return nil, apperrors.ErrForbidden
	}
	if current.Role == identity.RoleAdmin && role != identity.RoleAdmin {
		if s.userManagement == nil {
			return nil, apperrors.ErrNotImplemented
		}
		count, err := s.userManagement.CountByRole(ctx, identity.RoleAdmin)
		if err != nil {
			return nil, err
		}
		if count <= 1 {
			return nil, apperrors.ErrForbidden
		}
	}
	if err := s.users.UpdateRole(ctx, userID, role); err != nil {
		return nil, err
	}
	return s.users.GetByID(ctx, userID)
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (s *Service) validateReturnTo(returnTo string) error {
	u, err := url.Parse(returnTo)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return apperrors.ErrInvalidToken
	}
	origin := u.Scheme + "://" + u.Host
	if _, ok := s.allowedReturn[origin]; !ok {
		return apperrors.ErrForbidden
	}
	return nil
}

func normalizeEmail(email string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return "", fmt.Errorf("%w: email is required", apperrors.ErrValidation)
	}
	if _, err := mail.ParseAddress(normalized); err != nil {
		return "", fmt.Errorf("%w: invalid email", apperrors.ErrValidation)
	}
	return normalized, nil
}

func validatePassword(password string) error {
	if len(password) < minPasswordLength {
		return fmt.Errorf("%w: password must be at least %d characters", apperrors.ErrValidation, minPasswordLength)
	}
	return nil
}
