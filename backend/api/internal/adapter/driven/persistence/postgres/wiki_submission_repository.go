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

const wikiColumns = `id,submitter_id,article_id,title,body,kind,based_on_version,status,review_notes,reviewed_by,reviewed_at,created_at,updated_at`

// wikiColumnsQualified is wikiColumns with the article_submissions alias used by
// ListPendingForOwner's join against articles (which shares column names like id/title).
const wikiColumnsQualified = `s.id,s.submitter_id,s.article_id,s.title,s.body,s.kind,s.based_on_version,s.status,s.review_notes,s.reviewed_by,s.reviewed_at,s.created_at,s.updated_at`

func (r *WikiSubmissionRepository) Create(ctx context.Context, s wiki.Submission) (*wiki.Submission, error) {
	return scanWiki(r.pool.QueryRow(ctx, `INSERT INTO article_submissions
		(id,submitter_id,article_id,title,body,kind,based_on_version,status,review_notes,reviewed_by,reviewed_at,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING `+wikiColumns,
		s.ID, s.SubmitterID, s.ArticleID, s.Title, s.Body, string(s.Kind), s.BasedOnVersion, string(s.Status), s.ReviewNotes, s.ReviewedBy, s.ReviewedAt, s.CreatedAt, s.UpdatedAt))
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

func (r *WikiSubmissionRepository) ListPendingForOwner(ctx context.Context, ownerID uuid.UUID) ([]wiki.Submission, error) {
	return r.list(ctx, `SELECT `+wikiColumnsQualified+`
		FROM article_submissions s
		JOIN articles a ON a.id = s.article_id
		WHERE s.status='pending' AND a.author_id=$1
		ORDER BY s.created_at`, ownerID)
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
	var kind, status string
	err := row.Scan(&s.ID, &s.SubmitterID, &s.ArticleID, &s.Title, &s.Body, &kind, &s.BasedOnVersion, &status, &s.ReviewNotes, &s.ReviewedBy, &s.ReviewedAt, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan wiki submission: %w", err)
	}
	s.Kind = wiki.Kind(kind)
	s.Status = wiki.Status(status)
	return &s, nil
}
