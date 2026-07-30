-- +goose Up
ALTER TABLE articles
    ADD COLUMN IF NOT EXISTS search_vector tsvector
    GENERATED ALWAYS AS (
        to_tsvector('english', coalesce(title, '') || ' ' || coalesce(excerpt, '') || ' ' || coalesce(body, ''))
    ) STORED;

CREATE INDEX IF NOT EXISTS idx_articles_search_vector ON articles USING GIN (search_vector);

-- +goose Down
DROP INDEX IF EXISTS idx_articles_search_vector;
ALTER TABLE articles DROP COLUMN IF EXISTS search_vector;
