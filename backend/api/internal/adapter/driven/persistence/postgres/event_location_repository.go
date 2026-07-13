package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/port/outbound"
)

type EventLocationRepository struct {
	pool *Pool
}

func NewEventLocationRepository(pool *Pool) outbound.EventLocationRepository {
	return &EventLocationRepository{pool: pool}
}

const eventLocationColumns = `id, name, pin_coords, created_at, updated_at`

func (r *EventLocationRepository) FindOrCreate(ctx context.Context, loc events.Location) (*events.Location, error) {
	name := strings.TrimSpace(loc.Name)
	if name == "" {
		return nil, fmt.Errorf("location name is required")
	}

	pinCoords := normalizePinCoords(loc.PinCoords)
	now := time.Now().UTC()

	row := r.pool.QueryRow(ctx, `
		SELECT `+eventLocationColumns+`
		FROM event_locations
		WHERE lower(trim(name)) = lower(trim($1))
		LIMIT 1
	`, name)

	existing, err := scanEventLocation(row)
	if err == nil {
		if len(pinCoords) == 2 && !existing.HasPin() {
			updated, err := r.updatePinCoords(ctx, existing.ID, pinCoords, now)
			if err != nil {
				return nil, err
			}
			return updated, nil
		}
		return &existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("find event location: %w", err)
	}

	id := loc.ID
	if id == uuid.Nil {
		id = uuid.New()
	}
	createdAt := loc.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	updatedAt := loc.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = now
	}

	inserted, err := r.insert(ctx, id, name, pinCoords, createdAt, updatedAt)
	if err != nil {
		return nil, err
	}
	return inserted, nil
}

func (r *EventLocationRepository) GetByID(ctx context.Context, id uuid.UUID) (*events.Location, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+eventLocationColumns+`
		FROM event_locations
		WHERE id = $1
	`, id)

	loc, err := scanEventLocation(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get event location: %w", err)
	}
	return &loc, nil
}

func (r *EventLocationRepository) insert(
	ctx context.Context,
	id uuid.UUID,
	name string,
	pinCoords []float64,
	createdAt, updatedAt time.Time,
) (*events.Location, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO event_locations (id, name, pin_coords, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+eventLocationColumns,
		id, name, pinCoordsToDB(pinCoords), createdAt, updatedAt,
	)

	loc, err := scanEventLocation(row)
	if err != nil {
		return nil, fmt.Errorf("insert event location: %w", err)
	}
	return &loc, nil
}

func (r *EventLocationRepository) updatePinCoords(
	ctx context.Context,
	id uuid.UUID,
	pinCoords []float64,
	updatedAt time.Time,
) (*events.Location, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE event_locations
		SET pin_coords = $2, updated_at = $3
		WHERE id = $1
		RETURNING `+eventLocationColumns,
		id, pinCoordsToDB(pinCoords), updatedAt,
	)

	loc, err := scanEventLocation(row)
	if err != nil {
		return nil, fmt.Errorf("update event location pin: %w", err)
	}
	return &loc, nil
}

func scanEventLocation(row scannable) (events.Location, error) {
	var loc events.Location
	var pinCoords []float64
	err := row.Scan(&loc.ID, &loc.Name, &pinCoords, &loc.CreatedAt, &loc.UpdatedAt)
	if err != nil {
		return events.Location{}, err
	}
	loc.PinCoords = pinCoords
	return loc, nil
}

func normalizePinCoords(coords []float64) []float64 {
	if len(coords) != 2 {
		return nil
	}
	return []float64{coords[0], coords[1]}
}

func pinCoordsToDB(coords []float64) []float64 {
	if len(coords) != 2 {
		return nil
	}
	return coords
}
