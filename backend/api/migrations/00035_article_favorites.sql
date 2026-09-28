-- +goose Up
CREATE TABLE IF NOT EXISTS article_favorites (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    article_id UUID NOT NULL REFERENCES articles (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, article_id)
);

CREATE INDEX IF NOT EXISTS idx_article_favorites_article_id ON article_favorites (article_id);

-- +goose Down
DROP TABLE IF EXISTS article_favorites;
