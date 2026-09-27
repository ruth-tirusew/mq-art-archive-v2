-- +goose Up

-- Distinguishes a submission proposing a brand-new article from one editing an existing
-- one. Today Approve overwrites submission.ArticleID with the newly-created article's ID,
-- so after approval a new-article submission is indistinguishable from an edit — this
-- column is captured once, at submit time, so that distinction survives approval.
ALTER TABLE article_submissions ADD COLUMN IF NOT EXISTS kind TEXT NOT NULL DEFAULT 'edit'
    CHECK (kind IN ('new', 'edit'));
UPDATE article_submissions SET kind = 'new' WHERE article_id IS NULL AND kind = 'edit';

-- The target article's version at submission time, for an edit submission (NULL for a
-- new-article submission — there's nothing to conflict with yet). Compared against the
-- article's current version at approval time so a submission written against a
-- since-changed article can't silently overwrite it.
ALTER TABLE article_submissions ADD COLUMN IF NOT EXISTS based_on_version INT NULL;

-- Links a revision back to the submission that produced it (NULL for revisions from a
-- direct admin edit, which isn't submission-driven), so the article history can show which
-- submission a change came from in addition to who approved it (already on the submission
-- itself, via reviewed_by).
ALTER TABLE article_revisions ADD COLUMN IF NOT EXISTS submission_id UUID NULL
    REFERENCES article_submissions (id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE article_revisions DROP COLUMN IF EXISTS submission_id;
ALTER TABLE article_submissions DROP COLUMN IF EXISTS based_on_version;
ALTER TABLE article_submissions DROP COLUMN IF EXISTS kind;
