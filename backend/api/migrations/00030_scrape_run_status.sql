-- +goose Up
ALTER TABLE scrape_settings ADD COLUMN IF NOT EXISTS last_run_at TIMESTAMPTZ NULL;
ALTER TABLE scrape_settings ADD COLUMN IF NOT EXISTS last_success_at TIMESTAMPTZ NULL;
ALTER TABLE scrape_settings ADD COLUMN IF NOT EXISTS last_error TEXT NULL;

-- +goose Down
ALTER TABLE scrape_settings DROP COLUMN IF EXISTS last_error;
ALTER TABLE scrape_settings DROP COLUMN IF EXISTS last_success_at;
ALTER TABLE scrape_settings DROP COLUMN IF EXISTS last_run_at;
