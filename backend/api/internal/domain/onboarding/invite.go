package onboarding

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
)

// InviteStatus is derived from the invite's timestamps, never stored.
type InviteStatus string

const (
	InviteLive    InviteStatus = "live"
	InviteClaimed InviteStatus = "claimed"
	InviteRevoked InviteStatus = "revoked"
	InviteExpired InviteStatus = "expired"
)

// Invite brings a pre-vetted artist into the normal application flow with a prefilled name and
// reserved handle. The application it produces is approved automatically on submit. Only TokenHash is persisted; the raw token exists
// only in the link that is emailed.
type Invite struct {
	ID              uuid.UUID
	TokenHash       string
	Email           string
	DisplayName     string
	RequestedHandle string
	CreatedBy       *uuid.UUID
	ExpiresAt       time.Time
	ClaimedBy       *uuid.UUID
	ClaimedAt       *time.Time
	RevokedAt       *time.Time
	CreatedAt       time.Time
}

// Status reports the invite's state at now. Claimed and revoked take precedence over expiry so a
// used invite never reads as merely "expired".
func (i Invite) Status(now time.Time) InviteStatus {
	switch {
	case i.ClaimedAt != nil:
		return InviteClaimed
	case i.RevokedAt != nil:
		return InviteRevoked
	case !now.Before(i.ExpiresAt):
		return InviteExpired
	default:
		return InviteLive
	}
}

// Claimable reports whether the invite can still be used by an account with the given email.
// The email check is case-insensitive.
func (i Invite) Claimable(now time.Time, email string) bool {
	return i.Status(now) == InviteLive && strings.EqualFold(strings.TrimSpace(email), strings.TrimSpace(i.Email))
}

// HashInviteToken returns the hex SHA-256 of a raw invite token, the form stored and looked up.
// The token is 256 random bits, so a fast hash is sufficient (no need for a password KDF).
func HashInviteToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
