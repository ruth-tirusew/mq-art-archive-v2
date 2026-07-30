-- +goose Up
ALTER TABLE artist_profiles
    ADD COLUMN IF NOT EXISTS influences TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS in_residence BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS residency_place TEXT,
    ADD COLUMN IF NOT EXISTS open_for_commission BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE artist_profiles
    DROP COLUMN IF EXISTS open_for_commission,
    DROP COLUMN IF EXISTS residency_place,
    DROP COLUMN IF EXISTS in_residence,
    DROP COLUMN IF EXISTS influences;
