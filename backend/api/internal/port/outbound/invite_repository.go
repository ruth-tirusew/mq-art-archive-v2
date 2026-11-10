package outbound

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/onboarding"
)

// InviteRepository stores artist invites. Lookups never return the raw token; callers hash it
// with onboarding.HashInviteToken first.
type InviteRepository interface {
	// Create stores a new invite. Any expired, still-unclaimed invite holding the same handle is
	// revoked first, so expiry frees the handle. ErrConflict if a live invite holds it.
	Create(ctx context.Context, invite onboarding.Invite) (*onboarding.Invite, error)
	GetByID(ctx context.Context, id uuid.UUID) (*onboarding.Invite, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*onboarding.Invite, error)
	// GetLiveByHandle returns the live (unclaimed, unrevoked, unexpired) invite reserving the
	// handle, or ErrNotFound. Handle availability checks use it.
	GetLiveByHandle(ctx context.Context, handle string, now time.Time) (*onboarding.Invite, error)
	// List returns one page of invites, newest first, and the total number matching the filter.
	// Status is derived from the timestamps, so it is evaluated against now.
	List(ctx context.Context, filter onboarding.InviteListFilter, now time.Time) ([]onboarding.Invite, int, error)
	// Claim atomically marks a live invite claimed by the user. ErrConflict if it is already
	// claimed, revoked or expired, so two concurrent claims cannot both succeed.
	Claim(ctx context.Context, id, userID uuid.UUID, now time.Time) (*onboarding.Invite, error)
	// Revoke invalidates an unclaimed invite. ErrConflict if it was already claimed.
	Revoke(ctx context.Context, id uuid.UUID, now time.Time) error
	// Reopen clears the revoked and claimed marks. It only compensates for a step that failed after
	// the invite was revoked or claimed, so the invitee is not left without a working link.
	Reopen(ctx context.Context, id uuid.UUID) error
}
