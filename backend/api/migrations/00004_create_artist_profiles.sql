-- +goose Up
CREATE TABLE artist_profiles (
    id                UUID PRIMARY KEY,
    user_id           UUID NOT NULL REFERENCES users (id),
    slug              TEXT NOT NULL UNIQUE,
    display_name      TEXT NOT NULL,
    bio               TEXT NOT NULL DEFAULT '',
    contact_email     TEXT NOT NULL DEFAULT '',
    contact_phone     TEXT NOT NULL DEFAULT '',
    contact_website   TEXT NOT NULL DEFAULT '',
    contact_location  TEXT NOT NULL DEFAULT '',
    social_instagram  TEXT NOT NULL DEFAULT '',
    social_twitter    TEXT NOT NULL DEFAULT '',
    social_telegram   TEXT NOT NULL DEFAULT '',
    status            TEXT NOT NULL CHECK (status IN ('draft', 'pending', 'approved')),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_artist_profiles_user_id ON artist_profiles (user_id);
CREATE INDEX idx_artist_profiles_status ON artist_profiles (status);

-- +goose Down
DROP TABLE IF EXISTS artist_profiles;
