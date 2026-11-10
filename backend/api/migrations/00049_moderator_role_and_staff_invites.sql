-- +goose Up
-- A moderator is a staff member who can review and moderate content but cannot manage users or
-- change system settings (see the admin routes in the HTTP router).
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check
    CHECK (role IN ('public', 'artist', 'institution', 'contributor', 'moderator', 'admin'));

-- An admin invites a new moderator by email. Like artist invites, only the SHA-256 hash of the token is
-- stored, so a database leak cannot be replayed, and the link works once.
CREATE TABLE IF NOT EXISTS staff_invites (
    id         UUID PRIMARY KEY,
    token_hash TEXT NOT NULL UNIQUE,
    email      TEXT NOT NULL,
    role       TEXT NOT NULL CHECK (role IN ('moderator')),
    created_by UUID REFERENCES users (id) ON DELETE SET NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    claimed_by UUID REFERENCES users (id) ON DELETE SET NULL,
    claimed_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- One live invite per address: sending a new one replaces the old (the service revokes expired ones first).
CREATE UNIQUE INDEX IF NOT EXISTS uq_staff_invites_live_email
    ON staff_invites (LOWER(email))
    WHERE claimed_at IS NULL AND revoked_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS staff_invites;
UPDATE users SET role = 'contributor' WHERE role = 'moderator';
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check
    CHECK (role IN ('public', 'artist', 'institution', 'contributor', 'admin'));
