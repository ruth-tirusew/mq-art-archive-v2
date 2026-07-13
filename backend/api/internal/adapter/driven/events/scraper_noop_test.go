package events

import (
	"context"
	"testing"
	"time"

	"github.com/mq/api/internal/testutil/assist"
)

func TestScraperNoop_FetchEvents_returnsEmpty(t *testing.T) {
	s := NewScraperNoop()
	events, err := s.FetchEvents(context.Background(), time.Now().UTC())
	assist.NoError(t, err)
	assist.Len(t, 0, len(events))
}
