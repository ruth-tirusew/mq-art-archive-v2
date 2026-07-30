package events

import (
	"context"
	"sync"
	"time"

	"github.com/mq/api/config"
	domain "github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/domain/settings"
	"github.com/mq/api/internal/port/outbound"
)

// SwappableEventSource lets scrape settings hot-reload the active EventSource.
type SwappableEventSource struct {
	mu      sync.RWMutex
	current outbound.EventSource
}

func NewSwappableEventSource(initial outbound.EventSource) *SwappableEventSource {
	if initial == nil {
		initial = NewScraperNoop()
	}
	return &SwappableEventSource{current: initial}
}

func (s *SwappableEventSource) FetchEvents(ctx context.Context, since time.Time) ([]domain.Event, error) {
	s.mu.RLock()
	src := s.current
	s.mu.RUnlock()
	return src.FetchEvents(ctx, since)
}

func (s *SwappableEventSource) Swap(next outbound.EventSource) {
	if next == nil {
		next = NewScraperNoop()
	}
	s.mu.Lock()
	s.current = next
	s.mu.Unlock()
}

// EventSourceReloader rebuilds EventSource from scrape settings and swaps it in.
type EventSourceReloader struct {
	swapper     *SwappableEventSource
	sessionPath string
}

func NewEventSourceReloader(swapper *SwappableEventSource, sessionPath string) *EventSourceReloader {
	return &EventSourceReloader{swapper: swapper, sessionPath: sessionPath}
}

func (r *EventSourceReloader) Reload(cfg settings.ScrapeSettings) error {
	source := NewEventSourceFromSettings(cfg, r.sessionPath)
	r.swapper.Swap(source)
	return nil
}

// NewEventSourceFromSettings builds an EventSource from persisted scrape settings.
func NewEventSourceFromSettings(s settings.ScrapeSettings, sessionPath string) outbound.EventSource {
	cfg := config.Config{
		ScrapeEnabled:       s.ScrapeEnabled,
		ScrapeSources:       s.ScrapeSources,
		ScrapeUserAgent:     s.ScrapeUserAgent,
		ScrapeTimeout:       time.Duration(s.ScrapeTimeoutSeconds) * time.Second,
		TelegramEnabled:     s.TelegramEnabled,
		TelegramAPIID:       s.TelegramAPIID,
		TelegramAPIHash:     s.TelegramAPIHash,
		TelegramSessionPath: sessionPath,
		TelegramChannels:    s.TelegramChannels,
		TelegramKeywords:    s.TelegramKeywords,
		TelegramFetchLimit:  s.TelegramFetchLimit,
	}
	return NewEventSource(cfg)
}

// SettingsFromConfig maps env config into a domain scrape settings seed.
func SettingsFromConfig(cfg config.Config) settings.ScrapeSettings {
	timeout := int(cfg.ScrapeTimeout / time.Second)
	if timeout <= 0 {
		timeout = 30
	}
	interval := int(cfg.ScrapeInterval / time.Second)
	if interval <= 0 {
		interval = 21600
	}
	return settings.ScrapeSettings{
		ScrapeEnabled:         cfg.ScrapeEnabled,
		ScrapeSources:         append([]string(nil), cfg.ScrapeSources...),
		ScrapeUserAgent:       cfg.ScrapeUserAgent,
		ScrapeTimeoutSeconds:  timeout,
		ScrapeIntervalSeconds: interval,
		TelegramEnabled:       cfg.TelegramEnabled,
		TelegramAPIID:         cfg.TelegramAPIID,
		TelegramAPIHash:       cfg.TelegramAPIHash,
		TelegramChannels:      append([]string(nil), cfg.TelegramChannels...),
		TelegramKeywords:      append([]string(nil), cfg.TelegramKeywords...),
		TelegramFetchLimit:    cfg.TelegramFetchLimit,
	}
}
