package events

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	domain "github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

type Service struct {
	events     outbound.EventRepository
	locations  outbound.EventLocationRepository
	source     outbound.EventSource
	recipients outbound.EventNotificationRepository
	mailer     outbound.Mailer
}

func NewService(
	events outbound.EventRepository,
	locations outbound.EventLocationRepository,
	source outbound.EventSource,
	optional ...any,
) inbound.EventsService {
	service := &Service{
		events:    events,
		locations: locations,
		source:    source,
	}
	for _, dependency := range optional {
		switch value := dependency.(type) {
		case outbound.EventNotificationRepository:
			service.recipients = value
		case outbound.Mailer:
			service.mailer = value
		}
	}
	return service
}

func (s *Service) List(ctx context.Context, filter domain.ListFilter) ([]domain.Event, error) {
	return s.events.List(ctx, filter)
}

func (s *Service) ListPending(ctx context.Context) ([]domain.Event, error) {
	return s.events.List(ctx, domain.PendingFilter())
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

func (s *Service) AdminCreate(ctx context.Context, write domain.EventWrite) (*domain.Event, error) {
	return s.createEvent(ctx, write, "admin://events/")
}

func (s *Service) Submit(ctx context.Context, write domain.EventWrite) (*domain.Event, error) {
	write.Status = domain.EventStatusPending
	write.Slug = ""
	return s.createEvent(ctx, write, "public://events/")
}

func (s *Service) createEvent(ctx context.Context, write domain.EventWrite, sourcePrefix string) (*domain.Event, error) {
	now := time.Now().UTC()
	id := uuid.New()
	title := strings.TrimSpace(write.Title)
	if title == "" {
		return nil, apperrors.ErrValidation
	}
	if write.StartsAt.IsZero() {
		return nil, apperrors.ErrValidation
	}

	sourceURL := strings.TrimSpace(write.SourceURL)
	if sourceURL == "" {
		sourceURL = sourcePrefix + id.String()
	}
	slug := strings.TrimSpace(write.Slug)
	if slug == "" {
		slug = uniqueSlug(title, sourceURL)
	}
	status := write.Status
	if status == "" {
		status = domain.EventStatusPending
	}
	eventType := strings.TrimSpace(write.EventType)
	if eventType == "" {
		eventType = "Opening"
	}

	event := domain.Event{
		ID:          id,
		Title:       title,
		Description: write.Description,
		SourceURL:   sourceURL,
		ImageURL:    write.ImageURL,
		EventType:   eventType,
		Venue:       write.Venue,
		City:        write.City,
		Slug:        slug,
		StartsAt:    write.StartsAt.UTC(),
		EndsAt:      write.EndsAt,
		ScrapedAt:   now,
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.events.Save(ctx, event)
}

func (s *Service) AdminUpdateContent(ctx context.Context, id uuid.UUID, write domain.EventWrite) (*domain.Event, error) {
	event, err := s.events.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(write.Title)
	if title == "" {
		return nil, apperrors.ErrValidation
	}
	if write.StartsAt.IsZero() {
		return nil, apperrors.ErrValidation
	}

	event.Title = title
	event.Description = write.Description
	if strings.TrimSpace(write.SourceURL) != "" {
		event.SourceURL = strings.TrimSpace(write.SourceURL)
	}
	event.ImageURL = write.ImageURL
	if strings.TrimSpace(write.EventType) != "" {
		event.EventType = strings.TrimSpace(write.EventType)
	}
	event.Venue = write.Venue
	event.City = write.City
	if strings.TrimSpace(write.Slug) != "" {
		event.Slug = strings.TrimSpace(write.Slug)
	} else if event.Slug == "" {
		event.Slug = slugify(title)
	}
	event.StartsAt = write.StartsAt.UTC()
	event.EndsAt = write.EndsAt
	if write.Status != "" {
		event.Status = write.Status
	}
	event.UpdatedAt = time.Now().UTC()
	return s.events.Save(ctx, *event)
}

func (s *Service) AdminDelete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.events.GetByID(ctx, id); err != nil {
		return err
	}
	return s.events.Delete(ctx, id)
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
	s.sendSyncSummary(ctx, upserted)
	return upserted, nil
}

func (s *Service) sendSyncSummary(ctx context.Context, upserted int) {
	if s.recipients == nil || s.mailer == nil {
		return
	}
	emails, err := s.recipients.ListEventSummaryRecipients(ctx)
	if err != nil {
		log.Printf("event sync summary recipients: %v", err)
		return
	}
	body := "The event sync completed successfully. Upserted events: " + strconv.Itoa(upserted) + "."
	for _, email := range emails {
		if err := s.mailer.Send(ctx, email, "Artiv event sync summary", body); err != nil {
			log.Printf("event sync summary email to %s: %v", email, err)
		}
	}
}

func (s *Service) prepareEvent(ctx context.Context, event domain.Event) (domain.Event, error) {
	if event.Slug == "" {
		event.Slug = uniqueSlug(event.Title, event.SourceURL)
	}
	if event.Status == "" {
		event.Status = domain.EventStatusPending
	}
	if event.EventType == "" {
		event.EventType = "Opening"
	}
	if event.Location == nil || event.Location.Name == "" {
		return event, nil
	}

	loc, err := s.locations.FindOrCreate(ctx, *event.Location)
	if err != nil {
		return domain.Event{}, err
	}
	event.LocationID = &loc.ID
	event.Location = loc
	if event.Venue == "" {
		event.Venue = loc.Name
	}
	return event, nil
}
