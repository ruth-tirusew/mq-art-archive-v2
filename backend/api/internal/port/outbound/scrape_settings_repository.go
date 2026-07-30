package outbound

import (
	"context"

	"github.com/mq/api/internal/domain/settings"
)

type ScrapeSettingsRepository interface {
	Get(ctx context.Context) (*settings.ScrapeSettings, error)
	Upsert(ctx context.Context, s settings.ScrapeSettings) error
}

// EventSourceReloader rebuilds and swaps the in-process EventSource after settings change.
type EventSourceReloader interface {
	Reload(cfg settings.ScrapeSettings) error
}
