//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/domain/institution"
	"github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/domain/profile"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

func TestContentService_integration(t *testing.T) {
	app := integration.NewApp(t)
	ctx := context.Background()
	authorID := uuid.New()

	draft, err := app.Content.CreateDraft(ctx, authorID, "Service Draft", "body text")
	assist.NoError(t, err)
	assist.Equal(t, content.ArticleStatusDraft, draft.Status)
	assist.Equal(t, "service-draft", draft.Slug)

	_, err = app.Content.GetBySlug(ctx, "service-draft")
	assist.ErrorIs(t, err, apperrors.ErrNotFound)

	_, err = app.Content.CreateDraft(ctx, authorID, "  ", "body")
	assist.Error(t, err)

	published, err := app.Content.CreateDraft(ctx, authorID, "Service Published", "published body")
	assist.NoError(t, err)
	published.Status = content.ArticleStatusPublished
	published.UpdatedAt = time.Now().UTC()
	_, err = app.Pool.Exec(ctx, `
		UPDATE articles SET status = $1, updated_at = $2 WHERE id = $3
	`, string(content.ArticleStatusPublished), published.UpdatedAt, published.ID)
	assist.NoError(t, err)

	got, err := app.Content.GetBySlug(ctx, "service-published")
	assist.NoError(t, err)
	assist.Equal(t, "Service Published", got.Title)

	list, err := app.Content.ListPublished(ctx, content.PublicListFilter())
	assist.NoError(t, err)
	assist.GreaterOrEqual(t, len(list), 2) // seed article + our published draft
}

func TestProfileService_integration(t *testing.T) {
	app := integration.NewApp(t)
	ctx := context.Background()

	userID := uuid.New()
	integration.InsertUser(t, app.Pool, userID, identity.RoleArtist)

	now := time.Now().UTC()
	created, err := app.Profile.UpdateArtist(ctx, profile.ArtistProfile{
		ID:          uuid.New(),
		UserID:      userID,
		Slug:        "service-artist",
		DisplayName: "Service Artist",
		Bio:         "Bio",
		Status:      profile.ProfileStatusApproved,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	assist.NoError(t, err)

	got, err := app.Profile.GetArtistBySlug(ctx, "service-artist")
	assist.NoError(t, err)
	assist.Equal(t, "Service Artist", got.DisplayName)

	created.DisplayName = "Updated Artist"
	updated, err := app.Profile.UpdateArtist(ctx, *created)
	assist.NoError(t, err)
	assist.Equal(t, "Updated Artist", updated.DisplayName)

	byID, err := app.Profile.GetArtistByID(ctx, created.ID)
	assist.NoError(t, err)
	assist.Equal(t, "Updated Artist", byID.DisplayName)
}

func TestArtService_integration(t *testing.T) {
	app := integration.NewApp(t)
	ctx := context.Background()

	artistID, _ := integration.InsertArtistProfile(t, app.Pool, "art-service-artist", "Art Service Artist")

	created, err := app.Art.CreateDraft(ctx, artistID, "Studio Piece", "A description", "oil")
	assist.NoError(t, err)
	assist.Equal(t, art.ArtStatusDraft, created.Status)

	_, err = app.Art.CreateDraft(ctx, artistID, "", "desc", "oil")
	assist.Error(t, err)

	got, err := app.Art.GetByID(ctx, created.ID)
	assist.NoError(t, err)
	assist.Equal(t, created.ID, got.ID)

	publishedAt := time.Now().UTC()
	_, err = app.Pool.Exec(ctx, `
		UPDATE art_posts SET status = $1, published_at = $2, updated_at = $3 WHERE id = $4
	`, string(art.ArtStatusPublished), publishedAt, publishedAt, created.ID)
	assist.NoError(t, err)

	list, err := app.Art.ListByArtist(ctx, artistID)
	assist.NoError(t, err)
	assist.Len(t, 1, len(list))
	assist.Equal(t, "Studio Piece", list[0].Title)
}

func TestIdentityService_integration(t *testing.T) {
	app := integration.NewApp(t)
	ctx := context.Background()

	userID := uuid.New()
	integration.InsertUser(t, app.Pool, userID, identity.RoleArtist)

	got, err := app.Identity.GetUser(ctx, userID)
	assist.NoError(t, err)
	assist.Equal(t, userID, got.ID)
	assist.Equal(t, identity.RoleArtist, got.Role)

	_, err = app.Identity.GetUser(ctx, uuid.New())
	assist.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestOnboardingService_integration(t *testing.T) {
	app := integration.NewApp(t)
	ctx := context.Background()

	now := time.Now().UTC()
	appID := uuid.New()
	_, err := app.Pool.Exec(ctx, `
		INSERT INTO onboarding_applications (
			id, applicant_id, applicant_type, display_name, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, appID, uuid.New(), string(onboarding.ApplicantTypeArtist), "Pending Studio",
		string(onboarding.ApprovalStatusPending), now, now)
	assist.NoError(t, err)

	pending, err := app.Onboarding.ListPending(ctx)
	assist.NoError(t, err)
	assist.GreaterOrEqual(t, len(pending), 1)

	reviewerID := uuid.New()
	reviewed, err := app.Onboarding.Review(ctx, appID, reviewerID, onboarding.ApprovalStatusApproved, "looks good")
	assist.NoError(t, err)
	assist.Equal(t, onboarding.ApprovalStatusApproved, reviewed.Status)
	assist.Equal(t, "looks good", reviewed.Notes)
	assist.NotNil(t, reviewed.ReviewedBy)
	assist.Equal(t, reviewerID, *reviewed.ReviewedBy)

	pending, err = app.Onboarding.ListPending(ctx)
	assist.NoError(t, err)
	for _, p := range pending {
		assist.NotEqual(t, appID, p.ID)
	}
}

func TestInstitutionService_integration(t *testing.T) {
	app := integration.NewApp(t)
	ctx := context.Background()

	userID := uuid.New()
	integration.InsertUser(t, app.Pool, userID, identity.RoleInstitution)

	now := time.Now().UTC()
	created, err := app.Institution.Update(ctx, institution.Institution{
		ID:          uuid.New(),
		UserID:      userID,
		Slug:        "service-gallery",
		Name:        "Service Gallery",
		Description: "A gallery",
		Status:      institution.StatusApproved,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	assist.NoError(t, err)

	got, err := app.Institution.GetBySlug(ctx, "service-gallery")
	assist.NoError(t, err)
	assist.Equal(t, "Service Gallery", got.Name)

	created.Name = "Updated Gallery"
	updated, err := app.Institution.Update(ctx, *created)
	assist.NoError(t, err)
	assist.Equal(t, "Updated Gallery", updated.Name)

	byID, err := app.Institution.GetByID(ctx, created.ID)
	assist.NoError(t, err)
	assist.Equal(t, "Updated Gallery", byID.Name)
}

func TestEventsService_integration(t *testing.T) {
	app := integration.NewApp(t)
	ctx := context.Background()

	count, err := app.Events.Sync(ctx)
	assist.NoError(t, err)
	assist.Equal(t, 0, count)

	list, err := app.Events.List(ctx, events.PublicUpcomingFilter())
	assist.NoError(t, err)
	assist.Len(t, 3, len(list))
}
