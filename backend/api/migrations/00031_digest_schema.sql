-- +goose Up

-- Articles have no publish timestamp today (only created_at/updated_at, and updated_at
-- changes on every edit), so "published this week" can't be answered. Backfill existing
-- published articles from created_at as the best available approximation.
ALTER TABLE articles ADD COLUMN IF NOT EXISTS published_at TIMESTAMPTZ NULL;
UPDATE articles SET published_at = created_at WHERE status = 'published' AND published_at IS NULL;

-- Links a subscriber's Telegram chat to their account. Nullable/unique: most users won't
-- have one until they opt in through the bot.
ALTER TABLE user_notification_preferences ADD COLUMN IF NOT EXISTS telegram_chat_id TEXT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS user_notification_preferences_telegram_chat_id_key
    ON user_notification_preferences (telegram_chat_id) WHERE telegram_chat_id IS NOT NULL;

-- One-time tokens used to link a Telegram chat to a site account: the settings page
-- generates a token and a t.me/<bot>?start=<token> link; the bot resolves it on /start.
CREATE TABLE telegram_link_tokens (
    token       TEXT PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ NULL
);
CREATE INDEX IF NOT EXISTS telegram_link_tokens_user_id_idx ON telegram_link_tokens (user_id);

-- One row per digest send. period_start/period_end define the content window; completed_at
-- is set once every recipient has been attempted, so a crash mid-run is visible as a row
-- with started_at set and completed_at still null.
CREATE TABLE digest_runs (
    id           UUID PRIMARY KEY,
    period_start TIMESTAMPTZ NOT NULL,
    period_end   TIMESTAMPTZ NOT NULL,
    started_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ NULL
);

-- One row per recipient per channel per run, so a retried run can skip anyone already
-- sent to instead of mailing/messaging everyone twice.
CREATE TABLE digest_deliveries (
    id              UUID PRIMARY KEY,
    run_id          UUID NOT NULL REFERENCES digest_runs (id) ON DELETE CASCADE,
    recipient_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    channel         TEXT NOT NULL CHECK (channel IN ('email', 'telegram')),
    sent_at         TIMESTAMPTZ NULL,
    error           TEXT NULL,
    UNIQUE (run_id, recipient_id, channel)
);
CREATE INDEX IF NOT EXISTS digest_deliveries_run_id_idx ON digest_deliveries (run_id);

-- +goose Down
DROP TABLE IF EXISTS digest_deliveries;
DROP TABLE IF EXISTS digest_runs;
DROP TABLE IF EXISTS telegram_link_tokens;
DROP INDEX IF EXISTS user_notification_preferences_telegram_chat_id_key;
ALTER TABLE user_notification_preferences DROP COLUMN IF EXISTS telegram_chat_id;
ALTER TABLE articles DROP COLUMN IF EXISTS published_at;
