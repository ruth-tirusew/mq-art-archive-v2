-- +goose Up

-- A questionnaire is versioned as a sequence of revisions. A revision is a full snapshot of the
-- questions: it is editable only while 'draft' and immutable once 'published', so an answer can
-- always be shown next to the wording the artist actually saw. At most one draft exists at a time;
-- `number` is assigned on publish.
CREATE TABLE IF NOT EXISTS questionnaire_revisions (
    id           UUID PRIMARY KEY,
    number       INT UNIQUE,
    status       TEXT NOT NULL CHECK (status IN ('draft', 'published')),
    -- 'substantive' (new or changed question) is what makes artists eligible for a nudge;
    -- 'minor' (typo, wording polish) never does.
    change_kind  TEXT NOT NULL DEFAULT 'substantive' CHECK (change_kind IN ('minor', 'substantive')),
    note         TEXT NOT NULL DEFAULT '',
    created_by   UUID REFERENCES users (id) ON DELETE SET NULL,
    published_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT questionnaire_revisions_publish_shape CHECK (
        (status = 'draft' AND number IS NULL AND published_at IS NULL)
        OR (status = 'published' AND number IS NOT NULL AND published_at IS NOT NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_questionnaire_one_draft
    ON questionnaire_revisions (status) WHERE status = 'draft';

-- `key` is the question's stable identity across revisions. Answers attach to the key, so an
-- unchanged question carries forward without the artist re-answering. A change that alters what is
-- being asked must get a new key; the old answers then stay attached to the old key.
CREATE TABLE IF NOT EXISTS questionnaire_questions (
    id          UUID PRIMARY KEY,
    revision_id UUID NOT NULL REFERENCES questionnaire_revisions (id) ON DELETE CASCADE,
    key         TEXT NOT NULL CHECK (key ~ '^[a-z0-9_]{1,64}$'),
    position    INT NOT NULL,
    type        TEXT NOT NULL CHECK (type IN ('short_text', 'long_text', 'choice', 'link')),
    prompt_en   TEXT NOT NULL,
    prompt_am   TEXT NOT NULL DEFAULT '',
    help_en     TEXT NOT NULL DEFAULT '',
    help_am     TEXT NOT NULL DEFAULT '',
    required    BOOLEAN NOT NULL DEFAULT FALSE,
    -- Private answers are visible to admins only, never on the public profile.
    is_public   BOOLEAN NOT NULL DEFAULT TRUE,
    max_length  INT NOT NULL CHECK (max_length > 0),
    -- Only for type 'choice': [{"value": "...", "label_en": "...", "label_am": "..."}].
    options     JSONB NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(options) = 'array'),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (revision_id, key),
    UNIQUE (revision_id, position)
);

-- One current answer per (artist, question key). `revision_id` is the revision the artist last
-- wrote this answer against; it is what "answered an older revision" is measured from.
-- `format` 'rich' holds HTML sanitised against a fixed allowlist on the server before it is stored
-- (only valid for long_text questions); 'plain' is verbatim text.
CREATE TABLE IF NOT EXISTS questionnaire_answers (
    user_id      UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    question_key TEXT NOT NULL,
    revision_id  UUID NOT NULL REFERENCES questionnaire_revisions (id),
    body         TEXT NOT NULL,
    format       TEXT NOT NULL DEFAULT 'plain' CHECK (format IN ('plain', 'rich')),
    lang         TEXT NOT NULL DEFAULT 'en' CHECK (lang IN ('en', 'am', 'mixed')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, question_key)
);

CREATE INDEX IF NOT EXISTS idx_questionnaire_answers_revision ON questionnaire_answers (revision_id);

-- A nudge is an admin-triggered, optional invitation to answer questions added in a revision.
-- Dismissal is remembered per revision, so a banner stays dismissed until the next nudge.
CREATE TABLE IF NOT EXISTS questionnaire_nudges (
    revision_id  UUID NOT NULL REFERENCES questionnaire_revisions (id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    sent_by      UUID REFERENCES users (id) ON DELETE SET NULL,
    sent_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    dismissed_at TIMESTAMPTZ,
    PRIMARY KEY (revision_id, user_id)
);

-- Which revision an application was submitted against, so the reviewer sees the answers beside the
-- wording they responded to.
ALTER TABLE onboarding_applications
    ADD COLUMN IF NOT EXISTS questionnaire_revision_id UUID REFERENCES questionnaire_revisions (id);

-- +goose Down
ALTER TABLE onboarding_applications DROP COLUMN IF EXISTS questionnaire_revision_id;
DROP TABLE IF EXISTS questionnaire_nudges;
DROP TABLE IF EXISTS questionnaire_answers;
DROP TABLE IF EXISTS questionnaire_questions;
DROP TABLE IF EXISTS questionnaire_revisions;
