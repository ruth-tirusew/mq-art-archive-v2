-- +goose Up
CREATE TABLE onboarding_applications (
    id             UUID PRIMARY KEY,
    applicant_id   UUID NOT NULL,
    applicant_type TEXT NOT NULL CHECK (applicant_type IN ('artist', 'institution')),
    display_name   TEXT NOT NULL,
    notes          TEXT NOT NULL DEFAULT '',
    status         TEXT NOT NULL CHECK (status IN ('pending', 'approved', 'rejected')),
    reviewed_by    UUID,
    reviewed_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_onboarding_applications_status ON onboarding_applications (status);

-- +goose Down
DROP TABLE IF EXISTS onboarding_applications;
