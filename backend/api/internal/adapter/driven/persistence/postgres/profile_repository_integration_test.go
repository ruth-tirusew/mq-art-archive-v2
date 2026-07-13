//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/domain/profile"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

func insertUser(t *testing.T, pool *postgres.Pool, id uuid.UUID) {
	t.Helper()
	now := time.Now().UTC()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (id, email, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, id, id.String()+"@example.com", string(identity.RoleArtist), now, now)
	assist.NoError(t, err)
}

func TestProfileRepository_integration(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewProfileRepository(pool)
	ctx := context.Background()

	userID := uuid.New()
	insertUser(t, pool, userID)

	now := time.Now().UTC()
	created, err := repo.SaveArtist(ctx, profile.ArtistProfile{
		ID:          uuid.New(),
		UserID:      userID,
		Slug:        "abebe-art",
		DisplayName: "Abebe Art",
		Bio:         "Painter from Addis",
		Contact: profile.ContactInfo{
			Email:    "abebe@example.com",
			Location: "Addis Ababa",
		},
		Social: profile.SocialLinks{
			Instagram: "@abebe",
		},
		Status:    profile.ProfileStatusApproved,
		CreatedAt: now,
		UpdatedAt: now,
	})
	assist.NoError(t, err)
	assist.Equal(t, "abebe-art", created.Slug)

	bySlug, err := repo.GetArtistBySlug(ctx, "abebe-art")
	assist.NoError(t, err)
	assist.Equal(t, "Abebe Art", bySlug.DisplayName)
	assist.Equal(t, "Addis Ababa", bySlug.Contact.Location)
	assist.Equal(t, "@abebe", bySlug.Social.Instagram)

	byID, err := repo.GetArtistByID(ctx, created.ID)
	assist.NoError(t, err)
	assist.Equal(t, created.ID, byID.ID)

	created.DisplayName = "Abebe Updated"
	updated, err := repo.SaveArtist(ctx, *created)
	assist.NoError(t, err)
	assist.Equal(t, "Abebe Updated", updated.DisplayName)

	_, err = repo.GetArtistBySlug(ctx, "missing-slug")
	assist.ErrorIs(t, err, postgres.ErrNotFound)
}
