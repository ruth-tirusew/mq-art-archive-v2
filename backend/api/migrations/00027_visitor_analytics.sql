-- +goose Up
CREATE TABLE page_view_daily (
    entity_type TEXT NOT NULL CHECK (entity_type IN ('artist','post','article')),
    entity_id UUID NOT NULL,
    day DATE NOT NULL,
    count BIGINT NOT NULL DEFAULT 0 CHECK (count >= 0),
    PRIMARY KEY (entity_type, entity_id, day)
);
CREATE TABLE page_view_dedupe (
    hash TEXT PRIMARY KEY,
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_page_view_dedupe_expiry ON page_view_dedupe(expires_at);

-- +goose Down
DROP TABLE IF EXISTS page_view_dedupe;
DROP TABLE IF EXISTS page_view_daily;
