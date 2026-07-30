-- +goose Up
ALTER TABLE articles
    ADD COLUMN IF NOT EXISTS version INT NOT NULL DEFAULT 1;

CREATE TABLE article_revisions (
    id            UUID PRIMARY KEY,
    article_id    UUID NOT NULL REFERENCES articles (id) ON DELETE CASCADE,
    version       INT NOT NULL,
    editor_id     UUID NOT NULL,
    title         TEXT NOT NULL,
    body          TEXT NOT NULL DEFAULT '',
    slug          TEXT NOT NULL,
    category      TEXT NOT NULL DEFAULT 'General',
    excerpt       TEXT NOT NULL DEFAULT '',
    reading_time  INT NOT NULL DEFAULT 1,
    difficulty    TEXT NOT NULL DEFAULT 'Beginner',
    verified      BOOLEAN NOT NULL DEFAULT FALSE,
    status        TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT article_revisions_article_version_unique UNIQUE (article_id, version)
);

CREATE INDEX idx_article_revisions_article_id ON article_revisions (article_id, version DESC);

-- +goose Down
DROP TABLE IF EXISTS article_revisions;
ALTER TABLE articles DROP COLUMN IF EXISTS version;
