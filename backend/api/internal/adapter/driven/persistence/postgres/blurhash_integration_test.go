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
	"github.com/mq/api/internal/domain/media"
	"github.com/mq/api/internal/domain/profile"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

const (
	uploadedURL = "https://res.cloudinary.com/cloud/image/upload/v1/mq/one.webp"
	otherURL    = "https://res.cloudinary.com/cloud/image/upload/v1/mq/two.webp"
	externalURL = "https://images.unsplash.com/photo-1?w=1200"
)

func registerAsset(t *testing.T, pool *postgres.Pool, owner uuid.UUID, url, hash string) {
	t.Helper()
	w, h, b := 100, 100, 1000
	now := time.Now().UTC()
	_, err := postgres.NewMediaAssetRepository(pool).Create(context.Background(), media.Asset{
		ID: uuid.New(), OwnerUserID: owner, PublicID: "mq/" + uuid.NewString(), SecureURL: url, ResourceType: "image",
		Width: &w, Height: &h, Bytes: &b, Folder: "mq", Blurhash: hash, CreatedAt: now, UpdatedAt: now,
	})
	assist.NoError(t, err)
}

func TestArtMedia_takesItsBlurhashFromTheUploadedAsset(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewArtPostRepository(pool)
	ctx := context.Background()
	artistID, userID := integration.InsertArtistProfile(t, pool, "blur-artist", "Blur Artist")
	registerAsset(t, pool, userID, uploadedURL, "HASH-ONE")
	registerAsset(t, pool, userID, otherURL, "HASH-TWO")

	now := time.Now().UTC()
	post, err := repo.Create(ctx, art.ArtPost{
		ID: uuid.New(), ArtistID: artistID, Title: "Piece", Status: art.ArtStatusDraft, CreatedAt: now, UpdatedAt: now,
		Media: []art.MediaAsset{{URL: uploadedURL}, {URL: externalURL, SortOrder: 1}},
	})
	assist.NoError(t, err)
	assist.Equal(t, "HASH-ONE", post.Media[0].Blurhash)
	assist.Equal(t, "", post.Media[1].Blurhash) // not one of ours: no hash until the backfill

	// Editing the post keeps each image's hash; a swapped URL picks up the new asset's hash.
	byURL := map[string]art.MediaAsset{}
	for _, m := range post.Media {
		byURL[m.URL] = m
	}
	edited := *post
	edited.Media = []art.MediaAsset{
		{ID: byURL[uploadedURL].ID, URL: otherURL},
		{ID: byURL[externalURL].ID, URL: externalURL, SortOrder: 1},
	}
	edited.UpdatedAt = time.Now().UTC()
	updated, err := repo.Update(ctx, edited)
	assist.NoError(t, err)
	assist.Equal(t, "HASH-TWO", updated.Media[0].Blurhash)
	assist.Equal(t, "", updated.Media[1].Blurhash)

	// A row that already has a hash keeps it when only other fields change.
	_, err = pool.Exec(ctx, `UPDATE art_post_media SET blurhash = 'BACKFILLED' WHERE id = $1`, byURL[externalURL].ID)
	assist.NoError(t, err)
	edited.Title = "Renamed"
	again, err := repo.Update(ctx, *updated)
	assist.NoError(t, err)
	assist.Equal(t, "BACKFILLED", again.Media[1].Blurhash)
}

func TestPortrait_takesItsBlurhashFromTheUploadedAsset(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewProfileRepository(pool)
	ctx := context.Background()
	userID := uuid.New()
	integration.InsertUser(t, pool, userID, identity.RoleArtist)
	registerAsset(t, pool, userID, uploadedURL, "PORTRAIT-HASH")

	now := time.Now().UTC()
	saved, err := repo.SaveArtist(ctx, profile.ArtistProfile{
		ID: uuid.New(), UserID: userID, Slug: "portrait-artist", Handle: "portrait-artist", DisplayName: "P",
		PortraitURL: uploadedURL, Status: profile.ProfileStatusDraft, CreatedAt: now, UpdatedAt: now,
	})
	assist.NoError(t, err)
	assist.Equal(t, "PORTRAIT-HASH", saved.PortraitBlurhash)

	// Unrelated edits keep it; clearing the portrait clears it; an outside URL has none yet.
	saved.Bio = "new bio"
	same, err := repo.SaveArtist(ctx, *saved)
	assist.NoError(t, err)
	assist.Equal(t, "PORTRAIT-HASH", same.PortraitBlurhash)

	same.PortraitURL = externalURL
	external, err := repo.SaveArtist(ctx, *same)
	assist.NoError(t, err)
	assist.Equal(t, "", external.PortraitBlurhash)
}
