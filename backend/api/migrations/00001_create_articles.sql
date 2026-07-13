-- +goose Up
CREATE TABLE articles (
    id         UUID PRIMARY KEY,
    slug       TEXT NOT NULL UNIQUE,
    title      TEXT NOT NULL,
    body       TEXT NOT NULL DEFAULT '',
    status     TEXT NOT NULL CHECK (status IN ('draft', 'published', 'archived')),
    author_id  UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_articles_status ON articles (status);
CREATE INDEX idx_articles_author_id ON articles (author_id);

-- +goose Down
DROP TABLE IF EXISTS articles;
