package events

import "time"

// EventWrite is the mutable content for admin create/update.
type EventWrite struct {
	Title       string
	Description string
	SourceURL   string
	ImageURL    *string
	EventType   string
	Venue       string
	City        string
	Slug        string
	StartsAt    time.Time
	EndsAt      *time.Time
	Status      EventStatus
}
