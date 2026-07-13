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

type EventRepository struct {
	pool *Pool
}

func NewEventRepository(pool *Pool) outbound.EventRepository {
	return &EventRepository{pool: pool}
}

const eventColumns = `
	id, title, description, source_url, image_url, location_id, slug, event_type, venue, city,
	starts_at, ends_at, scraped_at, status, review_notes,
	reviewed_by, reviewed_at, created_at, updated_at
`

const eventSelectColumns = `
	e.id, e.title, e.description, e.source_url, e.image_url, e.location_id, e.slug, e.event_type, e.venue, e.city,
	e.starts_at, e.ends_at, e.scraped_at, e.status, e.review_notes,
	e.reviewed_by, e.reviewed_at, e.created_at, e.updated_at
`

const eventLocationJoinColumns = `l.id, l.name, l.pin_coords, l.created_at, l.updated_at`

func (r *EventRepository) UpsertBySourceURL(ctx context.Context, event events.Event) (*events.Event, error) {
	now := time.Now().UTC()
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = now
	}
	event.UpdatedAt = now
	if event.Status == "" {
		event.Status = events.EventStatusPending
	}

	row := r.pool.QueryRow(ctx, `
		INSERT INTO events (
			id, title, description, source_url, image_url, location_id, slug, event_type, venue, city,
			starts_at, ends_at, scraped_at, status, review_notes,
			reviewed_by, reviewed_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
		ON CONFLICT (source_url) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			image_url = COALESCE(EXCLUDED.image_url, events.image_url),
			location_id = COALESCE(EXCLUDED.location_id, events.location_id),
			starts_at = EXCLUDED.starts_at,
			ends_at = EXCLUDED.ends_at,
			scraped_at = EXCLUDED.scraped_at,
			updated_at = EXCLUDED.updated_at
		WHERE events.status = 'pending'
		RETURNING `+eventColumns,
		event.ID, event.Title, event.Description, event.SourceURL, event.ImageURL, event.LocationID,
		event.Slug, event.EventType, event.Venue, event.City,
		event.StartsAt, event.EndsAt, event.ScrapedAt, string(event.Status), event.ReviewNotes,
		event.ReviewedBy, event.ReviewedAt, event.CreatedAt, event.UpdatedAt,
	)

	saved, err := scanEvent(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return r.GetBySourceURL(ctx, event.SourceURL)
		}
		return nil, fmt.Errorf("upsert event: %w", err)
	}
	return r.attachLocation(ctx, saved)
}

func (r *EventRepository) List(ctx context.Context, filter events.ListFilter) ([]events.Event, error) {
	filter = normalizeEventListFilter(filter)

	query := `
		SELECT ` + eventSelectColumns + `, ` + eventLocationJoinColumns + `
		FROM events e
		LEFT JOIN event_locations l ON l.id = e.location_id
		WHERE 1=1`

	args := []any{}
	argPos := 1

	if filter.Status != nil {
		query += fmt.Sprintf(" AND e.status = $%d", argPos)
		args = append(args, string(*filter.Status))
		argPos++
	}
	if filter.UpcomingOnly {
		query += " AND e.starts_at >= NOW()"
	}
	if filter.EventType != "" {
		query += fmt.Sprintf(" AND e.event_type ILIKE $%d", argPos)
		args = append(args, "%"+filter.EventType+"%")
		argPos++
	}
	if filter.Query != "" {
		query += fmt.Sprintf(" AND (e.title ILIKE $%d OR e.description ILIKE $%d OR e.venue ILIKE $%d OR e.city ILIKE $%d)", argPos, argPos, argPos, argPos)
		args = append(args, "%"+filter.Query+"%")
		argPos++
	}

	switch {
	case filter.UpcomingOnly:
		query += " ORDER BY e.starts_at ASC"
	case filter.Status != nil && *filter.Status == events.EventStatusPending:
		query += " ORDER BY e.scraped_at DESC"
	default:
		query += " ORDER BY e.starts_at DESC"
	}

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filter.Limit)
		argPos++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, filter.Offset)
	}

	return r.queryEvents(ctx, query, args...)
}

func (r *EventRepository) GetBySlug(ctx context.Context, slug string) (*events.Event, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+eventSelectColumns+`, `+eventLocationJoinColumns+`
		FROM events e
		LEFT JOIN event_locations l ON l.id = e.location_id
		WHERE e.slug = $1 AND e.status = $2
	`, slug, string(events.EventStatusApproved))

	event, err := scanEventWithLocation(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get event by slug: %w", err)
	}
	return &event, nil
}

func (r *EventRepository) Search(ctx context.Context, query string, limit int) ([]events.Event, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []events.Event{}, nil
	}
	if limit <= 0 {
		limit = 20
	}

	sql := `
		SELECT ` + eventSelectColumns + `, ` + eventLocationJoinColumns + `
		FROM events e
		LEFT JOIN event_locations l ON l.id = e.location_id
		WHERE e.status = $1
		  AND e.search_vector @@ plainto_tsquery('english', $2)
		ORDER BY ts_rank(e.search_vector, plainto_tsquery('english', $2)) DESC, e.starts_at ASC
		LIMIT $3`

	return r.queryEvents(ctx, sql, string(events.EventStatusApproved), query, limit)
}

func normalizeEventListFilter(filter events.ListFilter) events.ListFilter {
	if filter.Status == nil && !filter.UpcomingOnly && filter.Limit == 0 {
		return events.PublicUpcomingFilter()
	}
	return filter
}

