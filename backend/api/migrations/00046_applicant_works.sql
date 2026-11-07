-- +goose Up
-- Work an applicant uploads to show with their application (an artist needs at least five). Each
-- row is one already-uploaded image plus what the applicant says about it. Deleting the image
-- removes the row.
CREATE TABLE IF NOT EXISTS applicant_works (
    id             UUID PRIMARY KEY,
    user_id        UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    media_asset_id UUID NOT NULL REFERENCES media_assets (id) ON DELETE CASCADE,
    title          TEXT NOT NULL,
    medium         TEXT NOT NULL DEFAULT '',
    year           INT,
    position       INT NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, media_asset_id)
);
CREATE INDEX IF NOT EXISTS idx_applicant_works_user ON applicant_works (user_id, position);

-- The works as they were when the application was submitted, so a decision made last year still
-- shows what was sent. Opaque to the database; the onboarding code owns its shape.
ALTER TABLE onboarding_applications ADD COLUMN IF NOT EXISTS works JSONB;

-- +goose Down
ALTER TABLE onboarding_applications DROP COLUMN IF EXISTS works;
DROP TABLE IF EXISTS applicant_works;
