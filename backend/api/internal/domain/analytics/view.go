package analytics

import (
	"time"

	"github.com/google/uuid"
)

type View struct {
	EntityType string    `json:"entity_type"`
	EntityID   uuid.UUID `json:"entity_id"`
	Day        time.Time `json:"day"`
	Count      int64     `json:"count"`
}
