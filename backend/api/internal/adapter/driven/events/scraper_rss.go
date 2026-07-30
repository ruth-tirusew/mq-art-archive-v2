package events

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"
	domain "github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/port/outbound"
)

type RSSConfig struct {
	Sources   []string
	UserAgent string
	Timeout   time.Duration
	Client    *http.Client // optional; for tests
}

type ScraperRSS struct {
	sources   []string
	userAgent string
	client    *http.Client
	parser    *gofeed.Parser
}

func NewScraperRSS(cfg RSSConfig) outbound.EventSource {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ua := cfg.UserAgent
	if ua == "" {
		ua = "mq-scraper/1.0"
	}
	client := cfg.Client
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}
	return &ScraperRSS{
		sources:   cfg.Sources,
		userAgent: ua,
		client:    client,
		parser:    gofeed.NewParser(),
	}
}

func (s *ScraperRSS) FetchEvents(ctx context.Context, since time.Time) ([]domain.Event, error) {
	var all []domain.Event
	now := time.Now().UTC()

	for _, src := range s.sources {
		src = strings.TrimSpace(src)
		if src == "" {
			continue
		}
		items, err := s.fetchSource(ctx, src, since, now)
		if err != nil {
			// tolerate per-feed failures
			continue
		}
		all = append(all, items...)
	}
	if all == nil {
		all = []domain.Event{}
	}
	return all, nil
}

func (s *ScraperRSS) fetchSource(ctx context.Context, url string, since, now time.Time) ([]domain.Event, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/json, application/xml, text/xml, */*")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}

	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	if strings.Contains(ct, "json") || strings.HasSuffix(strings.ToLower(url), ".json") {
		return parseJSONFeed(body, since, now)
	}
	return s.parseRSS(body, since, now)
}

func (s *ScraperRSS) parseRSS(body []byte, since, now time.Time) ([]domain.Event, error) {
	feed, err := s.parser.ParseString(string(body))
	if err != nil {
		return nil, err
	}

	var out []domain.Event
	for _, item := range feed.Items {
		if item == nil || strings.TrimSpace(item.Link) == "" {
			continue
		}
		published := item.PublishedParsed
		if published == nil {
			published = item.UpdatedParsed
		}
		if published != nil && !since.IsZero() && !published.After(since) {
			continue
		}
		startsAt := now
		if published != nil {
			startsAt = published.UTC()
		}
		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = item.Link
		}
		desc := strings.TrimSpace(item.Description)
		if desc == "" {
			desc = strings.TrimSpace(item.Content)
		}
		if len(desc) > 2000 {
			desc = desc[:2000]
		}
		ev := domain.Event{
			ID:          uuid.New(),
			Title:       title,
			Description: desc,
			SourceURL:   item.Link,
			StartsAt:    startsAt,
			ScrapedAt:   now,
			Status:      domain.EventStatusPending,
			EventType:   "Opening",
		}
		out = append(out, ev)
	}
	return out, nil
}

type jsonFeed struct {
	Items []jsonFeedItem `json:"items"`
}

type jsonFeedItem struct {
	ID            string `json:"id"`
	URL           string `json:"url"`
	ExternalURL   string `json:"external_url"`
	Title         string `json:"title"`
	ContentText   string `json:"content_text"`
	ContentHTML   string `json:"content_html"`
	Summary       string `json:"summary"`
	DatePublished string `json:"date_published"`
	DateModified  string `json:"date_modified"`
}

func parseJSONFeed(body []byte, since, now time.Time) ([]domain.Event, error) {
	var feed jsonFeed
	if err := json.Unmarshal(body, &feed); err != nil {
		return nil, err
	}

	var out []domain.Event
	for _, item := range feed.Items {
		link := strings.TrimSpace(item.URL)
		if link == "" {
			link = strings.TrimSpace(item.ExternalURL)
		}
		if link == "" {
			continue
		}
		published := parseFeedTime(item.DatePublished)
		if published.IsZero() {
			published = parseFeedTime(item.DateModified)
		}
		if !published.IsZero() && !since.IsZero() && !published.After(since) {
			continue
		}
		startsAt := now
		if !published.IsZero() {
			startsAt = published
		}
		desc := item.Summary
		if desc == "" {
			desc = item.ContentText
		}
		if desc == "" {
			desc = item.ContentHTML
		}
		if len(desc) > 2000 {
			desc = desc[:2000]
		}
		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = link
		}
		out = append(out, domain.Event{
			ID:          uuid.New(),
			Title:       title,
			Description: desc,
			SourceURL:   link,
			StartsAt:    startsAt,
			ScrapedAt:   now,
			Status:      domain.EventStatusPending,
			EventType:   "Opening",
		})
	}
	return out, nil
}

func parseFeedTime(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}
	}
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, raw); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}
