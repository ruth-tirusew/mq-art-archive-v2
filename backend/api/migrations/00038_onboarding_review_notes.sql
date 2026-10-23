-- +goose Up
ALTER TABLE onboarding_applications ADD COLUMN IF NOT EXISTS review_notes TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE onboarding_applications DROP COLUMN IF EXISTS review_notes;
