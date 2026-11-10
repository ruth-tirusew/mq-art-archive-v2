package inbound

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/identity"
)

// CreateStaffInviteInput is what an admin supplies to invite a moderator.
type CreateStaffInviteInput struct {
	Email string
	// Role must be identity.RoleModerator. Admins are made on the Users page, never by email invite.
	Role identity.Role
	// TTL defaults to 7 days; it must be between 1 hour and 30 days.
	TTL time.Duration
}

// StaffInviteWithToken carries the raw token and the link built from it. They exist only at creation or
// resend: only a hash is stored, so the admin UI should show the link once.
type StaffInviteWithToken struct {
	Invite identity.StaffInvite
	Token  string
	Link   string
}

// StaffInvitePreview is what the accept page shows before the invitee picks a password.
type StaffInvitePreview struct {
	Email     string
	Role      identity.Role
	Status    identity.StaffInviteStatus
	ExpiresAt time.Time
}

// StaffInviteService creates moderators by email. Everything except PreviewInvite and AcceptInvite is for
// admins only; the HTTP router enforces that.
type StaffInviteService interface {
	// CreateInvite emails a one-time link. ErrValidation for a bad address, role or lifetime; ErrConflict
	// if the address already has an account (change that person's role instead).
	CreateInvite(ctx context.Context, adminID uuid.UUID, in CreateStaffInviteInput) (*StaffInviteWithToken, error)
	// ResendInvite replaces an unclaimed invite with a fresh link (same address, role and lifetime).
	ResendInvite(ctx context.Context, adminID, inviteID uuid.UUID) (*StaffInviteWithToken, error)
	RevokeInvite(ctx context.Context, inviteID uuid.UUID) error
	ListInvites(ctx context.Context, limit, offset int) ([]identity.StaffInvite, int, error)
	// PreviewInvite looks an invite up by its raw token. ErrNotFound for an unknown token.
	PreviewInvite(ctx context.Context, rawToken string) (*StaffInvitePreview, error)
	// AcceptInvite creates the invitee's verified account with the invited role and signs it in. The
	// address always comes from the invite, never from the request. ErrNotFound for an unknown token,
	// ErrConflict if the invite was used, revoked or expired, or the address already has an account.
	AcceptInvite(ctx context.Context, rawToken, password string) (accessToken string, user *identity.User, err error)
}
