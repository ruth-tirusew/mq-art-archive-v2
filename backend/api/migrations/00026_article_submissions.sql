-- +goose Up
CREATE TABLE article_submissions (
    id UUID PRIMARY KEY,
    submitter_id UUID NOT NULL REFERENCES users(id),
    article_id UUID REFERENCES articles(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected')),
    review_notes TEXT NOT NULL DEFAULT '',
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_article_submissions_submitter ON article_submissions(submitter_id, created_at DESC);
CREATE INDEX idx_article_submissions_pending ON article_submissions(created_at) WHERE status = 'pending';

-- +goose Down
DROP TABLE IF EXISTS article_submissions;
