-- +goose Up
CREATE TABLE IF NOT EXISTS article_highlights (
    id UUID PRIMARY KEY,
    article_id UUID NOT NULL REFERENCES articles (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    quoted_text TEXT NOT NULL,
    text_offset_start INT NOT NULL,
    text_offset_end INT NOT NULL,
    article_version_at_anchor INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_article_highlights_article_id ON article_highlights (article_id);
CREATE INDEX IF NOT EXISTS idx_article_highlights_user_id ON article_highlights (user_id);

-- article_comments covers both anchored (Medium-style margin) comments and general,
-- unanchored "responses" at the bottom of the article — is_general distinguishes them.
-- Anchor columns are nullable rather than defaulted to 0/'' for a general comment, so a
-- comment can never be mistaken for one anchored at the very start of the article.
CREATE TABLE IF NOT EXISTS article_comments (
    id UUID PRIMARY KEY,
    article_id UUID NOT NULL REFERENCES articles (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    is_general BOOLEAN NOT NULL DEFAULT FALSE,
    quoted_text TEXT NULL,
    text_offset_start INT NULL,
    text_offset_end INT NULL,
    article_version_at_anchor INT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT article_comments_anchor_shape CHECK (
        is_general = TRUE
        OR (quoted_text IS NOT NULL AND text_offset_start IS NOT NULL AND text_offset_end IS NOT NULL
            AND article_version_at_anchor IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_article_comments_article_id ON article_comments (article_id);

-- +goose Down
DROP TABLE IF EXISTS article_comments;
DROP TABLE IF EXISTS article_highlights;
