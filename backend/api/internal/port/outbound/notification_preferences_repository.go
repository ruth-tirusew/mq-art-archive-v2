package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/identity"
)

type NotificationPreferencesRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*identity.NotificationPreferences, error)
	Upsert(ctx context.Context, prefs identity.NotificationPreferences) error
}
