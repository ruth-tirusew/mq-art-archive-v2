package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mq/api/internal/domain/wiki"
	"github.com/mq/api/internal/port/outbound"
)

type WikiSubmissionRepository struct{ pool *Pool }

func NewWikiSubmissionRepository(pool *Pool) outbound.WikiSubmissionRepository {
	return &WikiSubmissionRepository{pool: pool}
}

const wikiColumns = `id,submitter_id,article_id,title,body,status,review_notes,reviewed_by,reviewed_at,created_at,updated_at`

func (r *WikiSubmissionRepository) Create(ctx context.Context, s wiki.Submission) (*wiki.Submission, error) {
	return scanWiki(r.pool.QueryRow(ctx, `INSERT INTO article_submissions
		(id,submitter_id,article_id,title,body,status,review_notes,reviewed_by,reviewed_at,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING `+wikiColumns,
		s.ID, s.SubmitterID, s.ArticleID, s.Title, s.Body, string(s.Status), s.ReviewNotes, s.ReviewedBy, s.ReviewedAt, s.CreatedAt, s.UpdatedAt))
}

func (r *WikiSubmissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*wiki.Submission, error) {
	return scanWiki(r.pool.QueryRow(ctx, `SELECT `+wikiColumns+` FROM article_submissions WHERE id=$1`, id))
}

func (r *WikiSubmissionRepository) ListBySubmitter(ctx context.Context, id uuid.UUID) ([]wiki.Submission, error) {
	return r.list(ctx, `SELECT `+wikiColumns+` FROM article_submissions WHERE submitter_id=$1 ORDER BY created_at DESC`, id)
}

func (r *WikiSubmissionRepository) ListPending(ctx context.Context) ([]wiki.Submission, error) {
	return r.list(ctx, `SELECT `+wikiColumns+` FROM article_submissions WHERE status='pending' ORDER BY created_at`)
}

func (r *WikiSubmissionRepository) Update(ctx context.Context, s wiki.Submission) (*wiki.Submission, error) {
	return scanWiki(r.pool.QueryRow(ctx, `UPDATE article_submissions SET article_id=$2,status=$3,review_notes=$4,
		reviewed_by=$5,reviewed_at=$6,updated_at=$7 WHERE id=$1 RETURNING `+wikiColumns,
		s.ID, s.ArticleID, string(s.Status), s.ReviewNotes, s.ReviewedBy, s.ReviewedAt, s.UpdatedAt))
}

func (r *WikiSubmissionRepository) list(ctx context.Context, query string, args ...any) ([]wiki.Submission, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list wiki submissions: %w", err)
	}
	defer rows.Close()
	out := []wiki.Submission{}
	for rows.Next() {
		item, err := scanWiki(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func scanWiki(row scannable) (*wiki.Submission, error) {
	var s wiki.Submission
	var status string
	err := row.Scan(&s.ID, &s.SubmitterID, &s.ArticleID, &s.Title, &s.Body, &status, &s.ReviewNotes, &s.ReviewedBy, &s.ReviewedAt, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan wiki submission: %w", err)
	}
	s.Status = wiki.Status(status)
	return &s, nil
}
