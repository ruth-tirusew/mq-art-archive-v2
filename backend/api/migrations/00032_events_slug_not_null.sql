-- +goose Up
-- The application (internal/usecase/events) always generates a slug before inserting or
-- updating an event, so no row should ever have NULL here — but the column allowed it
-- since migration 00012 added it without NOT NULL. Backfill defensively before adding the
-- constraint, in case any row predates that guarantee.
UPDATE events SET slug = id::text WHERE slug IS NULL;
ALTER TABLE events ALTER COLUMN slug SET NOT NULL;

-- +goose Down
ALTER TABLE events ALTER COLUMN slug DROP NOT NULL;
