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

func newPendingApp(applicant uuid.UUID, handle string) onboarding.OnboardingApplication {
	now := time.Now().UTC()
	return onboarding.OnboardingApplication{
		ID: uuid.New(), ApplicantID: applicant, ApplicantType: onboarding.ApplicantTypeArtist,
		DisplayName: "Studio", RequestedHandle: handle, Status: onboarding.ApprovalStatusPending,
		CreatedAt: now, UpdatedAt: now,
	}
}

func TestOnboardingRepository_Transition_onlyOneDecisionWins(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewOnboardingRepository(pool)
	ctx := context.Background()

	saved, err := repo.Save(ctx, newPendingApp(uuid.New(), "first_handle"))
	assist.NoError(t, err)

	reviewer := uuid.New()
	now := time.Now().UTC()
	decided := *saved
	decided.Status = onboarding.ApprovalStatusApproved
	decided.ReviewNotes = "welcome"
	decided.ReviewedBy = &reviewer
	decided.ReviewedAt = &now
	decided.UpdatedAt = now

	got, err := repo.Transition(ctx, decided, onboarding.ApprovalStatusPending)
	assist.NoError(t, err)
	assist.Equal(t, onboarding.ApprovalStatusApproved, got.Status)
	assist.Equal(t, "welcome", got.ReviewNotes)

	// A second decision from a stale read no longer matches status = pending.
	decided.Status = onboarding.ApprovalStatusRejected
	_, err = repo.Transition(ctx, decided, onboarding.ApprovalStatusPending)
	assist.ErrorIs(t, err, postgres.ErrConflict)

	// Reopening is the same call with the statuses reversed.
	reopened := *got
	reopened.Status = onboarding.ApprovalStatusPending
	back, err := repo.Transition(ctx, reopened, onboarding.ApprovalStatusApproved)
	assist.NoError(t, err)
	assist.Equal(t, onboarding.ApprovalStatusPending, back.Status)

	decided.ID = uuid.New()
	_, err = repo.Transition(ctx, decided, onboarding.ApprovalStatusPending)
	assist.ErrorIs(t, err, postgres.ErrNotFound)
}

func TestOnboardingRepository_Save_mapsUniqueViolationsToConflict(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewOnboardingRepository(pool)
	ctx := context.Background()

	applicant := uuid.New()
	_, err := repo.Save(ctx, newPendingApp(applicant, "taken_handle"))
	assist.NoError(t, err)

	// Same handle, different applicant.
	_, err = repo.Save(ctx, newPendingApp(uuid.New(), "taken_handle"))
	assist.ErrorIs(t, err, postgres.ErrConflict)

	// Same applicant, second pending application.
	_, err = repo.Save(ctx, newPendingApp(applicant, "another_handle"))
	assist.ErrorIs(t, err, postgres.ErrConflict)
}
