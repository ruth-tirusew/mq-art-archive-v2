package onboarding

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	domain "github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/port/inbound"
)

const (
	defaultInviteTTL = 14 * 24 * time.Hour
	maxInviteTTL     = 60 * 24 * time.Hour
)

var _ inbound.InviteService = (*Service)(nil)

func (s *Service) requireInvites() error {
	if s.invites == nil {
		return apperrors.ErrNotImplemented
	}
	return nil
}

func newInviteToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate invite token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *Service) CreateInvite(ctx context.Context, adminID uuid.UUID, in inbound.CreateInviteInput) (*inbound.InviteWithToken, error) {
	if err := s.requireInvites(); err != nil {
		return nil, err
	}
	email := strings.TrimSpace(in.Email)
	if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email || len(email) > 254 {
		return nil, fmt.Errorf("%w: %s", apperrors.ErrValidation, "a valid email address is required")
	}
	// Accounts are stored lower-case, and the user lookup is an exact match.
	email = strings.ToLower(email)
	// Fail here, not when the invitee opens the link: an invite to someone who already has an
	// account can never be used (invited registration refuses existing addresses), and the admin
	// should hear that before an email goes out.
	if _, err := s.users.GetByEmail(ctx, email); err == nil {
		return nil, fmt.Errorf("%w: %s already has an Artiv account", apperrors.ErrConflict, email)
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}
	handle := strings.ToLower(strings.TrimSpace(in.Handle))
	if handle != "" {
		if !validHandle(handle) {
			return nil, fmt.Errorf("%w: %s", apperrors.ErrValidation, "handle must be 3-30 lowercase letters, numbers, or underscores")
		}
		available, err := s.handleAvailable(ctx, handle, nil)
		if err != nil {
			return nil, err
		}
		if !available {
			return nil, fmt.Errorf("%w: handle %q is already taken or reserved", apperrors.ErrConflict, handle)
		}
	}
	ttl := in.TTL
	if ttl == 0 {
		ttl = defaultInviteTTL
	}
	if ttl < time.Hour || ttl > maxInviteTTL {
		return nil, fmt.Errorf("%w: %s", apperrors.ErrValidation, "invite lifetime must be between 1 hour and 60 days")
	}

	token, err := newInviteToken()
	if err != nil {
		return nil, err
	}
	now := s.now()
	invite, err := s.invites.Create(ctx, domain.Invite{
		ID: uuid.New(), TokenHash: domain.HashInviteToken(token), Email: email,
		DisplayName: strings.TrimSpace(in.DisplayName), RequestedHandle: handle,
		CreatedBy: &adminID, ExpiresAt: now.Add(ttl), CreatedAt: now,
	})
	if err != nil {
		return nil, err
	}
	s.sendInviteEmail(ctx, invite, token)
	return &inbound.InviteWithToken{Invite: *invite, Token: token}, nil
}

func (s *Service) ResendInvite(ctx context.Context, adminID, inviteID uuid.UUID) (*inbound.InviteWithToken, error) {
	if err := s.requireInvites(); err != nil {
		return nil, err
	}
	old, err := s.invites.GetByID(ctx, inviteID)
	if err != nil {
		return nil, err
	}
	if old.ClaimedAt != nil {
		return nil, fmt.Errorf("%w: this invite was already used", apperrors.ErrConflict)
	}
	// Keep the lifetime the admin originally chose.
	ttl := min(max(old.ExpiresAt.Sub(old.CreatedAt), time.Hour), maxInviteTTL)
	// Revoke first so the old link dies and its handle reservation is released for the new one.
	if err := s.invites.Revoke(ctx, inviteID, s.now()); err != nil {
		return nil, err
	}
	created, err := s.CreateInvite(ctx, adminID, inbound.CreateInviteInput{
		Email: old.Email, DisplayName: old.DisplayName, Handle: old.RequestedHandle, TTL: ttl,
	})
	if err != nil {
		// Don't leave the invitee with a dead link and no replacement. An invite that was already
		// revoked before the resend stays revoked.
		if old.RevokedAt == nil {
			if reopenErr := s.invites.Reopen(context.WithoutCancel(ctx), inviteID); reopenErr != nil {
				return nil, errors.Join(err, fmt.Errorf("restore invite after failed resend: %w", reopenErr))
			}
		}
		return nil, err
	}
	return created, nil
}

func (s *Service) RevokeInvite(ctx context.Context, inviteID uuid.UUID) error {
	if err := s.requireInvites(); err != nil {
		return err
	}
	return s.invites.Revoke(ctx, inviteID, s.now())
}

func (s *Service) ListInvites(ctx context.Context, filter domain.InviteListFilter) ([]domain.Invite, int, error) {
	if err := s.requireInvites(); err != nil {
		return nil, 0, err
	}
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.invites.List(ctx, filter, s.now())
}

func (s *Service) PreviewInvite(ctx context.Context, rawToken string) (*inbound.InvitePreview, error) {
	if err := s.requireInvites(); err != nil {
		return nil, err
	}
	inv, err := s.invites.GetByTokenHash(ctx, domain.HashInviteToken(strings.TrimSpace(rawToken)))
	if err != nil {
		return nil, err
	}
	return &inbound.InvitePreview{
		Email: inv.Email, DisplayName: inv.DisplayName, Handle: inv.RequestedHandle, Status: inv.Status(s.now()),
	}, nil
}

