package events_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/testutil/assist"
	eventsuc "github.com/mq/api/internal/usecase/events"
)

type mockEventRepo struct {
	lastScrapedAt func(ctx context.Context) (time.Time, error)
	list          func(ctx context.Context, filter events.ListFilter) ([]events.Event, error)
	getByID       func(ctx context.Context, id uuid.UUID) (*events.Event, error)
	upsert        func(ctx context.Context, event events.Event) (*events.Event, error)
	save          func(ctx context.Context, event events.Event) (*events.Event, error)
}

func (m *mockEventRepo) UpsertBySourceURL(ctx context.Context, event events.Event) (*events.Event, error) {
	return m.upsert(ctx, event)
}

func (m *mockEventRepo) List(ctx context.Context, filter events.ListFilter) ([]events.Event, error) {
	if m.list != nil {
		return m.list(ctx, filter)
	}
	return nil, nil
}

func (m *mockEventRepo) Search(ctx context.Context, query string, limit int) ([]events.Event, error) {
	return nil, nil
}

func (m *mockEventRepo) GetByID(ctx context.Context, id uuid.UUID) (*events.Event, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return nil, nil
}

func (m *mockEventRepo) GetBySlug(ctx context.Context, slug string) (*events.Event, error) {
	return nil, nil
}

func (m *mockEventRepo) Save(ctx context.Context, event events.Event) (*events.Event, error) {
	if m.save != nil {
		return m.save(ctx, event)
	}
	return nil, nil
}

func (m *mockEventRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
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

type summaryRecipients struct{ emails []string }

func (r summaryRecipients) ListEventSummaryRecipients(context.Context) ([]string, error) {
	return r.emails, nil
}

type summaryMailer struct {
	sent int
	body string
}

func (m *summaryMailer) Send(_ context.Context, _, _, body string) error {
	m.sent++
	m.body = body
	return nil
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

func TestService_Sync_sendsSummaryToOptedInRecipients(t *testing.T) {
	repo := &mockEventRepo{
		lastScrapedAt: func(context.Context) (time.Time, error) { return time.Time{}, nil },
		upsert:        func(_ context.Context, event events.Event) (*events.Event, error) { return &event, nil },
	}
	source := &mockEventSource{fetch: func(context.Context, time.Time) ([]events.Event, error) {
		return []events.Event{{Title: "Show", SourceURL: "https://example.com/1"}}, nil
	}}
	mailer := &summaryMailer{}
	svc := eventsuc.NewService(repo, &mockLocationRepo{}, source, summaryRecipients{emails: []string{"one@example.com"}}, mailer)
	count, err := svc.Sync(context.Background())
	assist.NoError(t, err)
	assist.Equal(t, 1, count)
	assist.Equal(t, 1, mailer.sent)
}

func TestService_ListPending(t *testing.T) {
	repo := &mockEventRepo{
		list: func(ctx context.Context, filter events.ListFilter) ([]events.Event, error) {
			assist.NotNil(t, filter.Status)
			assist.Equal(t, events.EventStatusPending, *filter.Status)
			return []events.Event{{Title: "Pending Show"}}, nil
		},
	}
	svc := eventsuc.NewService(repo, &mockLocationRepo{}, &mockEventSource{})
	items, err := svc.ListPending(context.Background())
	assist.NoError(t, err)
	assist.Len(t, 1, len(items))
}

func TestService_Review(t *testing.T) {
	eventID := uuid.New()
	reviewerID := uuid.New()
	now := time.Now().UTC()
	repo := &mockEventRepo{
		getByID: func(ctx context.Context, id uuid.UUID) (*events.Event, error) {
			return &events.Event{ID: id, Title: "Show", Status: events.EventStatusPending, StartsAt: now}, nil
		},
		save: func(ctx context.Context, event events.Event) (*events.Event, error) {
			assist.Equal(t, events.EventStatusApproved, event.Status)
			assist.Equal(t, "looks good", event.ReviewNotes)
			return &event, nil
		},
	}
	svc := eventsuc.NewService(repo, &mockLocationRepo{}, &mockEventSource{})
	got, err := svc.Review(context.Background(), eventID, reviewerID, events.EventStatusApproved, "looks good")
	assist.NoError(t, err)
	assist.Equal(t, events.EventStatusApproved, got.Status)
}

func TestService_GetByID(t *testing.T) {
	eventID := uuid.New()
	repo := &mockEventRepo{
		getByID: func(ctx context.Context, id uuid.UUID) (*events.Event, error) {
			return &events.Event{ID: id, Title: "Show"}, nil
		},
	}
	svc := eventsuc.NewService(repo, &mockLocationRepo{}, &mockEventSource{})
	got, err := svc.GetByID(context.Background(), eventID)
	assist.NoError(t, err)
	assist.Equal(t, eventID, got.ID)
}
