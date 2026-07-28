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
	RoleAdmin       Role = "admin"
)

type User struct {
	ID              uuid.UUID
	Email           string
	Role            Role
	DisplayName     string
	AvatarURL       string
	HasPassword     bool // set when loaded with auth metadata; not persisted as a column
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type NotificationPreferences struct {
	UserID                  uuid.UUID
	EmailOnNewApplication   bool
	EmailOnEventSyncSummary bool
	NewsletterEnabled       bool
	UpdatedAt               time.Time
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
