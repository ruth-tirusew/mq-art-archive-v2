package digest

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	digestdomain "github.com/mq/api/internal/domain/digest"
	"github.com/mq/api/internal/testutil/assist"
)

type fakeRunRepo struct {
	lastCompletedRun  func(ctx context.Context) (*digestdomain.Run, error)
	getIncompleteRun  func(ctx context.Context) (*digestdomain.Run, error)
	startRunCallCount int
}

func (f *fakeRunRepo) LastCompletedRun(ctx context.Context) (*digestdomain.Run, error) {
	if f.lastCompletedRun != nil {
		return f.lastCompletedRun(ctx)
	}
	return nil, apperrors.ErrNotFound
}

func (f *fakeRunRepo) GetIncompleteRun(ctx context.Context) (*digestdomain.Run, error) {
	if f.getIncompleteRun != nil {
		return f.getIncompleteRun(ctx)
	}
	return nil, apperrors.ErrNotFound
}

func (f *fakeRunRepo) StartRun(ctx context.Context, periodStart, periodEnd time.Time) (*digestdomain.Run, error) {
	f.startRunCallCount++
	return &digestdomain.Run{ID: uuid.New(), PeriodStart: periodStart, PeriodEnd: periodEnd, StartedAt: time.Now().UTC()}, nil
}

func (f *fakeRunRepo) CompleteRun(ctx context.Context, runID uuid.UUID) error { return nil }

func (f *fakeRunRepo) HasSuccessfulDelivery(ctx context.Context, runID, recipientID uuid.UUID, channel digestdomain.Channel) (bool, error) {
	return false, nil
}

func (f *fakeRunRepo) RecordDelivery(ctx context.Context, runID, recipientID uuid.UUID, channel digestdomain.Channel, sentAt time.Time, deliveryErr error) error {
	return nil
}

func TestResolvePeriodStart_noPriorRun_fallsBackToDefaultLookback(t *testing.T) {
	svc := &Service{runs: &fakeRunRepo{}}
	now := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	start, err := svc.resolvePeriodStart(context.Background(), now)
	assist.NoError(t, err)
	assist.Equal(t, now.Add(-defaultLookback), start)
}

func TestResolvePeriodStart_usesPreviousRunsPeriodEnd(t *testing.T) {
	prevEnd := time.Date(2026, 1, 10, 9, 0, 0, 0, time.UTC)
	svc := &Service{runs: &fakeRunRepo{
		lastCompletedRun: func(ctx context.Context) (*digestdomain.Run, error) {
			return &digestdomain.Run{PeriodEnd: prevEnd}, nil
		},
	}}
	now := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	start, err := svc.resolvePeriodStart(context.Background(), now)
	assist.NoError(t, err)
	assist.Equal(t, prevEnd, start)
}

func TestResolvePeriodStart_propagatesUnexpectedError(t *testing.T) {
	boom := context.DeadlineExceeded
	svc := &Service{runs: &fakeRunRepo{
		lastCompletedRun: func(ctx context.Context) (*digestdomain.Run, error) {
			return nil, boom
		},
	}}

	_, err := svc.resolvePeriodStart(context.Background(), time.Now())
	assist.ErrorIs(t, err, boom)
}

func TestCurrentOrNewRun_resumesExistingIncompleteRun(t *testing.T) {
	existing := &digestdomain.Run{ID: uuid.New(), PeriodStart: time.Now().Add(-time.Hour), PeriodEnd: time.Now()}
	runs := &fakeRunRepo{
		getIncompleteRun: func(ctx context.Context) (*digestdomain.Run, error) {
			return existing, nil
		},
	}
	svc := &Service{runs: runs}

	got, err := svc.currentOrNewRun(context.Background())
	assist.NoError(t, err)
	assist.Equal(t, existing.ID, got.ID)
	assist.Equal(t, 0, runs.startRunCallCount)
}

func TestCurrentOrNewRun_startsFreshWhenNoneIncomplete(t *testing.T) {
	runs := &fakeRunRepo{}
	svc := &Service{runs: runs}

	got, err := svc.currentOrNewRun(context.Background())
	assist.NoError(t, err)
	assist.NotNil(t, got)
	assist.Equal(t, 1, runs.startRunCallCount)
}

func TestCurrentOrNewRun_propagatesUnexpectedError(t *testing.T) {
	boom := context.DeadlineExceeded
	runs := &fakeRunRepo{
		getIncompleteRun: func(ctx context.Context) (*digestdomain.Run, error) {
			return nil, boom
		},
	}
	svc := &Service{runs: runs}

	_, err := svc.currentOrNewRun(context.Background())
	assist.ErrorIs(t, err, boom)
	assist.Equal(t, 0, runs.startRunCallCount)
}
