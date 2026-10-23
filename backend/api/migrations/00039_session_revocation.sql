-- +goose Up
-- Tokens issued before this moment are rejected. Set when a password is changed or reset, so a
-- stolen session cannot outlive the password change that was meant to lock it out.
ALTER TABLE users ADD COLUMN IF NOT EXISTS sessions_valid_after TIMESTAMPTZ;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS sessions_valid_after;
