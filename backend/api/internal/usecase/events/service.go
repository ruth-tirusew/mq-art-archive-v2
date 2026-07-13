package events

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

type Service struct {
	events    outbound.EventRepository
	locations outbound.EventLocationRepository
	source    outbound.EventSource
}

func NewService(
	events outbound.EventRepository,
	locations outbound.EventLocationRepository,
	source outbound.EventSource,
) inbound.EventsService {
	return &Service{
		events:    events,
		locations: locations,
		source:    source,
	}
}

func (s *Service) List(ctx context.Context, filter domain.ListFilter) ([]domain.Event, error) {
	return s.events.List(ctx, filter)
}

func (s *Service) Search(ctx context.Context, query string, limit int) ([]domain.Event, error) {
	return s.events.Search(ctx, query, limit)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	return s.events.GetByID(ctx, id)
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*domain.Event, error) {
	return s.events.GetBySlug(ctx, slug)
}

func (s *Service) Review(
	ctx context.Context,
	id uuid.UUID,
	reviewerID uuid.UUID,
	status domain.EventStatus,
	notes string,
) (*domain.Event, error) {
	event, err := s.events.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	event.Status = status
	event.ReviewNotes = notes
	event.ReviewedBy = &reviewerID
	event.ReviewedAt = &now
	event.UpdatedAt = now

	return s.events.Save(ctx, *event)
}

func (s *Service) Sync(ctx context.Context) (int, error) {
	since, err := s.events.LastScrapedAt(ctx)
	if err != nil {
		return 0, fmt.Errorf("sync events: %w", err)
	}

	fetched, err := s.source.FetchEvents(ctx, since)
	if err != nil {
		return 0, fmt.Errorf("sync events: %w", err)
	}

	var upserted int
	for _, event := range fetched {
		prepared, err := s.prepareEvent(ctx, event)
		if err != nil {
			return upserted, err
		}
		if _, err := s.events.UpsertBySourceURL(ctx, prepared); err != nil {
			return upserted, fmt.Errorf("sync events: %w", err)
		}
		upserted++
	}
	return upserted, nil
}

func (s *Service) prepareEvent(ctx context.Context, event domain.Event) (domain.Event, error) {
	if event.Location == nil || event.Location.Name == "" {
		return event, nil
	}

	loc, err := s.locations.FindOrCreate(ctx, *event.Location)
	if err != nil {
		return domain.Event{}, err
	}
	event.LocationID = &loc.ID
	event.Location = loc
	return event, nil
}
