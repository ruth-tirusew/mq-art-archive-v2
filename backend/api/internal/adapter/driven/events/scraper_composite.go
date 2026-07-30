package events

import (
	"context"
	"log"
	"time"

	domain "github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/port/outbound"
)

type Composite struct {
	sources []outbound.EventSource
}

func NewComposite(sources ...outbound.EventSource) outbound.EventSource {
	return &Composite{sources: sources}
}

func (c *Composite) FetchEvents(ctx context.Context, since time.Time) ([]domain.Event, error) {
	seen := make(map[string]struct{})
	var merged []domain.Event

	for _, src := range c.sources {
		if src == nil {
			continue
		}
		events, err := src.FetchEvents(ctx, since)
		if err != nil {
			log.Printf("event source error: %v", err)
			continue
		}
		for _, ev := range events {
			if ev.SourceURL == "" {
				continue
			}
			if _, ok := seen[ev.SourceURL]; ok {
				continue
			}
			seen[ev.SourceURL] = struct{}{}
			merged = append(merged, ev)
		}
	}
	if merged == nil {
		merged = []domain.Event{}
	}
	return merged, nil
}
