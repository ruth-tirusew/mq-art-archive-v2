package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mq/api/internal/domain/questionnaire"
	"github.com/mq/api/internal/port/outbound"
)

type QuestionnaireNudgeRepository struct {
	pool *Pool
}

func NewQuestionnaireNudgeRepository(pool *Pool) outbound.QuestionnaireNudgeRepository {
	return &QuestionnaireNudgeRepository{pool: pool}
}

func (r *QuestionnaireNudgeRepository) ListUsersMissingAnswers(ctx context.Context, revisionID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ap.user_id
		FROM artist_profiles ap
		WHERE ap.status = 'approved'
		  AND EXISTS (
			SELECT 1 FROM questionnaire_questions q
			WHERE q.revision_id = $1
			  AND NOT EXISTS (
				SELECT 1 FROM questionnaire_answers a
				WHERE a.user_id = ap.user_id AND a.question_key = q.key AND btrim(a.body) <> ''
			  )
		  )
		ORDER BY ap.created_at
	`, revisionID)
	if err != nil {
		return nil, fmt.Errorf("list users missing answers: %w", err)
	}
	defer rows.Close()

	ids := []uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("list users missing answers: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *QuestionnaireNudgeRepository) Create(ctx context.Context, revisionID uuid.UUID, sentBy *uuid.UUID, userIDs []uuid.UUID, now time.Time) ([]uuid.UUID, error) {
	created := []uuid.UUID{}
	if len(userIDs) == 0 {
		return created, nil
	}
	rows, err := r.pool.Query(ctx, `
		INSERT INTO questionnaire_nudges (revision_id, user_id, sent_by, sent_at)
		SELECT $1, u, $2, $3 FROM unnest($4::uuid[]) AS u
		ON CONFLICT (revision_id, user_id) DO NOTHING
		RETURNING user_id
	`, revisionID, sentBy, now, userIDs)
	if err != nil {
		return nil, fmt.Errorf("create nudges: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("create nudges: %w", err)
		}
		created = append(created, id)
	}
	return created, rows.Err()
}

func (r *QuestionnaireNudgeRepository) GetActive(ctx context.Context, userID uuid.UUID) (*questionnaire.Nudge, error) {
	var n questionnaire.Nudge
	err := r.pool.QueryRow(ctx, `
		SELECT revision_id, user_id, sent_by, sent_at, dismissed_at
		FROM questionnaire_nudges
		WHERE user_id = $1 AND dismissed_at IS NULL
		ORDER BY sent_at DESC LIMIT 1
	`, userID).Scan(&n.RevisionID, &n.UserID, &n.SentBy, &n.SentAt, &n.DismissedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get active nudge: %w", err)
	}
	return &n, nil
}

func (r *QuestionnaireNudgeRepository) Dismiss(ctx context.Context, revisionID, userID uuid.UUID, now time.Time) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE questionnaire_nudges SET dismissed_at = COALESCE(dismissed_at, $3)
		WHERE revision_id = $1 AND user_id = $2
	`, revisionID, userID, now)
	if err != nil {
		return fmt.Errorf("dismiss nudge: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
