package events_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/events"
	eventsuc "github.com/mq/api/internal/usecase/events"
	"github.com/mq/api/internal/testutil/assist"
)

type mockEventRepo struct {
	lastScrapedAt func(ctx context.Context) (time.Time, error)
	upsert        func(ctx context.Context, event events.Event) (*events.Event, error)
}

func (m *mockEventRepo) UpsertBySourceURL(ctx context.Context, event events.Event) (*events.Event, error) {
	return m.upsert(ctx, event)
}

func (m *mockEventRepo) List(ctx context.Context, filter events.ListFilter) ([]events.Event, error) {
	return nil, nil
}

func (m *mockEventRepo) Search(ctx context.Context, query string, limit int) ([]events.Event, error) {
	return nil, nil
}

func (m *mockEventRepo) GetByID(ctx context.Context, id uuid.UUID) (*events.Event, error) {
	return nil, nil
}

func (m *mockEventRepo) GetBySlug(ctx context.Context, slug string) (*events.Event, error) {
	return nil, nil
}

func (m *mockEventRepo) Save(ctx context.Context, event events.Event) (*events.Event, error) {
	return nil, nil
}

func (m *mockEventRepo) LastScrapedAt(ctx context.Context) (time.Time, error) {
	return m.lastScrapedAt(ctx)
}

type mockLocationRepo struct {
	findOrCreate func(ctx context.Context, loc events.Location) (*events.Location, error)
}

func (m *mockLocationRepo) FindOrCreate(ctx context.Context, loc events.Location) (*events.Location, error) {
	return m.findOrCreate(ctx, loc)
}

func (m *mockLocationRepo) GetByID(ctx context.Context, id uuid.UUID) (*events.Location, error) {
	return nil, nil
}

type mockEventSource struct {
	fetch func(ctx context.Context, since time.Time) ([]events.Event, error)
}

func (m *mockEventSource) FetchEvents(ctx context.Context, since time.Time) ([]events.Event, error) {
	return m.fetch(ctx, since)
}

func TestService_Sync_noopSource(t *testing.T) {
	since := time.Now().UTC().Add(-24 * time.Hour)
	repo := &mockEventRepo{
		lastScrapedAt: func(ctx context.Context) (time.Time, error) {
			return since, nil
		},
	}
	source := &mockEventSource{
		fetch: func(ctx context.Context, gotSince time.Time) ([]events.Event, error) {
			assist.Equal(t, since, gotSince)
			return []events.Event{}, nil
		},
	}

	svc := eventsuc.NewService(repo, &mockLocationRepo{}, source)
	count, err := svc.Sync(context.Background())
	assist.NoError(t, err)
	assist.Equal(t, 0, count)
}

func TestService_Sync_upsertsWithLocation(t *testing.T) {
	locID := uuid.New()
	repo := &mockEventRepo{
		lastScrapedAt: func(ctx context.Context) (time.Time, error) {
			return time.Time{}, nil
		},
		upsert: func(ctx context.Context, event events.Event) (*events.Event, error) {
			assist.NotNil(t, event.LocationID)
			assist.Equal(t, locID, *event.LocationID)
			return &event, nil
		},
	}
	locations := &mockLocationRepo{
		findOrCreate: func(ctx context.Context, loc events.Location) (*events.Location, error) {
			assist.Equal(t, "Gallery One", loc.Name)
			return &events.Location{ID: locID, Name: loc.Name}, nil
		},
	}
	source := &mockEventSource{
		fetch: func(ctx context.Context, since time.Time) ([]events.Event, error) {
			return []events.Event{{
				Title:     "Show",
				SourceURL: "https://example.com/1",
				Location:  &events.Location{Name: "Gallery One"},
				StartsAt:  time.Now().UTC(),
				ScrapedAt: time.Now().UTC(),
			}}, nil
		},
	}

	svc := eventsuc.NewService(repo, locations, source)
	count, err := svc.Sync(context.Background())
	assist.NoError(t, err)
	assist.Equal(t, 1, count)
}
