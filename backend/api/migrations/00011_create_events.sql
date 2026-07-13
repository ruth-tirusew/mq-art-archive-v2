-- +goose Up
CREATE TABLE event_locations (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    pin_coords  DOUBLE PRECISION[],
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT event_locations_pin_coords_len CHECK (
        pin_coords IS NULL OR array_length(pin_coords, 1) = 2
    )
);

CREATE INDEX idx_event_locations_name ON event_locations (lower(trim(name)));

CREATE TABLE events (
    id            UUID PRIMARY KEY,
    title         TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    source_url    TEXT NOT NULL UNIQUE,
    image_url     TEXT,
    location_id   UUID REFERENCES event_locations (id) ON DELETE SET NULL,
    starts_at     TIMESTAMPTZ NOT NULL,
    ends_at       TIMESTAMPTZ,
    scraped_at    TIMESTAMPTZ NOT NULL,
    status        TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    review_notes  TEXT NOT NULL DEFAULT '',
    reviewed_by   UUID,
    reviewed_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    search_vector tsvector GENERATED ALWAYS AS (
        to_tsvector('english', coalesce(title, '') || ' ' || coalesce(description, ''))
    ) STORED
);

CREATE INDEX idx_events_starts_at ON events (starts_at);
CREATE INDEX idx_events_status ON events (status);
CREATE INDEX idx_events_location_id ON events (location_id);
CREATE INDEX idx_events_search_vector ON events USING GIN (search_vector);

-- +goose Down
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS event_locations;