func (r *EventRepository) GetByID(ctx context.Context, id uuid.UUID) (*events.Event, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+eventSelectColumns+`, `+eventLocationJoinColumns+`
		FROM events e
		LEFT JOIN event_locations l ON l.id = e.location_id
		WHERE e.id = $1
	`, id)

	event, err := scanEventWithLocation(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get event: %w", err)
	}
	return &event, nil
}

func (r *EventRepository) Save(ctx context.Context, event events.Event) (*events.Event, error) {
	event.UpdatedAt = time.Now().UTC()
	if event.CreatedAt.IsZero() {
		event.CreatedAt = event.UpdatedAt
	}

	row := r.pool.QueryRow(ctx, `
		INSERT INTO events (
			id, title, description, source_url, image_url, location_id, slug, event_type, venue, city,
			starts_at, ends_at, scraped_at, status, review_notes,
			reviewed_by, reviewed_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			image_url = EXCLUDED.image_url,
			location_id = EXCLUDED.location_id,
			slug = EXCLUDED.slug,
			event_type = EXCLUDED.event_type,
			venue = EXCLUDED.venue,
			city = EXCLUDED.city,
			starts_at = EXCLUDED.starts_at,
			ends_at = EXCLUDED.ends_at,
			scraped_at = EXCLUDED.scraped_at,
			status = EXCLUDED.status,
			review_notes = EXCLUDED.review_notes,
			reviewed_by = EXCLUDED.reviewed_by,
			reviewed_at = EXCLUDED.reviewed_at,
			updated_at = EXCLUDED.updated_at
		RETURNING `+eventColumns,
		event.ID, event.Title, event.Description, event.SourceURL, event.ImageURL, event.LocationID,
		event.Slug, event.EventType, event.Venue, event.City,
		event.StartsAt, event.EndsAt, event.ScrapedAt, string(event.Status), event.ReviewNotes,
		event.ReviewedBy, event.ReviewedAt, event.CreatedAt, event.UpdatedAt,
	)

	saved, err := scanEvent(row)
	if err != nil {
		return nil, fmt.Errorf("save event: %w", err)
	}
	return r.attachLocation(ctx, saved)
}

func (r *EventRepository) LastScrapedAt(ctx context.Context) (time.Time, error) {
	var scrapedAt *time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT MAX(scraped_at) FROM events
	`).Scan(&scrapedAt)
	if err != nil {
		return time.Time{}, fmt.Errorf("last scraped at: %w", err)
	}
	if scrapedAt == nil {
		return time.Time{}, nil
	}
	return *scrapedAt, nil
}

func (r *EventRepository) GetBySourceURL(ctx context.Context, sourceURL string) (*events.Event, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+eventSelectColumns+`, `+eventLocationJoinColumns+`
		FROM events e
		LEFT JOIN event_locations l ON l.id = e.location_id
		WHERE e.source_url = $1
	`, sourceURL)

	event, err := scanEventWithLocation(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get event by source url: %w", err)
	}
	return &event, nil
}

func (r *EventRepository) queryEvents(ctx context.Context, query string, args ...any) ([]events.Event, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	var result []events.Event
	for rows.Next() {
		event, err := scanEventWithLocation(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	if result == nil {
		result = []events.Event{}
	}
	return result, nil
}

func (r *EventRepository) attachLocation(ctx context.Context, event events.Event) (*events.Event, error) {
	if event.LocationID == nil {
		return &event, nil
	}
	locRepo := NewEventLocationRepository(r.pool)
	loc, err := locRepo.GetByID(ctx, *event.LocationID)
	if err != nil {
		return nil, err
	}
	event.Location = loc
	return &event, nil
}

func scanEvent(row scannable) (events.Event, error) {
	var event events.Event
	var status string
	err := row.Scan(
		&event.ID, &event.Title, &event.Description, &event.SourceURL, &event.ImageURL, &event.LocationID,
		&event.Slug, &event.EventType, &event.Venue, &event.City,
		&event.StartsAt, &event.EndsAt, &event.ScrapedAt, &status, &event.ReviewNotes,
		&event.ReviewedBy, &event.ReviewedAt, &event.CreatedAt, &event.UpdatedAt,
	)
	if err != nil {
		return events.Event{}, err
	}
	event.Status = events.EventStatus(status)
	return event, nil
}

func scanEventWithLocation(row scannable) (events.Event, error) {
	var event events.Event
	var status string
	var locID *uuid.UUID
	var locName *string
	var locPinCoords []float64
	var locCreatedAt, locUpdatedAt *time.Time

	err := row.Scan(
		&event.ID, &event.Title, &event.Description, &event.SourceURL, &event.ImageURL, &event.LocationID,
		&event.Slug, &event.EventType, &event.Venue, &event.City,
		&event.StartsAt, &event.EndsAt, &event.ScrapedAt, &status, &event.ReviewNotes,
		&event.ReviewedBy, &event.ReviewedAt, &event.CreatedAt, &event.UpdatedAt,
		&locID, &locName, &locPinCoords, &locCreatedAt, &locUpdatedAt,
	)
	if err != nil {
		return events.Event{}, err
	}
	event.Status = events.EventStatus(status)

	if locID != nil && locName != nil && locCreatedAt != nil && locUpdatedAt != nil {
		event.Location = &events.Location{
			ID:        *locID,
			Name:      *locName,
			PinCoords: locPinCoords,
			CreatedAt: *locCreatedAt,
			UpdatedAt: *locUpdatedAt,
		}
	}
	return event, nil
}
