//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

func uploadedAsset(t *testing.T, pool *postgres.Pool, owner uuid.UUID, n int, hash string) uuid.UUID {
	t.Helper()
	url := "https://res.cloudinary.com/cloud/image/upload/v1/mq/work-" + uuid.NewString() + ".webp"
	registerAsset(t, pool, owner, url, hash)
	var id uuid.UUID
	assist.NoError(t, pool.QueryRow(context.Background(), `SELECT id FROM media_assets WHERE secure_url = $1`, url).Scan(&id))
	return id
}

func TestApplicantWorkRepository_addListDelete(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewApplicantWorkRepository(pool)
	ctx := context.Background()
	user, other := uuid.New(), uuid.New()
	integration.InsertUser(t, pool, user, identity.RolePublic)
	integration.InsertUser(t, pool, other, identity.RolePublic)

	year := 2024
	first, err := repo.Add(ctx, onboarding.ApplicantWork{
		ID: uuid.New(), UserID: user, MediaAssetID: uploadedAsset(t, pool, user, 1, "HASH-A"),
		Title: "First", Medium: "oil", Year: &year, CreatedAt: time.Now().UTC(),
	})
	assist.NoError(t, err)
	assist.Equal(t, 0, first.Position)
	assist.Equal(t, "HASH-A", first.Blurhash) // the image's own hash comes along
	assist.Contains(t, first.URL, "res.cloudinary.com")

	second, err := repo.Add(ctx, onboarding.ApplicantWork{
		ID: uuid.New(), UserID: user, MediaAssetID: uploadedAsset(t, pool, user, 2, "HASH-B"), Title: "Second", CreatedAt: time.Now().UTC(),
	})
	assist.NoError(t, err)
	assist.Equal(t, 1, second.Position)

	got, err := repo.ListByUser(ctx, user)
	assist.NoError(t, err)
	assist.Len(t, 2, len(got))
	assist.Equal(t, "First", got[0].Title)
	assist.Equal(t, "Second", got[1].Title)
	assist.Equal(t, 2024, *got[0].Year)

	// Somebody else's list is separate, and so is deleting.
	theirs, err := repo.ListByUser(ctx, other)
	assist.NoError(t, err)
	assist.Len(t, 0, len(theirs))
	assist.ErrorIs(t, repo.Delete(ctx, other, first.ID), postgres.ErrNotFound)
	assist.NoError(t, repo.Delete(ctx, user, first.ID))
	assist.ErrorIs(t, repo.Delete(ctx, user, first.ID), postgres.ErrNotFound)
	got, _ = repo.ListByUser(ctx, user)
	assist.Len(t, 1, len(got))
}

func TestApplicantWorkRepository_refusesAnotherUsersImageAndDuplicates(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewApplicantWorkRepository(pool)
	ctx := context.Background()
	user, other := uuid.New(), uuid.New()
	integration.InsertUser(t, pool, user, identity.RolePublic)
	integration.InsertUser(t, pool, other, identity.RolePublic)
	theirImage := uploadedAsset(t, pool, other, 1, "")

	_, err := repo.Add(ctx, onboarding.ApplicantWork{ID: uuid.New(), UserID: user, MediaAssetID: theirImage, Title: "Stolen", CreatedAt: time.Now().UTC()})
	assist.ErrorIs(t, err, postgres.ErrNotFound)
	_, err = repo.Add(ctx, onboarding.ApplicantWork{ID: uuid.New(), UserID: user, MediaAssetID: uuid.New(), Title: "Nonexistent", CreatedAt: time.Now().UTC()})
	assist.ErrorIs(t, err, postgres.ErrNotFound)

	mine := uploadedAsset(t, pool, user, 2, "")
	_, err = repo.Add(ctx, onboarding.ApplicantWork{ID: uuid.New(), UserID: user, MediaAssetID: mine, Title: "One", CreatedAt: time.Now().UTC()})
	assist.NoError(t, err)
	_, err = repo.Add(ctx, onboarding.ApplicantWork{ID: uuid.New(), UserID: user, MediaAssetID: mine, Title: "Twice", CreatedAt: time.Now().UTC()})
	assist.ErrorIs(t, err, postgres.ErrConflict)
}

func TestOnboardingRepository_keepsTheSubmittedWorksOnTheApplication(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewOnboardingRepository(pool)
	ctx := context.Background()

	app := newPendingApp(uuid.New(), "works_handle")
	app.Works = []onboarding.ApplicantWork{
		{ID: uuid.New(), MediaAssetID: uuid.New(), URL: "https://res.cloudinary.com/c/image/upload/a.webp", Blurhash: "H1", Title: "One", Position: 0},
		{ID: uuid.New(), MediaAssetID: uuid.New(), URL: "https://res.cloudinary.com/c/image/upload/b.webp", Title: "Two", Position: 1},
	}
	saved, err := repo.Save(ctx, app)
	assist.NoError(t, err)
	assist.Len(t, 2, len(saved.Works))
	assist.Equal(t, "One", saved.Works[0].Title)
	assist.Equal(t, "H1", saved.Works[0].Blurhash)

	none, err := repo.Save(ctx, newPendingApp(uuid.New(), "no_works_handle"))
	assist.NoError(t, err)
	assist.Len(t, 0, len(none.Works))
}
