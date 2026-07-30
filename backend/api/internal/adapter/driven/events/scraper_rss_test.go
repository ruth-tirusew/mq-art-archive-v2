package events

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mq/api/internal/testutil/assist"
)

func TestScraperRSS_parseRSSFixture(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("testdata", "sample.rss"))
	assist.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	s := NewScraperRSS(RSSConfig{
		Sources: []string{srv.URL},
		Client:  srv.Client(),
	})

	since := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	events, err := s.FetchEvents(context.Background(), since)
	assist.NoError(t, err)
	assist.Len(t, 1, len(events))
	assist.Equal(t, "Blue Room Opening", events[0].Title)
	assist.Equal(t, "https://example.com/events/blue-room", events[0].SourceURL)
}

func TestScraperRSS_parseJSONFixture(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("testdata", "sample.json"))
	assist.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/feed+json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	s := NewScraperRSS(RSSConfig{
		Sources: []string{srv.URL + "/feed.json"},
		Client:  srv.Client(),
	})

	events, err := s.FetchEvents(context.Background(), time.Time{})
	assist.NoError(t, err)
	assist.Len(t, 1, len(events))
	assist.Equal(t, "Ceramic Fair", events[0].Title)
}

func TestScraperRSS_toleratesFailedFeed(t *testing.T) {
	s := NewScraperRSS(RSSConfig{
		Sources: []string{"http://127.0.0.1:1/not-listening"},
		Timeout: 100 * time.Millisecond,
	})
	events, err := s.FetchEvents(context.Background(), time.Time{})
	assist.NoError(t, err)
	assist.Len(t, 0, len(events))
}
