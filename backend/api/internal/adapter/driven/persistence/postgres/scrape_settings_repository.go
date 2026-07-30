package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mq/api/internal/domain/settings"
	"github.com/mq/api/internal/port/outbound"
)

type ScrapeSettingsRepository struct {
	pool *Pool
}

func NewScrapeSettingsRepository(pool *Pool) outbound.ScrapeSettingsRepository {
	return &ScrapeSettingsRepository{pool: pool}
}

func (r *ScrapeSettingsRepository) Get(ctx context.Context) (*settings.ScrapeSettings, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT
			scrape_enabled, scrape_sources, scrape_user_agent, scrape_timeout_seconds, scrape_interval_seconds,
			telegram_enabled, telegram_api_id, telegram_api_hash, telegram_channels, telegram_keywords, telegram_fetch_limit,
			updated_at, updated_by
		FROM scrape_settings
		WHERE id = 1
	`)

	var s settings.ScrapeSettings
	var updatedBy *uuid.UUID
	err := row.Scan(
		&s.ScrapeEnabled, &s.ScrapeSources, &s.ScrapeUserAgent, &s.ScrapeTimeoutSeconds, &s.ScrapeIntervalSeconds,
		&s.TelegramEnabled, &s.TelegramAPIID, &s.TelegramAPIHash, &s.TelegramChannels, &s.TelegramKeywords, &s.TelegramFetchLimit,
		&s.UpdatedAt, &updatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get scrape settings: %w", err)
	}
	s.UpdatedBy = updatedBy
	if s.ScrapeSources == nil {
		s.ScrapeSources = []string{}
	}
	if s.TelegramChannels == nil {
		s.TelegramChannels = []string{}
	}
	if s.TelegramKeywords == nil {
		s.TelegramKeywords = []string{}
	}
	return &s, nil
}

func (r *ScrapeSettingsRepository) Upsert(ctx context.Context, s settings.ScrapeSettings) error {
	now := time.Now().UTC()
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = now
	}
	if s.ScrapeSources == nil {
		s.ScrapeSources = []string{}
	}
	if s.TelegramChannels == nil {
		s.TelegramChannels = []string{}
	}
	if s.TelegramKeywords == nil {
		s.TelegramKeywords = []string{}
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO scrape_settings (
			id,
			scrape_enabled, scrape_sources, scrape_user_agent, scrape_timeout_seconds, scrape_interval_seconds,
			telegram_enabled, telegram_api_id, telegram_api_hash, telegram_channels, telegram_keywords, telegram_fetch_limit,
			updated_at, updated_by
		) VALUES (
			1,
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10, $11,
			$12, $13
		)
		ON CONFLICT (id) DO UPDATE SET
			scrape_enabled = EXCLUDED.scrape_enabled,
			scrape_sources = EXCLUDED.scrape_sources,
			scrape_user_agent = EXCLUDED.scrape_user_agent,
			scrape_timeout_seconds = EXCLUDED.scrape_timeout_seconds,
			scrape_interval_seconds = EXCLUDED.scrape_interval_seconds,
			telegram_enabled = EXCLUDED.telegram_enabled,
			telegram_api_id = EXCLUDED.telegram_api_id,
			telegram_api_hash = EXCLUDED.telegram_api_hash,
			telegram_channels = EXCLUDED.telegram_channels,
			telegram_keywords = EXCLUDED.telegram_keywords,
			telegram_fetch_limit = EXCLUDED.telegram_fetch_limit,
			updated_at = EXCLUDED.updated_at,
			updated_by = EXCLUDED.updated_by
	`,
		s.ScrapeEnabled, s.ScrapeSources, s.ScrapeUserAgent, s.ScrapeTimeoutSeconds, s.ScrapeIntervalSeconds,
		s.TelegramEnabled, s.TelegramAPIID, s.TelegramAPIHash, s.TelegramChannels, s.TelegramKeywords, s.TelegramFetchLimit,
		s.UpdatedAt, s.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("upsert scrape settings: %w", err)
	}
	return nil
}
