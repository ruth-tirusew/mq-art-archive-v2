//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	"github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/domain/profile"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

func insertArtistProfile(t *testing.T, pool *postgres.Pool) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	userID := uuid.New()
	now := time.Now().UTC()

	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, email, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, userID.String()+"@example.com", string(identity.RoleArtist), now, now)
	assist.NoError(t, err)

	artistID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO artist_profiles (
			id, user_id, slug, display_name, bio, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, artistID, userID, artistID.String()+"-slug", "Test Artist", "", string(profile.ProfileStatusApproved), now, now)
	assist.NoError(t, err)

	return artistID
}

func TestArtPostRepository_integration(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewArtPostRepository(pool)
	ctx := context.Background()

	artistID := insertArtistProfile(t, pool)
	now := time.Now().UTC()
	publishedAt := now.Add(-time.Hour)

	created, err := repo.Create(ctx, art.ArtPost{
		ID:          uuid.New(),
		ArtistID:    artistID,
		Title:       "Blue Horizon",
		Description: "Oil on canvas",
		Medium:      "oil",
		Media: []art.MediaAsset{
			{
				ID:        uuid.New(),
				URL:       "https://cdn.example.com/blue.jpg",
				MimeType:  "image/jpeg",
				Width:     800,
				Height:    600,
				SortOrder: 0,
			},
		},
		Status:      art.ArtStatusPublished,
		PublishedAt: &publishedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	assist.NoError(t, err)
	assist.Len(t, 1, len(created.Media))

	got, err := repo.GetByID(ctx, created.ID)
	assist.NoError(t, err)
	assist.Equal(t, "Blue Horizon", got.Title)
	assist.Equal(t, "https://cdn.example.com/blue.jpg", got.Media[0].URL)

	list, err := repo.ListByArtist(ctx, artistID)
	assist.NoError(t, err)
	assist.Len(t, 1, len(list))

	_, err = repo.GetByID(ctx, uuid.New())
	assist.ErrorIs(t, err, postgres.ErrNotFound)
}
