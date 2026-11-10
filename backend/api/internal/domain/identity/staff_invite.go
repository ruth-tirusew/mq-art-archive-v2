package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// StaffInviteStatus is derived from the invite's timestamps, never stored.
type StaffInviteStatus string

const (
	StaffInviteLive    StaffInviteStatus = "live"
	StaffInviteClaimed StaffInviteStatus = "claimed"
	StaffInviteRevoked StaffInviteStatus = "revoked"
	StaffInviteExpired StaffInviteStatus = "expired"
)

// StaffInvite lets an admin bring a new moderator in by email: the person follows a one-time link, picks
// a password and gets a verified account with Role. Only TokenHash is persisted; the raw token exists
// only in the emailed link.
type StaffInvite struct {
	ID        uuid.UUID
	TokenHash string
	Email     string
	Role      Role
	CreatedBy *uuid.UUID
	ExpiresAt time.Time
	ClaimedBy *uuid.UUID
	ClaimedAt *time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

// Status reports the invite's state at now. Claimed and revoked take precedence over expiry, so a used
// invite never reads as merely "expired".
func (i StaffInvite) Status(now time.Time) StaffInviteStatus {
	switch {
	case i.ClaimedAt != nil:
		return StaffInviteClaimed
	case i.RevokedAt != nil:
		return StaffInviteRevoked
	case !now.Before(i.ExpiresAt):
		return StaffInviteExpired
	default:
		return StaffInviteLive
	}
}

// HashStaffInviteToken returns the hex SHA-256 of a raw token, the form stored and looked up. The token
// is 256 random bits, so a fast hash is enough.
func HashStaffInviteToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
