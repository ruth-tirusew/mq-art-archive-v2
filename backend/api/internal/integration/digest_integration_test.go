//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

func TestDigestService_BuildWeekly_integration(t *testing.T) {
	app := integration.NewApp(t)
	ctx := context.Background()
	now := time.Now().UTC()

	// First call, no prior digest_runs row: the window should fall back to a 7-day
	// lookback ending now, not to some other default.
	first, err := app.Digest.BuildWeekly(ctx)
	assist.NoError(t, err)
	assist.WithinDuration(t, now.Add(-7*24*time.Hour), first.PeriodStart, time.Minute)
	assist.WithinDuration(t, now, first.PeriodEnd, time.Minute)

	authorID := uuid.New()
	integration.InsertUser(t, app.Pool, authorID, "artist")

	// An article published inside the window: should appear.
	recent := content.ArticleStatusPublished
	recentArticle, err := app.Content.AdminCreate(ctx, authorID, content.ArticleWrite{
		Title: "Recent Piece", Body: "body", Status: &recent,
	})
	assist.NoError(t, err)

	// An article published well before the window: should not appear, even though
	// AdminCreate just set PublishedAt to now — push it into the past directly.
	staleArticle, err := app.Content.AdminCreate(ctx, authorID, content.ArticleWrite{
		Title: "Stale Piece", Body: "body", Status: &recent,
	})
	assist.NoError(t, err)
	stalePublishedAt := now.Add(-30 * 24 * time.Hour)
	_, err = app.Pool.Exec(ctx, `UPDATE articles SET published_at = $1 WHERE id = $2`, stalePublishedAt, staleArticle.ID)
	assist.NoError(t, err)

	// An art post published inside the window: should appear.
	artistID, _ := integration.InsertArtistProfile(t, app.Pool, "digest-artist", "Digest Artist")
	artPost, err := app.Art.CreateDraft(ctx, artistID, art.ArtPostWrite{Title: "New Work", Description: "d", Medium: "oil"})
	assist.NoError(t, err)
	artPublishedAt := now.Add(-2 * 24 * time.Hour)
	_, err = app.Pool.Exec(ctx, `
		UPDATE art_posts SET status = $1, published_at = $2, updated_at = $2 WHERE id = $3
	`, string(art.ArtStatusPublished), artPublishedAt, artPost.ID)
	assist.NoError(t, err)

	// An event starting within the next 7 days: should appear.
	soonEventID := uuid.New()
	insertEvent(t, app, soonEventID, "Soon Event", now.Add(3*24*time.Hour), "approved")

	// An event starting 20 days out: outside the upcoming window, should not appear.
	farEventID := uuid.New()
	insertEvent(t, app, farEventID, "Far Event", now.Add(20*24*time.Hour), "approved")

	// A pending (unapproved) event starting soon: should not appear either.
	pendingEventID := uuid.New()
	insertEvent(t, app, pendingEventID, "Pending Event", now.Add(1*24*time.Hour), "pending")

	digest, err := app.Digest.BuildWeekly(ctx)
	assist.NoError(t, err)

	assertArticleIDs(t, digest.Articles, recentArticle.ID, true)
	assertArticleIDs(t, digest.Articles, staleArticle.ID, false)
	assertArtPostIDs(t, digest.ArtPosts, artPost.ID, true)
	assertEventIDs(t, digest.Events, soonEventID, true)
	assertEventIDs(t, digest.Events, farEventID, false)
	assertEventIDs(t, digest.Events, pendingEventID, false)

	// Completing a run should advance the next window's start to this run's period end,
	// not fall back to the 7-day default again.
	run, err := app.Digest.StartRun(ctx, *digest)
	assist.NoError(t, err)
	assist.NoError(t, app.Digest.CompleteRun(ctx, run.ID))

	second, err := app.Digest.BuildWeekly(ctx)
	assist.NoError(t, err)
	assist.WithinDuration(t, digest.PeriodEnd, second.PeriodStart, time.Second)
}

func insertEvent(t *testing.T, app *integration.App, id uuid.UUID, title string, startsAt time.Time, status string) {
	t.Helper()
	now := time.Now().UTC()
	_, err := app.Pool.Exec(context.Background(), `
		INSERT INTO events (id, title, description, source_url, slug, starts_at, scraped_at, status, created_at, updated_at)
		VALUES ($1, $2, '', $3, $4, $5, $6, $7, $6, $6)
	`, id, title, "https://example.com/"+id.String(), id.String(), startsAt, now, status)
	assist.NoError(t, err)
}

func assertArticleIDs(t *testing.T, articles []content.Article, id uuid.UUID, wantPresent bool) {
	t.Helper()
	found := false
	for _, a := range articles {
		if a.ID == id {
			found = true
		}
	}
	if found != wantPresent {
		t.Fatalf("article %s presence = %v, want %v", id, found, wantPresent)
	}
}

func assertArtPostIDs(t *testing.T, posts []art.ArtPostWithArtist, id uuid.UUID, wantPresent bool) {
	t.Helper()
	found := false
	for _, p := range posts {
		if p.ID == id {
			found = true
		}
	}
	if found != wantPresent {
		t.Fatalf("art post %s presence = %v, want %v", id, found, wantPresent)
	}
}

func assertEventIDs(t *testing.T, evts []events.Event, id uuid.UUID, wantPresent bool) {
	t.Helper()
	found := false
	for _, e := range evts {
		if e.ID == id {
			found = true
		}
	}
	if found != wantPresent {
		t.Fatalf("event %s presence = %v, want %v", id, found, wantPresent)
	}
}
