package outbound

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/events"
)

type EventRepository interface {
	UpsertBySourceURL(ctx context.Context, event events.Event) (*events.Event, error)
	List(ctx context.Context, filter events.ListFilter) ([]events.Event, error)
	Search(ctx context.Context, query string, limit int) ([]events.Event, error)
	GetByID(ctx context.Context, id uuid.UUID) (*events.Event, error)
	GetBySlug(ctx context.Context, slug string) (*events.Event, error)
	Save(ctx context.Context, event events.Event) (*events.Event, error)
	Delete(ctx context.Context, id uuid.UUID) error
	LastScrapedAt(ctx context.Context) (time.Time, error)
}
