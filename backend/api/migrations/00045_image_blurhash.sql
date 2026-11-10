-- +goose Up
-- Every image carries a blur hash: a ~30 character string the browser turns into a soft placeholder
-- while the real image loads. Empty means "not computed yet"; the backfill command fills those in.
ALTER TABLE media_assets ADD COLUMN IF NOT EXISTS blurhash TEXT NOT NULL DEFAULT '';
ALTER TABLE art_post_media ADD COLUMN IF NOT EXISTS blurhash TEXT NOT NULL DEFAULT '';
ALTER TABLE artist_profiles ADD COLUMN IF NOT EXISTS portrait_blurhash TEXT NOT NULL DEFAULT '';
ALTER TABLE events ADD COLUMN IF NOT EXISTS image_blurhash TEXT NOT NULL DEFAULT '';

-- Writes to art_post_media and artist_profiles look the hash up by URL.
CREATE INDEX IF NOT EXISTS idx_media_assets_secure_url ON media_assets (secure_url);

-- +goose Down
DROP INDEX IF EXISTS idx_media_assets_secure_url;
ALTER TABLE events DROP COLUMN IF EXISTS image_blurhash;
ALTER TABLE artist_profiles DROP COLUMN IF EXISTS portrait_blurhash;
ALTER TABLE art_post_media DROP COLUMN IF EXISTS blurhash;
ALTER TABLE media_assets DROP COLUMN IF EXISTS blurhash;
