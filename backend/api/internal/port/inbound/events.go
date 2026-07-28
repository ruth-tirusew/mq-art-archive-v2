package inbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/events"
)

type EventsService interface {
	List(ctx context.Context, filter events.ListFilter) ([]events.Event, error)
	ListPending(ctx context.Context) ([]events.Event, error)
	Search(ctx context.Context, query string, limit int) ([]events.Event, error)
	GetByID(ctx context.Context, id uuid.UUID) (*events.Event, error)
	GetBySlug(ctx context.Context, slug string) (*events.Event, error)
	AdminCreate(ctx context.Context, write events.EventWrite) (*events.Event, error)
	Submit(ctx context.Context, write events.EventWrite) (*events.Event, error)
	AdminUpdateContent(ctx context.Context, id uuid.UUID, write events.EventWrite) (*events.Event, error)
	AdminDelete(ctx context.Context, id uuid.UUID) error
	Review(ctx context.Context, id uuid.UUID, reviewerID uuid.UUID, status events.EventStatus, notes string) (*events.Event, error)
	Sync(ctx context.Context) (int, error)
}
