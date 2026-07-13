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

func TestOnboardingRepository_integration(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewOnboardingRepository(pool)
	ctx := context.Background()

	now := time.Now().UTC()
	app := onboarding.OnboardingApplication{
		ID:            uuid.New(),
		ApplicantID:   uuid.New(),
		ApplicantType: onboarding.ApplicantTypeArtist,
		DisplayName:   "New Studio",
		Status:        onboarding.ApprovalStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	created, err := repo.Save(ctx, app)
	assist.NoError(t, err)
	assist.Equal(t, onboarding.ApprovalStatusPending, created.Status)

	pending, err := repo.ListPending(ctx)
	assist.NoError(t, err)
	assist.GreaterOrEqual(t, len(pending), 1)

	got, err := repo.GetByID(ctx, created.ID)
	assist.NoError(t, err)
	assist.Equal(t, "New Studio", got.DisplayName)

	reviewerID := uuid.New()
	reviewedAt := time.Now().UTC()
	created.Status = onboarding.ApprovalStatusApproved
	created.ReviewedBy = &reviewerID
	created.ReviewedAt = &reviewedAt
	created.UpdatedAt = reviewedAt

	updated, err := repo.Save(ctx, *created)
	assist.NoError(t, err)
	assist.Equal(t, onboarding.ApprovalStatusApproved, updated.Status)

	pending, err = repo.ListPending(ctx)
	assist.NoError(t, err)
	for _, p := range pending {
		assist.NotEqual(t, created.ID, p.ID)
	}

	_, err = repo.GetByID(ctx, uuid.New())
	assist.ErrorIs(t, err, postgres.ErrNotFound)
}
