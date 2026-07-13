package identity

import (
	"time"

	"github.com/google/uuid"
)

type OAuthAccount struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Provider       string
	ProviderUserID string
	Email          string
	CreatedAt      time.Time
}
