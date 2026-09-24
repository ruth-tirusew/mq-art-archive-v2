package outbound

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/digest"
)

// DigestRecipientRepository lists subscribers eligible for the digest on at least one
// channel. It does not filter by channel itself — callers check Recipient.EmailOptedIn
// and Recipient.TelegramChatID to decide which channel(s) to use per recipient.
type DigestRecipientRepository interface {
	ListDigestRecipients(ctx context.Context) ([]digest.Recipient, error)
}

// DigestRunRepository tracks digest send attempts and per-recipient, per-channel delivery,
// so a retried run can skip recipients already sent to instead of sending twice.
type DigestRunRepository interface {
	// LastCompletedRun returns the most recently completed run, or ErrNotFound if none exists.
	LastCompletedRun(ctx context.Context) (*digest.Run, error)
	StartRun(ctx context.Context, periodStart, periodEnd time.Time) (*digest.Run, error)
	CompleteRun(ctx context.Context, runID uuid.UUID) error
	HasSuccessfulDelivery(ctx context.Context, runID, recipientID uuid.UUID, channel digest.Channel) (bool, error)
	// RecordDelivery upserts the delivery attempt for (runID, recipientID, channel).
	// deliveryErr nil means success; sentAt should be the attempt time either way.
	RecordDelivery(ctx context.Context, runID, recipientID uuid.UUID, channel digest.Channel, sentAt time.Time, deliveryErr error) error
}
