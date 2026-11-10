-- +goose Up
-- An applicant may have only one pending application. If double submits already left several,
-- keep the newest and close the rest so the index can be built.
UPDATE onboarding_applications o
SET status = 'rejected',
    review_notes = 'Superseded by a newer application.',
    updated_at = now()
WHERE o.status = 'pending'
  AND EXISTS (
      SELECT 1 FROM onboarding_applications n
      WHERE n.applicant_id = o.applicant_id AND n.status = 'pending'
        AND (n.created_at, n.id) > (o.created_at, o.id)
  );

CREATE UNIQUE INDEX IF NOT EXISTS uq_onboarding_one_pending_per_applicant
    ON onboarding_applications (applicant_id)
    WHERE status = 'pending';

-- +goose Down
DROP INDEX IF EXISTS uq_onboarding_one_pending_per_applicant;
