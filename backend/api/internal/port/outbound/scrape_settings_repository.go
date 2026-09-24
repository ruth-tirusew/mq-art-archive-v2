package outbound

import (
	"context"
	"time"

	"github.com/mq/api/internal/domain/settings"
)

type ScrapeSettingsRepository interface {
	Get(ctx context.Context) (*settings.ScrapeSettings, error)
	Upsert(ctx context.Context, s settings.ScrapeSettings) error
	// RecordRunResult is called by the scraper process after each sync attempt.
	// runErr is nil on success; on failure, last_success_at is left unchanged.
	RecordRunResult(ctx context.Context, runAt time.Time, runErr error) error
}

// EventSourceReloader rebuilds and swaps the in-process EventSource after settings change.
type EventSourceReloader interface {
	Reload(cfg settings.ScrapeSettings) error
}
