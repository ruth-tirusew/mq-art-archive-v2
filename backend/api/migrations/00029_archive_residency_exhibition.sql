-- +goose Up
ALTER TABLE art_posts ADD COLUMN IF NOT EXISTS residency TEXT NULL;
ALTER TABLE art_posts ADD COLUMN IF NOT EXISTS exhibition TEXT NULL;

-- +goose Down
ALTER TABLE art_posts DROP COLUMN IF EXISTS exhibition;
ALTER TABLE art_posts DROP COLUMN IF EXISTS residency;
