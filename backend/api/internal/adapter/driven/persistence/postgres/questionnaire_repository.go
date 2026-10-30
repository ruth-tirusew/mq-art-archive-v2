package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/questionnaire"
	"github.com/mq/api/internal/port/outbound"
)

type QuestionnaireRepository struct {
	pool *Pool
}

func NewQuestionnaireRepository(pool *Pool) outbound.QuestionnaireRepository {
	return &QuestionnaireRepository{pool: pool}
}

const questionnaireRevisionColumns = `id, number, status, change_kind, note, created_by, published_at, created_at, updated_at`

const questionColumns = `id, revision_id, key, position, type, prompt_en, prompt_am, help_en, help_am,
	required, is_public, max_length, options`

// questionnaireQueryer is satisfied by both the pool and an open transaction.
type questionnaireQueryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (r *QuestionnaireRepository) GetDraft(ctx context.Context) (*questionnaire.Revision, error) {
	return r.getOne(ctx, `WHERE status = 'draft'`)
}

func (r *QuestionnaireRepository) GetLatestPublished(ctx context.Context) (*questionnaire.Revision, error) {
	return r.getOne(ctx, `WHERE status = 'published' ORDER BY number DESC LIMIT 1`)
}

func (r *QuestionnaireRepository) GetRevision(ctx context.Context, id uuid.UUID) (*questionnaire.Revision, error) {
	return r.getOne(ctx, `WHERE id = $1`, id)
}

