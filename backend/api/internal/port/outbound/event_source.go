package outbound

import (
	"context"
	"time"

	"github.com/mq/api/internal/domain/events"
)

type EventSource interface {
	FetchEvents(ctx context.Context, since time.Time) ([]events.Event, error)
}
