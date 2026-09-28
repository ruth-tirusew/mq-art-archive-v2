package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/engagement"
	"github.com/mq/api/internal/port/outbound"
)

type HighlightRepository struct{ pool *Pool }

func NewHighlightRepository(pool *Pool) outbound.HighlightRepository {
	return &HighlightRepository{pool: pool}
}

const highlightColumns = `id,article_id,user_id,quoted_text,text_offset_start,text_offset_end,article_version_at_anchor,created_at`

func (r *HighlightRepository) Create(ctx context.Context, h engagement.Highlight) (*engagement.Highlight, error) {
	return scanHighlight(r.pool.QueryRow(ctx, `
		INSERT INTO article_highlights (id,article_id,user_id,quoted_text,text_offset_start,text_offset_end,article_version_at_anchor,created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING `+highlightColumns,
		h.ID, h.ArticleID, h.UserID, h.QuotedText, h.TextOffsetStart, h.TextOffsetEnd, h.ArticleVersionAtAnchor, h.CreatedAt))
}

func (r *HighlightRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM article_highlights WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete highlight: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *HighlightRepository) ListByArticle(ctx context.Context, articleID uuid.UUID) ([]engagement.Highlight, error) {
	return r.list(ctx, `SELECT `+highlightColumns+` FROM article_highlights WHERE article_id=$1 ORDER BY created_at`, articleID)
}

func (r *HighlightRepository) ListMineByArticle(ctx context.Context, articleID, userID uuid.UUID) ([]engagement.Highlight, error) {
	return r.list(ctx, `SELECT `+highlightColumns+` FROM article_highlights WHERE article_id=$1 AND user_id=$2 ORDER BY created_at`, articleID, userID)
}

func (r *HighlightRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]engagement.Highlight, error) {
	return r.list(ctx, `SELECT `+highlightColumns+` FROM article_highlights WHERE user_id=$1 ORDER BY created_at DESC`, userID)
}

func (r *HighlightRepository) list(ctx context.Context, query string, args ...any) ([]engagement.Highlight, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list highlights: %w", err)
	}
	defer rows.Close()

	out := []engagement.Highlight{}
	for rows.Next() {
		h, err := scanHighlight(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *h)
	}
	return out, rows.Err()
}

func scanHighlight(row scannable) (*engagement.Highlight, error) {
	var h engagement.Highlight
	err := row.Scan(&h.ID, &h.ArticleID, &h.UserID, &h.QuotedText, &h.TextOffsetStart, &h.TextOffsetEnd, &h.ArticleVersionAtAnchor, &h.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan highlight: %w", err)
	}
	return &h, nil
}
