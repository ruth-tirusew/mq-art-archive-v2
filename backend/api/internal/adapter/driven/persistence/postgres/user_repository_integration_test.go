//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

func TestUserRepository_integration(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewUserRepository(pool)
	ctx := context.Background()

	id := uuid.New()
	now := time.Now().UTC()

	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, email, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, id, "artist@example.com", string(identity.RoleArtist), now, now)
	assist.NoError(t, err)

	got, err := repo.GetByID(ctx, id)
	assist.NoError(t, err)
	assist.Equal(t, "artist@example.com", got.Email)
	assist.Equal(t, identity.RoleArtist, got.Role)

	_, err = repo.GetByID(ctx, uuid.New())
	assist.ErrorIs(t, err, postgres.ErrNotFound)
}
