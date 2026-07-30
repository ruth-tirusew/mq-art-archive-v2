package events

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/testutil/assist"
)

type stubSource struct {
	events []domain.Event
	err    error
}

func (s *stubSource) FetchEvents(ctx context.Context, since time.Time) ([]domain.Event, error) {
	return s.events, s.err
}

func TestComposite_mergesAndDedupes(t *testing.T) {
	a := &stubSource{events: []domain.Event{
		{Title: "A", SourceURL: "https://example.com/a"},
		{Title: "Dup", SourceURL: "https://example.com/dup"},
	}}
	b := &stubSource{events: []domain.Event{
		{Title: "B", SourceURL: "https://example.com/b"},
		{Title: "Dup2", SourceURL: "https://example.com/dup"},
	}}
	c := NewComposite(a, b)
	events, err := c.FetchEvents(context.Background(), time.Time{})
	assist.NoError(t, err)
	assist.Len(t, 3, len(events))
}

func TestComposite_partialFailure(t *testing.T) {
	a := &stubSource{err: errors.New("boom")}
	b := &stubSource{events: []domain.Event{{Title: "Ok", SourceURL: "https://example.com/ok"}}}
	c := NewComposite(a, b)
	events, err := c.FetchEvents(context.Background(), time.Time{})
	assist.NoError(t, err)
	assist.Len(t, 1, len(events))
}
