package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/questionnaire"
	"github.com/mq/api/internal/port/outbound"
)

type QuestionnaireAnswerRepository struct {
	pool *Pool
}

func NewQuestionnaireAnswerRepository(pool *Pool) outbound.QuestionnaireAnswerRepository {
	return &QuestionnaireAnswerRepository{pool: pool}
}

const answerColumns = `user_id, question_key, revision_id, body, format, lang, created_at, updated_at`

func (r *QuestionnaireAnswerRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]questionnaire.Answer, error) {
	return r.ListByUsers(ctx, []uuid.UUID{userID})
}

func (r *QuestionnaireAnswerRepository) ListByUsers(ctx context.Context, userIDs []uuid.UUID) ([]questionnaire.Answer, error) {
	answers := []questionnaire.Answer{}
	if len(userIDs) == 0 {
		return answers, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT `+answerColumns+` FROM questionnaire_answers
		WHERE user_id = ANY($1) ORDER BY user_id, question_key
	`, userIDs)
	if err != nil {
		return nil, fmt.Errorf("list questionnaire answers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var a questionnaire.Answer
		var format, lang string
		if err := rows.Scan(&a.UserID, &a.QuestionKey, &a.RevisionID, &a.Body, &format, &lang, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("list questionnaire answers: %w", err)
		}
		a.Format, a.Lang = questionnaire.Format(format), questionnaire.Lang(lang)
		answers = append(answers, a)
	}
	return answers, rows.Err()
}

func (r *QuestionnaireAnswerRepository) Upsert(ctx context.Context, answers []questionnaire.Answer) error {
	if len(answers) == 0 {
		return nil
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("upsert answers: begin: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, a := range answers {
		if _, err := tx.Exec(ctx, `
			INSERT INTO questionnaire_answers (`+answerColumns+`)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			ON CONFLICT (user_id, question_key) DO UPDATE SET
				revision_id = EXCLUDED.revision_id,
				body = EXCLUDED.body,
				format = EXCLUDED.format,
				lang = EXCLUDED.lang,
				updated_at = EXCLUDED.updated_at
		`, a.UserID, a.QuestionKey, a.RevisionID, a.Body, string(a.Format), string(a.Lang), a.CreatedAt, a.UpdatedAt); err != nil {
			return fmt.Errorf("upsert answer %q: %w", a.QuestionKey, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("upsert answers: commit: %w", err)
	}
	return nil
}

func (r *QuestionnaireAnswerRepository) IsLockedForReview(ctx context.Context, userID uuid.UUID) (bool, error) {
	var locked bool
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE((
			SELECT status = 'pending' FROM onboarding_applications
			WHERE applicant_id = $1 ORDER BY created_at DESC LIMIT 1
		), false)
	`, userID).Scan(&locked)
	if err != nil {
		return false, fmt.Errorf("check review lock: %w", err)
	}
	return locked, nil
}