func (r *QuestionnaireRepository) getOne(ctx context.Context, where string, args ...any) (*questionnaire.Revision, error) {
	rev, err := scanQuestionnaireRevision(r.pool.QueryRow(ctx, `SELECT `+questionnaireRevisionColumns+` FROM questionnaire_revisions `+where, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get questionnaire revision: %w", err)
	}
	if rev.Questions, err = loadQuestions(ctx, r.pool, rev.ID); err != nil {
		return nil, err
	}
	return &rev, nil
}

func (r *QuestionnaireRepository) ListPublished(ctx context.Context) ([]questionnaire.Revision, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+questionnaireRevisionColumns+` FROM questionnaire_revisions
		WHERE status = 'published' ORDER BY number DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list questionnaire revisions: %w", err)
	}
	defer rows.Close()

	revs := []questionnaire.Revision{}
	for rows.Next() {
		rev, err := scanQuestionnaireRevision(rows)
		if err != nil {
			return nil, fmt.Errorf("list questionnaire revisions: %w", err)
		}
		revs = append(revs, rev)
	}
	return revs, rows.Err()
}

func (r *QuestionnaireRepository) SaveDraft(ctx context.Context, draft questionnaire.Revision) (*questionnaire.Revision, error) {
	if draft.Status != questionnaire.RevisionDraft {
		return nil, fmt.Errorf("save draft: revision is %q: %w", draft.Status, ErrConflict)
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("save draft: begin: %w", err)
	}
	defer tx.Rollback(ctx)

	// The WHERE on the update branch is what makes published revisions immutable: touching one
	// matches no row, which surfaces below as ErrNoRows.
	var id uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO questionnaire_revisions (id, status, change_kind, note, created_by, created_at, updated_at)
		VALUES ($1, 'draft', $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			change_kind = EXCLUDED.change_kind,
			note = EXCLUDED.note,
			updated_at = EXCLUDED.updated_at
		WHERE questionnaire_revisions.status = 'draft'
		RETURNING id
	`, draft.ID, string(draft.ChangeKind), draft.Note, draft.CreatedBy, draft.CreatedAt, draft.UpdatedAt).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.Is(err, pgx.ErrNoRows) || (errors.As(err, &pgErr) && pgErr.Code == "23505") {
			return nil, fmt.Errorf("save draft: %w", ErrConflict)
		}
		return nil, fmt.Errorf("save draft: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM questionnaire_questions WHERE revision_id = $1`, id); err != nil {
		return nil, fmt.Errorf("save draft: clear questions: %w", err)
	}
	for _, q := range draft.Questions {
		if q.ID == uuid.Nil {
			q.ID = uuid.New()
		}
		options := q.Options
		if options == nil {
			options = []questionnaire.Option{}
		}
		optionsJSON, err := json.Marshal(options)
		if err != nil {
			return nil, fmt.Errorf("save draft: encode options: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO questionnaire_questions (`+questionColumns+`)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		`, q.ID, id, q.Key, q.Position, string(q.Type), q.PromptEN, q.PromptAM, q.HelpEN, q.HelpAM,
			q.Required, q.IsPublic, q.MaxLength, optionsJSON); err != nil {
			return nil, fmt.Errorf("save draft: insert question %q: %w", q.Key, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("save draft: commit: %w", err)
	}
	return r.GetRevision(ctx, id)
}

func (r *QuestionnaireRepository) DiscardDraft(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM questionnaire_revisions WHERE id = $1 AND status = 'draft'`, id)
	if err != nil {
		return fmt.Errorf("discard draft: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *QuestionnaireRepository) Publish(ctx context.Context, id uuid.UUID, now time.Time) (*questionnaire.Revision, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("publish: begin: %w", err)
	}
	defer tx.Rollback(ctx)

	// Row lock serialises concurrent publishes of the same draft.
	rev, err := scanQuestionnaireRevision(tx.QueryRow(ctx,
		`SELECT `+questionnaireRevisionColumns+` FROM questionnaire_revisions WHERE id = $1 FOR UPDATE`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("publish: load: %w", err)
	}
	if rev.Status != questionnaire.RevisionDraft {
		return nil, fmt.Errorf("publish: revision already published: %w", ErrConflict)
	}
	if rev.Questions, err = loadQuestions(ctx, tx, id); err != nil {
		return nil, err
	}
	// Re-validate against what is actually stored, not what a caller last saw.
	if err := rev.Validate(); err != nil {
		var ve *questionnaire.ValidationError
		if errors.As(err, &ve) {
			return nil, fmt.Errorf("%w: %s", apperrors.ErrValidation, ve.Msg)
		}
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE questionnaire_revisions SET
			status = 'published',
			number = (SELECT COALESCE(MAX(number), 0) + 1 FROM questionnaire_revisions),
			published_at = $2,
			updated_at = $2
		WHERE id = $1
	`, id, now); err != nil {
		return nil, fmt.Errorf("publish: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("publish: commit: %w", err)
	}
	return r.GetRevision(ctx, id)
}

func loadQuestions(ctx context.Context, q questionnaireQueryer, revisionID uuid.UUID) ([]questionnaire.Question, error) {
	rows, err := q.Query(ctx, `SELECT `+questionColumns+` FROM questionnaire_questions WHERE revision_id = $1 ORDER BY position`, revisionID)
	if err != nil {
		return nil, fmt.Errorf("load questions: %w", err)
	}
	defer rows.Close()

	questions := []questionnaire.Question{}
	for rows.Next() {
		var qn questionnaire.Question
		var qType string
		var options []byte
		if err := rows.Scan(&qn.ID, &qn.RevisionID, &qn.Key, &qn.Position, &qType, &qn.PromptEN, &qn.PromptAM,
			&qn.HelpEN, &qn.HelpAM, &qn.Required, &qn.IsPublic, &qn.MaxLength, &options); err != nil {
			return nil, fmt.Errorf("load questions: %w", err)
		}
		qn.Type = questionnaire.QuestionType(qType)
		if err := json.Unmarshal(options, &qn.Options); err != nil {
			return nil, fmt.Errorf("load questions: decode options for %q: %w", qn.Key, err)
		}
		if len(qn.Options) == 0 {
			qn.Options = nil
		}
		questions = append(questions, qn)
	}
	return questions, rows.Err()
}

func scanQuestionnaireRevision(row scannable) (questionnaire.Revision, error) {
	var rev questionnaire.Revision
	var status, kind string
	err := row.Scan(&rev.ID, &rev.Number, &status, &kind, &rev.Note, &rev.CreatedBy, &rev.PublishedAt, &rev.CreatedAt, &rev.UpdatedAt)
	if err != nil {
		return questionnaire.Revision{}, err
	}
	rev.Status = questionnaire.RevisionStatus(status)
	rev.ChangeKind = questionnaire.ChangeKind(kind)
	return rev, nil
}
