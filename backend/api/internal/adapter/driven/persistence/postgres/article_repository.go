package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/port/outbound"
)

type ArticleRepository struct {
	pool *Pool
}

func NewArticleRepository(pool *Pool) outbound.ArticleRepository {
	return &ArticleRepository{pool: pool}
}

const articleColumns = `id, slug, title, body, category, excerpt, reading_time, difficulty, verified, contributors, status, author_id, created_at, updated_at`

func (r *ArticleRepository) ListPublished(ctx context.Context, filter content.ListFilter) ([]content.Article, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT ` + articleColumns + `
		FROM articles
		WHERE status = $1`
	args := []any{string(content.ArticleStatusPublished)}
	argPos := 2

	if filter.Category != "" && !strings.EqualFold(filter.Category, "all") {
		query += fmt.Sprintf(" AND category ILIKE $%d", argPos)
		args = append(args, filter.Category)
		argPos++
	}
	if filter.Query != "" {
		query += fmt.Sprintf(" AND (title ILIKE $%d OR excerpt ILIKE $%d OR body ILIKE $%d)", argPos, argPos, argPos)
		args = append(args, "%"+filter.Query+"%")
		argPos++
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list published articles: %w", err)
	}
	defer rows.Close()

	var articles []content.Article
	for rows.Next() {
		article, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list published articles: %w", err)
	}
	if articles == nil {
		articles = []content.Article{}
	}
	return articles, nil
}

func (r *ArticleRepository) GetBySlug(ctx context.Context, slug string) (*content.Article, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+articleColumns+`
		FROM articles
		WHERE slug = $1 AND status = $2
	`, slug, string(content.ArticleStatusPublished))

	article, err := scanArticle(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get article by slug: %w", err)
	}
	return &article, nil
}

func (r *ArticleRepository) Create(ctx context.Context, article content.Article) (*content.Article, error) {
	slug := article.Slug
	for attempt := 0; attempt < 3; attempt++ {
		_, err := r.pool.Exec(ctx, `
			INSERT INTO articles (id, slug, title, body, category, excerpt, reading_time, difficulty, verified, contributors, status, author_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		`,
			article.ID,
			slug,
			article.Title,
			article.Body,
			article.Category,
			article.Excerpt,
			article.ReadingTime,
			article.Difficulty,
			article.Verified,
			article.Contributors,
			string(article.Status),
			article.AuthorID,
			article.CreatedAt,
			article.UpdatedAt,
		)
		if err == nil {
			article.Slug = slug
			return &article, nil
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, "slug") {
			slug = fmt.Sprintf("%s-%s", article.Slug, uuid.New().String()[:8])
			continue
		}
		return nil, fmt.Errorf("create article: %w", err)
	}
	return nil, fmt.Errorf("create article: slug conflict")
}

type scannable interface {
	Scan(dest ...any) error
}

func scanArticle(row scannable) (content.Article, error) {
	var article content.Article
	var status string
	err := row.Scan(
		&article.ID,
		&article.Slug,
		&article.Title,
		&article.Body,
		&article.Category,
		&article.Excerpt,
		&article.ReadingTime,
		&article.Difficulty,
		&article.Verified,
		&article.Contributors,
		&status,
		&article.AuthorID,
		&article.CreatedAt,
		&article.UpdatedAt,
	)
	if err != nil {
		return content.Article{}, err
	}
	article.Status = content.ArticleStatus(status)
	return article, nil
}
