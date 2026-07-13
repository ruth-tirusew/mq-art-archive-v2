//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/domain/profile"
	"github.com/mq/api/internal/testutil/assist"
)

func InsertUser(t *testing.T, pool *postgres.Pool, id uuid.UUID, role identity.Role) {
	t.Helper()
	now := time.Now().UTC()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (id, email, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, id, id.String()+"@example.com", string(role), now, now)
	assist.NoError(t, err)
}

func InsertArtistProfile(t *testing.T, pool *postgres.Pool, slug, displayName string) (artistID, userID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	userID = uuid.New()
	now := time.Now().UTC()

	InsertUser(t, pool, userID, identity.RoleArtist)

	artistID = uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO artist_profiles (
			id, user_id, slug, handle, display_name, bio, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, artistID, userID, slug, slug, displayName, "", string(profile.ProfileStatusApproved), now, now)
	assist.NoError(t, err)

	return artistID, userID
}
