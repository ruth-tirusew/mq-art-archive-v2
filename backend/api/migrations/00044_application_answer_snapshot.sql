-- +goose Up
-- The answers an applicant submitted, frozen with the application. Live answers keep changing
-- (artists edit them after approval, and re-applicants overwrite them), so a reviewer looking at an
-- old decision must read this copy. The column is opaque to onboarding; the questionnaire context
-- owns its shape.
ALTER TABLE onboarding_applications
    ADD COLUMN IF NOT EXISTS questionnaire_answers JSONB;

-- Best available copy for applications that predate the snapshot: the current answers.
UPDATE onboarding_applications a
SET questionnaire_answers = COALESCE((
    SELECT jsonb_agg(jsonb_build_object(
        'key', qa.question_key, 'body', qa.body, 'format', qa.format, 'lang', qa.lang,
        'revision_id', qa.revision_id, 'updated_at', qa.updated_at
    ) ORDER BY qa.question_key)
    FROM questionnaire_answers qa WHERE qa.user_id = a.applicant_id
), '[]'::jsonb)
WHERE a.questionnaire_revision_id IS NOT NULL AND a.questionnaire_answers IS NULL;

-- +goose Down
ALTER TABLE onboarding_applications DROP COLUMN IF EXISTS questionnaire_answers;
