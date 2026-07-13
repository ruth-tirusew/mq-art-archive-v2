package dto

import (
	"time"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/events"
)

type EventLocationResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	PinCoords []float64 `json:"pin_coords,omitempty"`
}

type EventResponse struct {
	ID          uuid.UUID              `json:"id"`
	Slug        string                 `json:"slug"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	SourceURL   string                 `json:"source_url,omitempty"`
	ImageURL    *string                `json:"image_url,omitempty"`
	EventType   string                 `json:"event_type"`
	Venue       string                 `json:"venue,omitempty"`
	City        string                 `json:"city,omitempty"`
	Location    *EventLocationResponse `json:"location,omitempty"`
	StartsAt    time.Time              `json:"starts_at"`
	EndsAt      *time.Time             `json:"ends_at,omitempty"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

func ToEventResponse(e domain.Event) EventResponse {
	resp := EventResponse{
		ID:          e.ID,
		Slug:        e.Slug,
		Title:       e.Title,
		Description: e.Description,
		SourceURL:   e.SourceURL,
		ImageURL:    e.ImageURL,
		EventType:   e.EventType,
		Venue:       e.Venue,
		City:        e.City,
		StartsAt:    e.StartsAt,
		EndsAt:      e.EndsAt,
		Status:      string(e.Status),
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
	if e.Location != nil {
		resp.Location = &EventLocationResponse{
			ID:        e.Location.ID,
			Name:      e.Location.Name,
			PinCoords: e.Location.PinCoords,
		}
	}
	return resp
}

func ToEventResponses(events []domain.Event) []EventResponse {
	out := make([]EventResponse, len(events))
	for i, e := range events {
		out[i] = ToEventResponse(e)
	}
	return out
}
