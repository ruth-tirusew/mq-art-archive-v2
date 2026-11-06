//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	"github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

func TestInviteRepository_Reopen_restoresRevokedAndClaimedInvites(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewInviteRepository(pool)
	ctx := context.Background()
	now := time.Now().UTC()

	inv, err := repo.Create(ctx, onboarding.Invite{
		ID: uuid.New(), TokenHash: "hash-1", Email: "ada@example.com", RequestedHandle: "ada_1",
		ExpiresAt: now.Add(24 * time.Hour), CreatedAt: now,
	})
	assist.NoError(t, err)

	assist.NoError(t, repo.Revoke(ctx, inv.ID, now))
	assist.NoError(t, repo.Reopen(ctx, inv.ID))
	got, err := repo.GetByID(ctx, inv.ID)
	assist.NoError(t, err)
	assist.Equal(t, onboarding.InviteLive, got.Status(now))

	_, err = repo.Claim(ctx, inv.ID, integration.InsertAdminUser(t, pool), now)
	assist.NoError(t, err)
	assist.NoError(t, repo.Reopen(ctx, inv.ID))
	got, err = repo.GetByID(ctx, inv.ID)
	assist.NoError(t, err)
	assist.Equal(t, onboarding.InviteLive, got.Status(now))

	assist.ErrorIs(t, repo.Reopen(ctx, uuid.New()), postgres.ErrNotFound)
}
