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

type ReviewEventRequest struct {
	Status string `json:"status" binding:"required"`
	Notes  string `json:"notes"`
}

type AdminEventWriteRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	SourceURL   string     `json:"source_url"`
	ImageURL    *string    `json:"image_url"`
	EventType   string     `json:"event_type"`
	Venue       string     `json:"venue"`
	City        string     `json:"city"`
	Slug        string     `json:"slug"`
	StartsAt    time.Time  `json:"starts_at" binding:"required"`
	EndsAt      *time.Time `json:"ends_at"`
	Status      string     `json:"status"`
}

type EventSubmissionRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	SourceURL   string     `json:"source_url"`
	EventType   string     `json:"event_type"`
	Venue       string     `json:"venue"`
	City        string     `json:"city"`
	StartsAt    time.Time  `json:"starts_at" binding:"required"`
	EndsAt      *time.Time `json:"ends_at"`
}

func (r EventSubmissionRequest) ToWrite() domain.EventWrite {
	return domain.EventWrite{
		Title:       r.Title,
		Description: r.Description,
		SourceURL:   r.SourceURL,
		EventType:   r.EventType,
		Venue:       r.Venue,
		City:        r.City,
		StartsAt:    r.StartsAt,
		EndsAt:      r.EndsAt,
		Status:      domain.EventStatusPending,
	}
}

func (r AdminEventWriteRequest) ToWrite() domain.EventWrite {
	return domain.EventWrite{
		Title:       r.Title,
		Description: r.Description,
		SourceURL:   r.SourceURL,
		ImageURL:    r.ImageURL,
		EventType:   r.EventType,
		Venue:       r.Venue,
		City:        r.City,
		Slug:        r.Slug,
		StartsAt:    r.StartsAt,
		EndsAt:      r.EndsAt,
		Status:      domain.EventStatus(r.Status),
	}
}

type SyncEventsResponse struct {
	Upserted int `json:"upserted"`
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
	ReviewNotes string                 `json:"review_notes,omitempty"`
	ReviewedBy  *uuid.UUID             `json:"reviewed_by,omitempty"`
	ReviewedAt  *time.Time             `json:"reviewed_at,omitempty"`
	ScrapedAt   *time.Time             `json:"scraped_at,omitempty"`
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
		ReviewNotes: e.ReviewNotes,
		ReviewedBy:  e.ReviewedBy,
		ReviewedAt:  e.ReviewedAt,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
	if !e.ScrapedAt.IsZero() {
		scrapedAt := e.ScrapedAt
		resp.ScrapedAt = &scrapedAt
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
