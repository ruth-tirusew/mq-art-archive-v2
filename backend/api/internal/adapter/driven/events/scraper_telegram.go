package events

import (
	"context"
	"log"
	"strings"
	"time"

	domain "github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/port/outbound"
)

type TelegramScraperConfig struct {
	Channels   []string
	Keywords   []string
	FetchLimit int
	Fetcher    ChannelMessageFetcher
}

type ScraperTelegram struct {
	channels   []string
	keywords   []string
	fetchLimit int
	fetcher    ChannelMessageFetcher
}

func NewScraperTelegram(cfg TelegramScraperConfig) outbound.EventSource {
	limit := cfg.FetchLimit
	if limit <= 0 {
		limit = 50
	}
	return &ScraperTelegram{
		channels:   cfg.Channels,
		keywords:   cfg.Keywords,
		fetchLimit: limit,
		fetcher:    cfg.Fetcher,
	}
}

func (s *ScraperTelegram) FetchEvents(ctx context.Context, since time.Time) ([]domain.Event, error) {
	if s.fetcher == nil {
		return []domain.Event{}, nil
	}

	var all []domain.Event
	for _, channel := range s.channels {
		channel = strings.TrimPrefix(strings.TrimSpace(channel), "@")
		if channel == "" {
			continue
		}
		msgs, err := s.fetcher.FetchChannelMessages(ctx, channel, since, s.fetchLimit)
		if err != nil {
			log.Printf("telegram channel %s: %v", channel, err)
			continue
		}
		for _, msg := range msgs {
			ev, ok := ParseTelegramMessage(msg, s.keywords)
			if !ok {
				continue
			}
			all = append(all, *ev)
		}
	}
	if all == nil {
		all = []domain.Event{}
	}
	return all, nil
}
