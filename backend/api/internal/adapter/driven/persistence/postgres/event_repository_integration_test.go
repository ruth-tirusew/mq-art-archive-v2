//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	"github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

func TestEventLocationRepository_integration(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewEventLocationRepository(pool)
	ctx := context.Background()

	loc, err := repo.FindOrCreate(ctx, events.Location{
		Name:      "National Museum of Ethiopia",
		PinCoords: []float64{9.0108, 38.7613},
	})
	assist.NoError(t, err)
	assist.Equal(t, "National Museum of Ethiopia", loc.Name)
	assist.Len(t, 2, len(loc.PinCoords))
	assist.Equal(t, 9.0108, loc.PinCoords[0])
	assist.Equal(t, 38.7613, loc.PinCoords[1])

	same, err := repo.FindOrCreate(ctx, events.Location{Name: "  national museum of ethiopia  "})
	assist.NoError(t, err)
	assist.Equal(t, loc.ID, same.ID)

	withPin, err := repo.FindOrCreate(ctx, events.Location{
		Name:      "National Museum of Ethiopia",
		PinCoords: []float64{9.0108, 38.7613},
	})
	assist.NoError(t, err)
	assist.Equal(t, loc.ID, withPin.ID)
	if !withPin.HasPin() {
		t.Fatal("expected location to have pin coords")
	}

	got, err := repo.GetByID(ctx, loc.ID)
	assist.NoError(t, err)
	assist.Equal(t, loc.Name, got.Name)

	_, err = repo.GetByID(ctx, uuid.New())
	assist.ErrorIs(t, err, postgres.ErrNotFound)
}

func TestEventRepository_integration(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	locRepo := postgres.NewEventLocationRepository(pool)
	eventRepo := postgres.NewEventRepository(pool)
	ctx := context.Background()

	venue, err := locRepo.FindOrCreate(ctx, events.Location{
		Name:      "Addis Fine Art",
		PinCoords: []float64{9.0320, 38.7617},
	})
	assist.NoError(t, err)

	imageURL := "https://example.com/event-cover.jpg"
	startsAt := time.Now().UTC().Add(48 * time.Hour)
	scrapedAt := time.Now().UTC()

	created, err := eventRepo.UpsertBySourceURL(ctx, events.Event{
		Title:       "Opening Night",
		Description: "Group exhibition opening",
		SourceURL:   "https://example.com/events/opening-night",
		ImageURL:    &imageURL,
		LocationID:  &venue.ID,
		StartsAt:    startsAt,
		ScrapedAt:   scrapedAt,
		Status:      events.EventStatusPending,
	})
	assist.NoError(t, err)
	assist.Equal(t, events.EventStatusPending, created.Status)
	assist.NotNil(t, created.ImageURL)
	assist.Equal(t, imageURL, *created.ImageURL)
	assist.NotNil(t, created.Location)
	assist.Equal(t, "Addis Fine Art", created.Location.Name)

	updated, err := eventRepo.UpsertBySourceURL(ctx, events.Event{
		Title:       "Opening Night (revised)",
		Description: "Updated description",
		SourceURL:   "https://example.com/events/opening-night",
		StartsAt:    startsAt,
		ScrapedAt:   scrapedAt.Add(time.Hour),
	})
	assist.NoError(t, err)
	assist.Equal(t, "Opening Night (revised)", updated.Title)

	updated.Status = events.EventStatusApproved
	saved, err := eventRepo.Save(ctx, *updated)
	assist.NoError(t, err)
	assist.Equal(t, events.EventStatusApproved, saved.Status)

	skipped, err := eventRepo.UpsertBySourceURL(ctx, events.Event{
		Title:     "Should not overwrite",
		SourceURL: "https://example.com/events/opening-night",
		StartsAt:  startsAt,
		ScrapedAt: scrapedAt.Add(2 * time.Hour),
	})
	assist.NoError(t, err)
	assist.Equal(t, "Opening Night (revised)", skipped.Title)
	assist.Equal(t, events.EventStatusApproved, skipped.Status)

	list, err := eventRepo.List(ctx, events.PublicUpcomingFilter())
	assist.NoError(t, err)
	assist.GreaterOrEqual(t, len(list), 1)
	found := false
	for _, e := range list {
		if e.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected created event in upcoming list")
	}

	pending, err := eventRepo.List(ctx, events.PendingFilter())
	assist.NoError(t, err)
	assist.Len(t, 0, len(pending))

	last, err := eventRepo.LastScrapedAt(ctx)
	assist.NoError(t, err)
	if last.IsZero() {
		t.Fatal("expected non-zero last scraped at")
	}

	got, err := eventRepo.GetByID(ctx, created.ID)
	assist.NoError(t, err)
	assist.Equal(t, created.ID, got.ID)
	assist.NotNil(t, got.Location)
	if !got.Location.HasPin() {
		t.Fatal("expected location pin coords")
	}

	_, err = eventRepo.GetByID(ctx, uuid.New())
	assist.ErrorIs(t, err, postgres.ErrNotFound)
}

func TestEventRepository_integration_noLocationOrImage(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewEventRepository(pool)
	ctx := context.Background()

	created, err := repo.UpsertBySourceURL(ctx, events.Event{
		Title:     "TBA Venue",
		SourceURL: "https://example.com/events/tba",
		StartsAt:  time.Now().UTC().Add(24 * time.Hour),
		ScrapedAt: time.Now().UTC(),
	})
	assist.NoError(t, err)
	if created.ImageURL != nil {
		t.Fatal("expected nil image url")
	}
	if created.LocationID != nil {
		t.Fatal("expected nil location id")
	}
	if created.Location != nil {
		t.Fatal("expected nil location")
	}
}

func TestEventRepository_Search_integration(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewEventRepository(pool)
	ctx := context.Background()

	event, err := repo.UpsertBySourceURL(ctx, events.Event{
		Title:       "Contemporary Ethiopian Ceramics Fair",
		Description: "Annual showcase of ceramic artists from Addis Ababa",
		SourceURL:   "https://example.com/events/ceramics-fair",
		StartsAt:    time.Now().UTC().Add(72 * time.Hour),
		ScrapedAt:   time.Now().UTC(),
		Status:      events.EventStatusApproved,
	})
	assist.NoError(t, err)

	results, err := repo.Search(ctx, "ceramics Addis", 10)
	assist.NoError(t, err)
	assist.GreaterOrEqual(t, len(results), 1)
	assist.Equal(t, event.ID, results[0].ID)

	empty, err := repo.Search(ctx, "   ", 10)
	assist.NoError(t, err)
	assist.Len(t, 0, len(empty))
}
