package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/identity"
)

type NotificationPreferencesRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*identity.NotificationPreferences, error)
	Upsert(ctx context.Context, prefs identity.NotificationPreferences) error
	// SetTelegramChatID links or unlinks (chatID nil) a user's Telegram chat. It's a
	// dedicated single-field update, not a full-row Upsert, so the Telegram bot's link/stop
	// flow can't accidentally clobber a user's other notification preferences — the same
	// class of bug fixed in UpdateNotificationPreferences (see internal/usecase/auth).
	// Creates the preferences row with defaults first if the user doesn't have one yet.
	SetTelegramChatID(ctx context.Context, userID uuid.UUID, chatID *string) error
	// GetByTelegramChatID finds the user linked to a chat, for handling inbound bot
	// messages (e.g. /stop) where only the chat ID is known. Returns ErrNotFound if no
	// user has linked that chat.
	GetByTelegramChatID(ctx context.Context, chatID string) (*identity.NotificationPreferences, error)
}
