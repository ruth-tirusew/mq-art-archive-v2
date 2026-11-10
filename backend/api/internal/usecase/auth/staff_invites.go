package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

const (
	defaultStaffInviteTTL = 7 * 24 * time.Hour
	maxStaffInviteTTL     = 30 * 24 * time.Hour
)

// StaffInviteService lets admins bring new moderators in by email. It lives in the auth package because
// it creates accounts and signs people in, and shares the email and password rules with registration.
type StaffInviteService struct {
	users     outbound.UserRepository
	passwords outbound.PasswordHasher
	tokens    outbound.TokenIssuer
	invites   outbound.StaffInviteRepository
	mailer    outbound.Mailer
	adminURL  string
	now       func() time.Time
}

func NewStaffInviteService(
	users outbound.UserRepository,
	passwords outbound.PasswordHasher,
	tokens outbound.TokenIssuer,
	invites outbound.StaffInviteRepository,
	mailer outbound.Mailer,
	adminAppURL string,
) inbound.StaffInviteService {
	return &StaffInviteService{
		users: users, passwords: passwords, tokens: tokens, invites: invites, mailer: mailer,
		adminURL: strings.TrimRight(adminAppURL, "/"), now: func() time.Time { return time.Now().UTC() },
	}
}

func (s *StaffInviteService) CreateInvite(ctx context.Context, adminID uuid.UUID, in inbound.CreateStaffInviteInput) (*inbound.StaffInviteWithToken, error) {
	if in.Role != identity.RoleModerator {
		return nil, fmt.Errorf("%w: only moderators can be invited by email; to make someone an admin, change their role on the Users page", apperrors.ErrValidation)
	}
	email, err := normalizeEmail(in.Email)
	if err != nil {
		return nil, err
	}
	ttl := in.TTL
	if ttl == 0 {
		ttl = defaultStaffInviteTTL
	}
	if ttl < time.Hour || ttl > maxStaffInviteTTL {
		return nil, fmt.Errorf("%w: invite lifetime must be between 1 hour and 30 days", apperrors.ErrValidation)
	}
	// Refuse up front, like artist invites: an invite to someone who already has an account could never be
	// accepted (accepting creates the account), and the admin should hear that before an email goes out.
	if _, err := s.users.GetByEmail(ctx, email); err == nil {
		return nil, fmt.Errorf("%w: %s already has an Artiv account. Change their role on the Users page instead", apperrors.ErrConflict, email)
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}

	token, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	now := s.now()
	invite, err := s.invites.Create(ctx, identity.StaffInvite{
		ID: uuid.New(), TokenHash: identity.HashStaffInviteToken(token), Email: email, Role: in.Role,
		CreatedBy: &adminID, ExpiresAt: now.Add(ttl), CreatedAt: now,
	})
	if err != nil {
		return nil, err
	}
	link := s.acceptLink(token)
	s.sendInviteEmail(ctx, invite, link)
	return &inbound.StaffInviteWithToken{Invite: *invite, Token: token, Link: link}, nil
}

func (s *StaffInviteService) acceptLink(token string) string {
	return s.adminURL + "/accept-invite?token=" + url.QueryEscape(token)
}

// sendInviteEmail is best-effort: the invite exists either way, the admin is shown the link too, and a
// resend is one click. Failures are logged (never with the link, which is a way in).
func (s *StaffInviteService) sendInviteEmail(ctx context.Context, inv *identity.StaffInvite, link string) {
	if s.mailer == nil {
		return
	}
	var b strings.Builder
	b.WriteString("Hello,\n\nYou've been invited to help moderate Artiv.\n\n")
	fmt.Fprintf(&b, "Open this link to choose a password and get started:\n%s\n\n", link)
	fmt.Fprintf(&b, "The link works once and expires on %s. If you weren't expecting this, you can ignore it.\n", inv.ExpiresAt.UTC().Format("2 January 2006"))
	if err := s.mailer.Send(ctx, inv.Email, "You're invited to moderate Artiv", b.String()); err != nil {
		log.Printf("staff invite %s: email to %s not sent: %v", inv.ID, inv.Email, err)
	}
}

func (s *StaffInviteService) ResendInvite(ctx context.Context, adminID, inviteID uuid.UUID) (*inbound.StaffInviteWithToken, error) {
	old, err := s.invites.GetByID(ctx, inviteID)
	if err != nil {
		return nil, err
	}
	if old.ClaimedAt != nil {
		return nil, fmt.Errorf("%w: this invite was already used", apperrors.ErrConflict)
	}
	// Keep the lifetime the admin originally chose. Creating the new invite revokes the old link.
	ttl := min(max(old.ExpiresAt.Sub(old.CreatedAt), time.Hour), maxStaffInviteTTL)
	return s.CreateInvite(ctx, adminID, inbound.CreateStaffInviteInput{Email: old.Email, Role: old.Role, TTL: ttl})
}

func (s *StaffInviteService) RevokeInvite(ctx context.Context, inviteID uuid.UUID) error {
	return s.invites.Revoke(ctx, inviteID, s.now())
}

func (s *StaffInviteService) ListInvites(ctx context.Context, limit, offset int) ([]identity.StaffInvite, int, error) {
	return s.invites.List(ctx, limit, offset)
}

func (s *StaffInviteService) PreviewInvite(ctx context.Context, rawToken string) (*inbound.StaffInvitePreview, error) {
	inv, err := s.invites.GetByTokenHash(ctx, identity.HashStaffInviteToken(rawToken))
	if err != nil {
		return nil, err
	}
	return &inbound.StaffInvitePreview{Email: inv.Email, Role: inv.Role, Status: inv.Status(s.now()), ExpiresAt: inv.ExpiresAt}, nil
}

func (s *StaffInviteService) AcceptInvite(ctx context.Context, rawToken, password string) (string, *identity.User, error) {
	// Check the password before touching the invite, so a typo never burns the one-time link.
	if err := validatePassword(password); err != nil {
		return "", nil, err
	}
	inv, err := s.invites.GetByTokenHash(ctx, identity.HashStaffInviteToken(rawToken))
	if err != nil {
		return "", nil, err
	}
	now := s.now()
	if inv.Status(now) != identity.StaffInviteLive {
		return "", nil, apperrors.ErrConflict
	}
	hash, err := s.passwords.Hash(password)
	if err != nil {
		return "", nil, err
	}
	// The address and the role come from the invite on the server, never from the request.
	user, err := s.invites.Accept(ctx, inv.ID, identity.User{Email: inv.Email, Role: inv.Role, EmailVerifiedAt: &now}, hash, now)
	if err != nil {
		return "", nil, err
	}
	token, err := s.tokens.Issue(ctx, user)
	if err != nil {
		return "", nil, err
	}
	user.HasPassword = true
	return token, user, nil
}
