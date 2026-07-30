package inbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/settings"
)

type SettingsService interface {
	GetScrapeSettings(ctx context.Context) (*settings.ScrapeSettingsView, error)
	UpdateScrapeSettings(ctx context.Context, updatedBy uuid.UUID, update settings.ScrapeSettingsUpdate) (*settings.ScrapeSettingsView, error)
}
