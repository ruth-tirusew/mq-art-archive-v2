-- +goose Up

-- An invite lets an admin bring a pre-vetted artist straight into the normal application flow:
-- the claimed invite pre-fills the name and handle and is recorded on the resulting application.
-- It does not bypass review; an admin still approves the application.
-- Only the SHA-256 hash of the token is stored, so a database leak cannot be replayed.
CREATE TABLE IF NOT EXISTS artist_invites (
    id               UUID PRIMARY KEY,
    token_hash       TEXT NOT NULL UNIQUE,
    email            TEXT NOT NULL,
    display_name     TEXT NOT NULL DEFAULT '',
    requested_handle TEXT NOT NULL DEFAULT '',
    created_by       UUID REFERENCES users (id) ON DELETE SET NULL,
    expires_at       TIMESTAMPTZ NOT NULL,
    claimed_by       UUID REFERENCES users (id) ON DELETE SET NULL,
    claimed_at       TIMESTAMPTZ,
    revoked_at       TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT artist_invites_claim_shape CHECK ((claimed_by IS NULL) = (claimed_at IS NULL))
);

CREATE INDEX IF NOT EXISTS idx_artist_invites_email ON artist_invites (LOWER(email));

-- A handle can be reserved by at most one live (unclaimed, unrevoked) invite. Expiry is checked by
-- the service, since a partial index cannot depend on the current time; resending revokes the old row.
CREATE UNIQUE INDEX IF NOT EXISTS uq_artist_invites_live_handle
    ON artist_invites (LOWER(requested_handle))
    WHERE requested_handle <> '' AND claimed_at IS NULL AND revoked_at IS NULL;

ALTER TABLE onboarding_applications
    ADD COLUMN IF NOT EXISTS invite_id UUID REFERENCES artist_invites (id);

-- +goose Down
ALTER TABLE onboarding_applications DROP COLUMN IF EXISTS invite_id;
DROP TABLE IF EXISTS artist_invites;
