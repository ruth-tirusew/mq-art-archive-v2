-- +goose Up
ALTER TABLE artist_profiles
    ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ;

UPDATE artist_profiles
SET approved_at = COALESCE(updated_at, created_at)
WHERE status = 'approved'
  AND approved_at IS NULL;

-- +goose Down
ALTER TABLE artist_profiles
    DROP COLUMN IF EXISTS approved_at;
