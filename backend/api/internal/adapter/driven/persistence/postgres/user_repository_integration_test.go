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

func TestUserRepository_CreateWithPassword_canCreateAVerifiedAccount(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewUserRepository(pool)
	ctx := context.Background()

	at := time.Now().UTC().Truncate(time.Second)
	verified := identity.User{ID: uuid.New(), Email: "v@example.com", Role: identity.RolePublic, EmailVerifiedAt: &at}
	plain := identity.User{ID: uuid.New(), Email: "p@example.com", Role: identity.RolePublic}
	assist.NoError(t, repo.CreateWithPassword(ctx, verified, "hash"))
	assist.NoError(t, repo.CreateWithPassword(ctx, plain, "hash"))

	got, err := repo.GetByID(ctx, verified.ID)
	assist.NoError(t, err)
	assist.NotNil(t, got.EmailVerifiedAt)

	got, err = repo.GetByID(ctx, plain.ID)
	assist.NoError(t, err)
	if got.EmailVerifiedAt != nil {
		t.Error("an account created without EmailVerifiedAt must stay unverified")
	}
}