func (s *Service) liveInviteEmail(ctx context.Context, rawToken string) (string, error) {
	if err := s.requireInvites(); err != nil {
		return "", err
	}
	inv, err := s.invites.GetByTokenHash(ctx, domain.HashInviteToken(strings.TrimSpace(rawToken)))
	if err != nil {
		return "", err
	}
	if status := inv.Status(s.now()); status != domain.InviteLive {
		return "", fmt.Errorf("%w: this invite is %s", apperrors.ErrConflict, status)
	}
	return inv.Email, nil
}

func (s *Service) RegisterWithInvite(ctx context.Context, rawToken, password string) (*inbound.RegisteredInvitee, error) {
	if err := s.requireInvites(); err != nil {
		return nil, err
	}
	if s.registrar == nil {
		return nil, apperrors.ErrNotImplemented
	}
	email, err := s.liveInviteEmail(ctx, rawToken)
	if err != nil {
		return nil, err
	}
	token, user, err := s.registrar.RegisterInvited(ctx, email, password)
	if err != nil {
		return nil, err
	}
	return &inbound.RegisteredInvitee{AccessToken: token, User: user}, nil
}

func (s *Service) SubmitWithInvite(
	ctx context.Context,
	applicantID uuid.UUID,
	rawToken, displayName, requestedHandle, notes string,
) (*domain.OnboardingApplication, error) {
	if err := s.requireInvites(); err != nil {
		return nil, err
	}
	inv, err := s.invites.GetByTokenHash(ctx, domain.HashInviteToken(strings.TrimSpace(rawToken)))
	if err != nil {
		return nil, err
	}
	user, err := s.users.GetByID(ctx, applicantID)
	if err != nil {
		return nil, err
	}
	// The invite is bound to an email address. Without a verified address anyone could register
	// someone else's email and claim their invite.
	if user.EmailVerifiedAt == nil {
		return nil, fmt.Errorf("%w: verify your email address before using an invite", apperrors.ErrForbidden)
	}
	now := s.now()
	switch status := inv.Status(now); status {
	case domain.InviteLive:
	default:
		return nil, fmt.Errorf("%w: this invite is %s", apperrors.ErrConflict, status)
	}
	if !inv.Claimable(now, user.Email) {
		return nil, fmt.Errorf("%w: this invite was sent to a different email address", apperrors.ErrForbidden)
	}

	if strings.TrimSpace(displayName) == "" {
		displayName = inv.DisplayName
	}
	if strings.TrimSpace(requestedHandle) == "" {
		requestedHandle = inv.RequestedHandle
	}
	app, err := s.prepareApplication(ctx, applicantID, domain.ApplicantTypeArtist, displayName, requestedHandle, notes, &inv.ID)
	if err != nil {
		return nil, err
	}
	app.InviteID = &inv.ID

	// Everything that can reasonably fail has been checked, so claim (the single-use step) last
	// before saving. If Save then fails, the invite is reopened below.
	if _, err := s.invites.Claim(ctx, inv.ID, applicantID, now); err != nil {
		return nil, err
	}
	saved, err := s.applications.Save(ctx, *app)
	if err != nil {
		// The invite was spent but no application exists: give the link back.
		if reopenErr := s.invites.Reopen(context.WithoutCancel(ctx), inv.ID); reopenErr != nil {
			return nil, errors.Join(err, fmt.Errorf("restore invite after failed submit: %w", reopenErr))
		}
		return nil, err
	}
	// The invite was issued by an admin to a pre-vetted artist, so there is nothing left to
	// review: approve now, crediting the admin who invited them.
	approved, err := s.review(ctx, saved.ID, inv.CreatedBy, domain.ApprovalStatusApproved, "")
	if err != nil {
		// Provisioning failed and review reopened the application, so it stays pending for an
		// admin to approve by hand. The invite is spent and the submission stands.
		log.Printf("invite %s: auto-approval of application %s failed, left pending: %v", inv.ID, saved.ID, err)
		return saved, nil
	}
	return approved, nil
}

// sendInviteEmail is best-effort, like decision emails: the admin also gets the link back from
// CreateInvite, which matters while no mail key is configured.
func (s *Service) sendInviteEmail(ctx context.Context, inv *domain.Invite, token string) {
	if s.mailer == nil {
		return
	}
	var b strings.Builder
	name := inv.DisplayName
	if name == "" {
		name = "there"
	}
	fmt.Fprintf(&b, "Hi %s,\n\nYou've been invited to join Artiv as an artist.\n\n", name)
	fmt.Fprintf(&b, "Open this link to create your account and tell us about your work:\n%s/apply?invite=%s\n\n", s.webAppURL, token)
	fmt.Fprintf(&b, "You'll set a password for %s, answer a few questions, and your profile will be approved automatically.\n\n", inv.Email)
	fmt.Fprintf(&b, "The link works once and expires on %s.\n", inv.ExpiresAt.UTC().Format("2 January 2006"))
	// The invite exists either way and an admin can resend it, so a failed send is logged, not
	// returned. The error never contains the message body, so the link stays out of the log.
	if err := s.mailer.Send(ctx, inv.Email, "You're invited to Artiv", b.String()); err != nil {
		log.Printf("invite %s: email to %s not sent: %v", inv.ID, inv.Email, err)
	}
}
