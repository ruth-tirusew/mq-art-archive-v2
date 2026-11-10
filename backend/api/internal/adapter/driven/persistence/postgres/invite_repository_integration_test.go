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

func TestInviteRepository_List_filtersByDerivedStatusAndPages(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewInviteRepository(pool)
	ctx := context.Background()
	now := time.Now().UTC()

	mk := func(n string, expires time.Time, created time.Time) onboarding.Invite {
		inv, err := repo.Create(ctx, onboarding.Invite{
			ID: uuid.New(), TokenHash: "h-" + n, Email: n + "@example.com", ExpiresAt: expires, CreatedAt: created,
		})
		assist.NoError(t, err)
		return *inv
	}
	live := mk("live", now.Add(time.Hour), now.Add(-1*time.Minute))
	expired := mk("expired", now.Add(-time.Hour), now.Add(-2*time.Minute))
	revoked := mk("revoked", now.Add(time.Hour), now.Add(-3*time.Minute))
	claimed := mk("claimed", now.Add(time.Hour), now.Add(-4*time.Minute))
	// Claimed and revoked both set: claimed wins, exactly as Invite.Status says.
	both := mk("both", now.Add(time.Hour), now.Add(-5*time.Minute))
	assist.NoError(t, repo.Revoke(ctx, revoked.ID, now))
	admin := integration.InsertAdminUser(t, pool)
	_, err := repo.Claim(ctx, claimed.ID, admin, now)
	assist.NoError(t, err)
	_, err = repo.Claim(ctx, both.ID, admin, now)
	assist.NoError(t, err)
	// Revoke refuses a claimed invite, so a row with both timestamps can only come from outside it.
	_, err = pool.Exec(ctx, `UPDATE artist_invites SET revoked_at = $2 WHERE id = $1`, both.ID, now)
	assist.NoError(t, err)

	ids := func(invs []onboarding.Invite) []uuid.UUID {
		out := []uuid.UUID{}
		for _, i := range invs {
			out = append(out, i.ID)
		}
		return out
	}
	for _, c := range []struct {
		status onboarding.InviteStatus
		want   []uuid.UUID
	}{
		{onboarding.InviteLive, []uuid.UUID{live.ID}},
		{onboarding.InviteExpired, []uuid.UUID{expired.ID}},
		{onboarding.InviteRevoked, []uuid.UUID{revoked.ID}},
		{onboarding.InviteClaimed, []uuid.UUID{claimed.ID, both.ID}},
	} {
		st := c.status
		got, total, err := repo.List(ctx, onboarding.InviteListFilter{Status: &st, Limit: 50}, now)
		assist.NoError(t, err)
		assist.Equal(t, len(c.want), total)
		for i, id := range c.want {
			assist.Equal(t, id, ids(got)[i])
		}
	}

	// Paging: newest first, total unaffected by the window.
	page, total, err := repo.List(ctx, onboarding.InviteListFilter{Limit: 2, Offset: 1}, now)
	assist.NoError(t, err)
	assist.Equal(t, 5, total)
	assist.Len(t, 2, len(page))
	assist.Equal(t, expired.ID, page[0].ID)
	assist.Equal(t, revoked.ID, page[1].ID)
}
