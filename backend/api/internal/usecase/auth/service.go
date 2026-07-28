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

func (s *Service) Login(ctx context.Context, email, password string) (string, *identity.User, error) {
	normalized, err := normalizeEmail(email)
	if err != nil {
		return "", nil, apperrors.ErrUnauthorized
	}

	user, passwordHash, err := s.users.GetAuthByEmail(ctx, normalized)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return "", nil, apperrors.ErrUnauthorized
		}
		return "", nil, err
	}
	if passwordHash == nil || *passwordHash == "" {
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

	user, err := s.users.GetByEmail(ctx, info.Email)
	if err == nil {
		if linkErr := s.oauthAccounts.Create(ctx, identity.OAuthAccount{
			UserID:         user.ID,
			Provider:       provider,
			ProviderUserID: info.ProviderUserID,
			Email:          info.Email,
		}); linkErr != nil {
			return nil, linkErr
		}
		return user, nil
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
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

func (s *Service) UpdateNotificationPreferences(ctx context.Context, userID uuid.UUID, prefs identity.NotificationPreferences) (*identity.NotificationPreferences, error) {
	if s.notifications == nil {
		return nil, apperrors.ErrNotImplemented
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
