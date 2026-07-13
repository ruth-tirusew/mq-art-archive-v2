package events

import (
	"time"

	"github.com/google/uuid"
)

type EventStatus string

const (
	EventStatusPending  EventStatus = "pending"
	EventStatusApproved EventStatus = "approved"
	EventStatusRejected EventStatus = "rejected"
)

type Event struct {
	ID          uuid.UUID
	Slug        string
	Title       string
	Description string
	SourceURL   string
	ImageURL    *string
	EventType   string
	Venue       string
	City        string
	LocationID  *uuid.UUID
	Location    *Location
	StartsAt    time.Time
	EndsAt      *time.Time
	ScrapedAt   time.Time
	Status      EventStatus
	ReviewNotes string
	ReviewedBy  *uuid.UUID
	ReviewedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
