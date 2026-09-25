package outbound

import (
	"context"

	"github.com/mq/api/internal/domain/digest"
)

// TelegramLinkRepository stores the one-time tokens used to link a Telegram chat to a
// site account (see digest.LinkToken).
type TelegramLinkRepository interface {
	Create(ctx context.Context, token digest.LinkToken) error
	// Consume atomically marks the token used and returns it, or ErrNotFound if the token
	// doesn't exist, is expired, or was already consumed — a token can only ever link one
	// chat, so a replayed /start with the same token must not re-trigger the link.
	Consume(ctx context.Context, token string) (*digest.LinkToken, error)
}
