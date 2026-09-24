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
	lastCompletedRun func(ctx context.Context) (*digestdomain.Run, error)
}

func (f *fakeRunRepo) LastCompletedRun(ctx context.Context) (*digestdomain.Run, error) {
	if f.lastCompletedRun != nil {
		return f.lastCompletedRun(ctx)
	}
	return nil, apperrors.ErrNotFound
}

func (f *fakeRunRepo) StartRun(ctx context.Context, periodStart, periodEnd time.Time) (*digestdomain.Run, error) {
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
