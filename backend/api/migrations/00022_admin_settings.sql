-- +goose Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS display_name TEXT,
    ADD COLUMN IF NOT EXISTS avatar_url TEXT;

CREATE TABLE user_notification_preferences (
    user_id                      UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    email_on_new_application     BOOLEAN NOT NULL DEFAULT TRUE,
    email_on_event_sync_summary  BOOLEAN NOT NULL DEFAULT FALSE,
    newsletter_enabled           BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at                   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE scrape_settings (
    id               INT PRIMARY KEY CHECK (id = 1),
    scrape_enabled   BOOLEAN NOT NULL DEFAULT FALSE,
    scrape_sources   TEXT[] NOT NULL DEFAULT '{}',
    scrape_user_agent TEXT NOT NULL DEFAULT 'mq-scraper/1.0',
    scrape_timeout_seconds INT NOT NULL DEFAULT 30,
    scrape_interval_seconds INT NOT NULL DEFAULT 21600,
    telegram_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    telegram_api_id  INT NOT NULL DEFAULT 0,
    telegram_api_hash TEXT NOT NULL DEFAULT '',
    telegram_channels TEXT[] NOT NULL DEFAULT '{}',
    telegram_keywords TEXT[] NOT NULL DEFAULT '{}',
    telegram_fetch_limit INT NOT NULL DEFAULT 50,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by       UUID REFERENCES users (id) ON DELETE SET NULL
);

-- +goose Down
DROP TABLE IF EXISTS scrape_settings;
DROP TABLE IF EXISTS user_notification_preferences;
ALTER TABLE users
    DROP COLUMN IF EXISTS display_name,
    DROP COLUMN IF EXISTS avatar_url;
