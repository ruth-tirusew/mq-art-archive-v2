-- +goose Up
ALTER TABLE onboarding_applications ADD COLUMN requested_handle TEXT;
CREATE UNIQUE INDEX uq_onboarding_active_requested_handle
    ON onboarding_applications (LOWER(requested_handle))
    WHERE requested_handle IS NOT NULL AND status IN ('pending', 'approved');

-- +goose Down
DROP INDEX IF EXISTS uq_onboarding_active_requested_handle;
ALTER TABLE onboarding_applications DROP COLUMN IF EXISTS requested_handle;
