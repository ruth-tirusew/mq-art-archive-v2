package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/events"
)

type EventLocationRepository interface {
	FindOrCreate(ctx context.Context, loc events.Location) (*events.Location, error)
	GetByID(ctx context.Context, id uuid.UUID) (*events.Location, error)
}
