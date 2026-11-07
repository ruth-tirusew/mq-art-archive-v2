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

func TestOnboardingRepository_List_filtersPagesAndOrders(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewOnboardingRepository(pool)
	ctx := context.Background()

	base := time.Now().UTC().Add(-time.Hour)
	save := func(handle string, age time.Duration, status onboarding.ApprovalStatus, reviewedAgo time.Duration) onboarding.OnboardingApplication {
		app := newPendingApp(uuid.New(), handle)
		app.CreatedAt = base.Add(age)
		app.UpdatedAt = app.CreatedAt
		app.Status = status
		if status != onboarding.ApprovalStatusPending {
			at := base.Add(reviewedAgo)
			app.ReviewedAt = &at
		}
		saved, err := repo.Save(ctx, app)
		assist.NoError(t, err)
		return *saved
	}
	oldPending := save("old_pending", 0, onboarding.ApprovalStatusPending, 0)
	newPending := save("new_pending", time.Minute, onboarding.ApprovalStatusPending, 0)
	early := save("early_approved", 2*time.Minute, onboarding.ApprovalStatusApproved, 10*time.Minute)
	late := save("late_rejected", 3*time.Minute, onboarding.ApprovalStatusRejected, 20*time.Minute)

	pendingStatus := onboarding.ApprovalStatusPending
	got, total, err := repo.List(ctx, onboarding.ListFilter{Status: &pendingStatus, Limit: 10})
	assist.NoError(t, err)
	assist.Equal(t, 2, total)
	assist.Len(t, 2, len(got))
	assist.Equal(t, oldPending.ID, got[0].ID) // the queue is oldest first
	assist.Equal(t, newPending.ID, got[1].ID)

	// Everything, newest decision first; undecided ones sort by when they were submitted.
	all, total, err := repo.List(ctx, onboarding.ListFilter{Limit: 10})
	assist.NoError(t, err)
	assist.Equal(t, 4, total)
	assist.Equal(t, late.ID, all[0].ID)
	assist.Equal(t, early.ID, all[1].ID)

	// Paging keeps the total.
	page, total, err := repo.List(ctx, onboarding.ListFilter{Limit: 1, Offset: 1})
	assist.NoError(t, err)
	assist.Equal(t, 4, total)
	assist.Len(t, 1, len(page))
	assist.Equal(t, all[1].ID, page[0].ID)

	rejected := onboarding.ApprovalStatusRejected
	only, total, err := repo.List(ctx, onboarding.ListFilter{Status: &rejected, Limit: 10})
	assist.NoError(t, err)
	assist.Equal(t, 1, total)
	assist.Equal(t, late.ID, only[0].ID)
}

func TestOnboardingRepository_keepsTheSubmittedAnswerSnapshotThroughADecision(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewOnboardingRepository(pool)
	ctx := context.Background()

	app := newPendingApp(uuid.New(), "snapshot_handle")
	app.QuestionnaireAnswers = []byte(`[{"key":"bio","body":"as submitted"}]`)
	saved, err := repo.Save(ctx, app)
	assist.NoError(t, err)
	assist.Equal(t, `{"key": "bio", "body": "as submitted"}`, string(saved.QuestionnaireAnswers[1:len(saved.QuestionnaireAnswers)-1]))

	reviewer, at := uuid.New(), time.Now().UTC()
	decided := *saved
	decided.Status, decided.ReviewedBy, decided.ReviewedAt, decided.UpdatedAt = onboarding.ApprovalStatusApproved, &reviewer, &at, at
	decided.QuestionnaireAnswers = nil // Transition must never touch the snapshot
	got, err := repo.Transition(ctx, decided, onboarding.ApprovalStatusPending)
	assist.NoError(t, err)
	assist.Equal(t, len(saved.QuestionnaireAnswers), len(got.QuestionnaireAnswers))

	none, err := repo.Save(ctx, newPendingApp(uuid.New(), "no_snapshot_handle"))
	assist.NoError(t, err)
	if none.QuestionnaireAnswers != nil {
		t.Errorf("an application without answers should read back nil, got %s", none.QuestionnaireAnswers)
	}
}
