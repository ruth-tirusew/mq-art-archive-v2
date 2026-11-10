package outbound

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/identity"
)

// StaffInviteRepository stores moderator invites. Lookups never return the raw token; callers hash it with
// identity.HashStaffInviteToken first.
type StaffInviteRepository interface {
	// Create stores a new invite. Any other unclaimed invite for the same address is revoked first, so a
	// resend replaces the old link instead of leaving two working ones.
	Create(ctx context.Context, invite identity.StaffInvite) (*identity.StaffInvite, error)
	GetByID(ctx context.Context, id uuid.UUID) (*identity.StaffInvite, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*identity.StaffInvite, error)
	// List returns one page of invites, newest first, and the total count.
	List(ctx context.Context, limit, offset int) ([]identity.StaffInvite, int, error)
	// Accept creates the invited account and marks the invite claimed in ONE transaction, so an invite
	// that was revoked, expired or already used can never produce an account. ErrConflict if the invite
	// is not live or the address already has an account.
	Accept(ctx context.Context, inviteID uuid.UUID, user identity.User, passwordHash string, now time.Time) (*identity.User, error)
	// Revoke invalidates an unclaimed invite. ErrConflict if it was already claimed.
	Revoke(ctx context.Context, id uuid.UUID, now time.Time) error
}
