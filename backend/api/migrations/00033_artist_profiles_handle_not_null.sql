-- +goose Up
-- Same situation as events.slug (migration 00032): internal/usecase writes
-- nullIfEmpty(p.Handle, p.Slug), so a handle is always written, and reads only work
-- because of a defensive COALESCE(handle, '') in artistProfileColumns. Enforce the
-- guarantee at the schema level too. Migration 00012 already backfilled handle = slug
-- for any pre-existing NULLs; this is a defensive re-run in case any slipped through since.
UPDATE artist_profiles SET handle = slug WHERE handle IS NULL;
ALTER TABLE artist_profiles ALTER COLUMN handle SET NOT NULL;

-- +goose Down
ALTER TABLE artist_profiles ALTER COLUMN handle DROP NOT NULL;
