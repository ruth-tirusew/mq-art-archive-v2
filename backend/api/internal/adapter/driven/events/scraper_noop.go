package events

import (
	"context"
	"time"

	"github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/port/outbound"
)

type ScraperNoop struct{}

func NewScraperNoop() outbound.EventSource {
	return &ScraperNoop{}
}

func (s *ScraperNoop) FetchEvents(ctx context.Context, since time.Time) ([]events.Event, error) {
	return []events.Event{}, nil
}
