package inbound

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/domain/content"
	digestdomain "github.com/mq/api/internal/domain/digest"
	"github.com/mq/api/internal/domain/events"
)

// Digest is the content window for one send: events starting soon, plus art posts and
// articles published since the previous run (or the last 7 days, if there's no prior run).
// It lives here rather than in internal/domain/digest because, like SearchResults, it's an
// intentional read-model aggregate across bounded contexts.
type Digest struct {
	PeriodStart time.Time
	PeriodEnd   time.Time
	Events      []events.Event
	ArtPosts    []art.ArtPostWithArtist
	Articles    []content.Article
}

// DigestService is what the HTTP layer and the cmd/digest scheduler depend on — content
// building, recipients and run bookkeeping (used by the scheduler), plus the Telegram
// linking and email unsubscribe flows (used by HTTP handlers).
type DigestService interface {
	BuildWeekly(ctx context.Context) (*Digest, error)
	// SendWeekly builds, sends, and records one full digest run. This is what the
	// cmd/digest scheduler calls; BuildWeekly/Recipients/StartRun/etc. below are its
	// building blocks, exposed separately for testing and for callers that need finer
	// control.
	SendWeekly(ctx context.Context) error
	Recipients(ctx context.Context) ([]digestdomain.Recipient, error)
	StartRun(ctx context.Context, d Digest) (*digestdomain.Run, error)
	CompleteRun(ctx context.Context, runID uuid.UUID) error
	HasSuccessfulDelivery(ctx context.Context, runID, recipientID uuid.UUID, channel digestdomain.Channel) (bool, error)
	RecordDelivery(ctx context.Context, runID, recipientID uuid.UUID, channel digestdomain.Channel, sentAt time.Time, deliveryErr error) error

	CreateTelegramLink(ctx context.Context, userID uuid.UUID) (string, error)
	HandleTelegramStart(ctx context.Context, chatID, text string) error
	HandleTelegramStop(ctx context.Context, chatID string) error

	SignUnsubscribeToken(userID uuid.UUID) string
	Unsubscribe(ctx context.Context, token string) error
}
