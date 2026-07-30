package events

import (
	"github.com/mq/api/config"
	"github.com/mq/api/internal/port/outbound"
)

// NewEventSource builds the composite EventSource from config flags.
func NewEventSource(cfg config.Config) outbound.EventSource {
	var sources []outbound.EventSource

	if cfg.ScrapeEnabled && len(cfg.ScrapeSources) > 0 {
		sources = append(sources, NewScraperRSS(RSSConfig{
			Sources:   cfg.ScrapeSources,
			UserAgent: cfg.ScrapeUserAgent,
			Timeout:   cfg.ScrapeTimeout,
		}))
	}

	if cfg.TelegramEnabled && len(cfg.TelegramChannels) > 0 {
		var fetcher ChannelMessageFetcher
		if cfg.TelegramAPIID > 0 && cfg.TelegramAPIHash != "" {
			fetcher = NewGotdFetcher(TelegramClientConfig{
				APIID:       cfg.TelegramAPIID,
				APIHash:     cfg.TelegramAPIHash,
				SessionPath: cfg.TelegramSessionPath,
			})
		}
		sources = append(sources, NewScraperTelegram(TelegramScraperConfig{
			Channels:   cfg.TelegramChannels,
			Keywords:   cfg.TelegramKeywords,
			FetchLimit: cfg.TelegramFetchLimit,
			Fetcher:    fetcher,
		}))
	}

	if len(sources) == 0 {
		return NewScraperNoop()
	}
	if len(sources) == 1 {
		return sources[0]
	}
	return NewComposite(sources...)
}
