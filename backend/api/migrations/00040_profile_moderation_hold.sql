-- +goose Up
-- Set when an editor takes a live profile down. While true the artist cannot publish it again
-- themselves; an editor re-approving it clears the flag.
ALTER TABLE artist_profiles ADD COLUMN IF NOT EXISTS moderation_hold BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE artist_profiles DROP COLUMN IF EXISTS moderation_hold;
