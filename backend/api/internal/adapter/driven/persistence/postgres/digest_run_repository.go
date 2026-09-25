package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mq/api/internal/domain/digest"
	"github.com/mq/api/internal/port/outbound"
)

type DigestRunRepository struct {
	pool *Pool
}

func NewDigestRunRepository(pool *Pool) outbound.DigestRunRepository {
	return &DigestRunRepository{pool: pool}
}

func (r *DigestRunRepository) LastCompletedRun(ctx context.Context) (*digest.Run, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, period_start, period_end, started_at, completed_at
		FROM digest_runs
		WHERE completed_at IS NOT NULL
		ORDER BY completed_at DESC
		LIMIT 1
	`)

	var run digest.Run
	err := row.Scan(&run.ID, &run.PeriodStart, &run.PeriodEnd, &run.StartedAt, &run.CompletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("last completed digest run: %w", err)
	}
	return &run, nil
}

func (r *DigestRunRepository) GetIncompleteRun(ctx context.Context) (*digest.Run, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, period_start, period_end, started_at, completed_at
		FROM digest_runs
		WHERE completed_at IS NULL
		ORDER BY started_at DESC
		LIMIT 1
	`)

	var run digest.Run
	err := row.Scan(&run.ID, &run.PeriodStart, &run.PeriodEnd, &run.StartedAt, &run.CompletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get incomplete digest run: %w", err)
	}
	return &run, nil
}

func (r *DigestRunRepository) StartRun(ctx context.Context, periodStart, periodEnd time.Time) (*digest.Run, error) {
	run := digest.Run{
		ID:          uuid.New(),
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		StartedAt:   time.Now().UTC(),
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO digest_runs (id, period_start, period_end, started_at)
		VALUES ($1, $2, $3, $4)
	`, run.ID, run.PeriodStart, run.PeriodEnd, run.StartedAt)
	if err != nil {
		return nil, fmt.Errorf("start digest run: %w", err)
	}
	return &run, nil
}

func (r *DigestRunRepository) CompleteRun(ctx context.Context, runID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE digest_runs SET completed_at = NOW() WHERE id = $1`, runID)
	if err != nil {
		return fmt.Errorf("complete digest run: %w", err)
	}
	return nil
}

func (r *DigestRunRepository) HasSuccessfulDelivery(ctx context.Context, runID, recipientID uuid.UUID, channel digest.Channel) (bool, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM digest_deliveries
			WHERE run_id = $1 AND recipient_id = $2 AND channel = $3
			  AND sent_at IS NOT NULL AND error IS NULL
		)
	`, runID, recipientID, string(channel))
	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("check digest delivery: %w", err)
	}
	return exists, nil
}

func (r *DigestRunRepository) RecordDelivery(ctx context.Context, runID, recipientID uuid.UUID, channel digest.Channel, sentAt time.Time, deliveryErr error) error {
	var errMsg *string
	if deliveryErr != nil {
		msg := deliveryErr.Error()
		errMsg = &msg
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO digest_deliveries (id, run_id, recipient_id, channel, sent_at, error)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (run_id, recipient_id, channel) DO UPDATE SET
			sent_at = EXCLUDED.sent_at,
			error = EXCLUDED.error
	`, uuid.New(), runID, recipientID, string(channel), sentAt, errMsg)
	if err != nil {
		return fmt.Errorf("record digest delivery: %w", err)
	}
	return nil
}
