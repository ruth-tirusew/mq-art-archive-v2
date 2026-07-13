-- +goose Up
CREATE TABLE institution_profiles (
    id                UUID PRIMARY KEY,
    user_id           UUID NOT NULL REFERENCES users (id),
    slug              TEXT NOT NULL UNIQUE,
    name              TEXT NOT NULL,
    description       TEXT NOT NULL DEFAULT '',
    contact_email     TEXT NOT NULL DEFAULT '',
    contact_phone     TEXT NOT NULL DEFAULT '',
    contact_website   TEXT NOT NULL DEFAULT '',
    contact_location  TEXT NOT NULL DEFAULT '',
    status            TEXT NOT NULL CHECK (status IN ('draft', 'pending', 'approved')),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_institution_profiles_user_id ON institution_profiles (user_id);
CREATE INDEX idx_institution_profiles_status ON institution_profiles (status);

-- +goose Down
DROP TABLE IF EXISTS institution_profiles;
