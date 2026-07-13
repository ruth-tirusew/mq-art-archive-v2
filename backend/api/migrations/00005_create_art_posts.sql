-- +goose Up
CREATE TABLE art_posts (
    id           UUID PRIMARY KEY,
    artist_id    UUID NOT NULL REFERENCES artist_profiles (id),
    title        TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    medium       TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL CHECK (status IN ('draft', 'published', 'archived')),
    published_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE art_post_media (
    id          UUID PRIMARY KEY,
    art_post_id UUID NOT NULL REFERENCES art_posts (id) ON DELETE CASCADE,
    url         TEXT NOT NULL,
    mime_type   TEXT NOT NULL DEFAULT '',
    width       INT NOT NULL DEFAULT 0,
    height      INT NOT NULL DEFAULT 0,
    sort_order  INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_art_posts_artist_id ON art_posts (artist_id);
CREATE INDEX idx_art_posts_status ON art_posts (status);
CREATE INDEX idx_art_post_media_post_id ON art_post_media (art_post_id);

-- +goose Down
DROP TABLE IF EXISTS art_post_media;
DROP TABLE IF EXISTS art_posts;
