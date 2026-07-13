package events

import (
	"time"

	"github.com/google/uuid"
)

// Location is a normalized venue. PinCoords holds [latitude, longitude] when a map pin is known.
type Location struct {
	ID        uuid.UUID
	Name      string
	PinCoords []float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// HasPin reports whether geographic coordinates are stored.
func (l Location) HasPin() bool {
	return len(l.PinCoords) == 2
}
