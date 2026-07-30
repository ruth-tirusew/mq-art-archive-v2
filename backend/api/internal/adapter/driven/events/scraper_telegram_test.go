package events

import (
	"context"
	"testing"
	"time"

	"github.com/mq/api/internal/testutil/assist"
)

type mockFetcher struct {
	msgs map[string][]TelegramMessage
	err  error
}

func (m *mockFetcher) FetchChannelMessages(ctx context.Context, channel string, since time.Time, limit int) ([]TelegramMessage, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.msgs[channel], nil
}

func TestScraperTelegram_mapsMessages(t *testing.T) {
	fetcher := &mockFetcher{
		msgs: map[string][]TelegramMessage{
			"addisart": {{
				Channel:   "addisart",
				MessageID: 7,
				Date:      time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
				Text:      "Exhibition opening tonight in Addis Ababa",
			}},
		},
	}

	s := NewScraperTelegram(TelegramScraperConfig{
		Channels: []string{"addisart"},
		Keywords: []string{"exhibition"},
		Fetcher:  fetcher,
	})

	events, err := s.FetchEvents(context.Background(), time.Time{})
	assist.NoError(t, err)
	assist.Len(t, 1, len(events))
	assist.Equal(t, "https://t.me/addisart/7", events[0].SourceURL)
}

func TestScraperTelegram_toleratesChannelError(t *testing.T) {
	s := NewScraperTelegram(TelegramScraperConfig{
		Channels: []string{"broken"},
		Fetcher:  &mockFetcher{err: context.DeadlineExceeded},
	})
	events, err := s.FetchEvents(context.Background(), time.Time{})
	assist.NoError(t, err)
	assist.Len(t, 0, len(events))
}
