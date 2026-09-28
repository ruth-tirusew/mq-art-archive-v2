package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/engagement"
	"github.com/mq/api/internal/port/outbound"
)

type CommentRepository struct{ pool *Pool }

func NewCommentRepository(pool *Pool) outbound.CommentRepository {
	return &CommentRepository{pool: pool}
}

const commentColumns = `id,article_id,user_id,body,is_general,quoted_text,text_offset_start,text_offset_end,article_version_at_anchor,created_at`

// Anchor columns are nullable in the schema (see migration 00037) but a general comment
// still writes its Anchor zero value ("" / 0 / 0 / 0) rather than SQL NULL — the
// is_general flag is what actually discriminates, so there's no need for nullable Go
// types on the read side either.
func (r *CommentRepository) Create(ctx context.Context, c engagement.Comment) (*engagement.Comment, error) {
	return scanComment(r.pool.QueryRow(ctx, `
		INSERT INTO article_comments (id,article_id,user_id,body,is_general,quoted_text,text_offset_start,text_offset_end,article_version_at_anchor,created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING `+commentColumns,
		c.ID, c.ArticleID, c.UserID, c.Body, c.IsGeneral, c.QuotedText, c.TextOffsetStart, c.TextOffsetEnd, c.ArticleVersionAtAnchor, c.CreatedAt))
}

func (r *CommentRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM article_comments WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CommentRepository) ListByArticle(ctx context.Context, articleID uuid.UUID) ([]engagement.Comment, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+commentColumns+` FROM article_comments WHERE article_id=$1 ORDER BY created_at`, articleID)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()

	out := []engagement.Comment{}
	for rows.Next() {
		c, err := scanComment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

func scanComment(row scannable) (*engagement.Comment, error) {
	var c engagement.Comment
	err := row.Scan(&c.ID, &c.ArticleID, &c.UserID, &c.Body, &c.IsGeneral, &c.QuotedText, &c.TextOffsetStart, &c.TextOffsetEnd, &c.ArticleVersionAtAnchor, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan comment: %w", err)
	}
	return &c, nil
}
