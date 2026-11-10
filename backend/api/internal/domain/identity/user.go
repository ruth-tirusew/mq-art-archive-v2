package identity

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RolePublic      Role = "public"
	RoleArtist      Role = "artist"
	RoleInstitution Role = "institution"
	RoleContributor Role = "contributor"
	// RoleModerator reviews and moderates content but cannot manage users or change system settings.
	RoleModerator Role = "moderator"
	RoleAdmin     Role = "admin"
)

type User struct {
	ID              uuid.UUID
	Email           string
	Role            Role
	DisplayName     string
	AvatarURL       string
	HasPassword     bool // set when loaded with auth metadata; not persisted as a column
	GoogleLinked    bool // set by GetMe; not persisted as a column
	EmailVerifiedAt *time.Time
	// SessionsValidAfter rejects any access token issued before it (see password change/reset).
	SessionsValidAfter *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type NotificationPreferences struct {
	UserID                  uuid.UUID
	EmailOnNewApplication   bool
	EmailOnEventSyncSummary bool
	NewsletterEnabled       bool
	// TelegramChatID is set once the user links their Telegram account via the bot's
	// /start flow. Nil until linked.
	TelegramChatID *string
	UpdatedAt      time.Time
}

func DefaultNotificationPreferences(userID uuid.UUID) NotificationPreferences {
	return NotificationPreferences{
		UserID:                  userID,
		EmailOnNewApplication:   true,
		EmailOnEventSyncSummary: false,
		NewsletterEnabled:       false,
		UpdatedAt:               time.Now().UTC(),
	}
}

// IsStaff reports whether the role may use the admin app. Staff can do different things inside it: see the
// admin routes in the HTTP router for which are admin-only.
func (r Role) IsStaff() bool { return r == RoleAdmin || r == RoleModerator }
