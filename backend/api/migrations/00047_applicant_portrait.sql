-- +goose Up
-- An artist applicant also shows a portrait. It is stored beside their works (same upload, same
-- ownership checks) but is not one of them: it is not counted towards the minimum and is never
-- published as an artwork. At most one per user, and sending a new one replaces it.
ALTER TABLE applicant_works ADD COLUMN IF NOT EXISTS kind TEXT NOT NULL DEFAULT 'work';
ALTER TABLE applicant_works DROP CONSTRAINT IF EXISTS applicant_works_kind_check;
ALTER TABLE applicant_works ADD CONSTRAINT applicant_works_kind_check CHECK (kind IN ('work', 'portrait'));
CREATE UNIQUE INDEX IF NOT EXISTS idx_applicant_works_one_portrait
    ON applicant_works (user_id) WHERE kind = 'portrait';

-- The portrait as it was when the application was submitted, frozen like the works.
ALTER TABLE onboarding_applications ADD COLUMN IF NOT EXISTS portrait JSONB;

-- +goose Down
ALTER TABLE onboarding_applications DROP COLUMN IF EXISTS portrait;
DROP INDEX IF EXISTS idx_applicant_works_one_portrait;
DELETE FROM applicant_works WHERE kind = 'portrait';
ALTER TABLE applicant_works DROP CONSTRAINT IF EXISTS applicant_works_kind_check;
ALTER TABLE applicant_works DROP COLUMN IF EXISTS kind;
