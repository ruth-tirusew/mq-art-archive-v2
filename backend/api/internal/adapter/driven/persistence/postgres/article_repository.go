package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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

const articleColumns = `id, slug, title, body, category, excerpt, reading_time, difficulty, verified, contributors, status, author_id, version, created_at, updated_at`

const revisionColumns = `id, article_id, version, editor_id, title, body, slug, category, excerpt, reading_time, difficulty, verified, status, created_at`

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

func (r *ArticleRepository) Search(ctx context.Context, query string, limit int) ([]content.Article, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []content.Article{}, nil
	}
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.pool.Query(ctx, `
		SELECT `+articleColumns+`
		FROM articles
		WHERE status = $1
		  AND search_vector @@ plainto_tsquery('english', $2)
		ORDER BY ts_rank(search_vector, plainto_tsquery('english', $2)) DESC, created_at DESC
		LIMIT $3
	`, string(content.ArticleStatusPublished), query, limit)
	if err != nil {
		return nil, fmt.Errorf("search articles: %w", err)
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
		return nil, fmt.Errorf("search articles: %w", err)
	}
	if articles == nil {
		articles = []content.Article{}
	}
	return articles, nil
}

func (r *ArticleRepository) ListAdmin(ctx context.Context, status *content.ArticleStatus, limit, offset int) ([]content.Article, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT ` + articleColumns + `
		FROM articles
		WHERE 1=1`
	args := []any{}
	argPos := 1
	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, string(*status))
		argPos++
	}
	query += " ORDER BY updated_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list admin articles: %w", err)
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
		return nil, fmt.Errorf("list admin articles: %w", err)
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

func (r *ArticleRepository) GetByID(ctx context.Context, id uuid.UUID) (*content.Article, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+articleColumns+`
		FROM articles
		WHERE id = $1
	`, id)

	article, err := scanArticle(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get article by id: %w", err)
	}
	return &article, nil
}

func (r *ArticleRepository) Create(ctx context.Context, article content.Article) (*content.Article, error) {
	if article.Version <= 0 {
		article.Version = 1
	}
	slug := article.Slug
	for attempt := 0; attempt < 3; attempt++ {
		_, err := r.pool.Exec(ctx, `
			INSERT INTO articles (id, slug, title, body, category, excerpt, reading_time, difficulty, verified, contributors, status, author_id, version, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
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
			article.Version,
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

func (r *ArticleRepository) Update(ctx context.Context, article content.Article) (*content.Article, error) {
	slug := article.Slug
	for attempt := 0; attempt < 3; attempt++ {
		tag, err := r.pool.Exec(ctx, `
			UPDATE articles SET
				slug = $2, title = $3, body = $4, category = $5, excerpt = $6,
				reading_time = $7, difficulty = $8, verified = $9, contributors = $10,
				status = $11, version = $12, updated_at = $13
			WHERE id = $1
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
			article.Version,
			article.UpdatedAt,
		)
		if err == nil {
			if tag.RowsAffected() == 0 {
				return nil, ErrNotFound
			}
			article.Slug = slug
			return &article, nil
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, "slug") {
			slug = fmt.Sprintf("%s-%s", article.Slug, uuid.New().String()[:8])
			continue
		}
		return nil, fmt.Errorf("update article: %w", err)
	}
	return nil, fmt.Errorf("update article: slug conflict")
}

func (r *ArticleRepository) InsertRevision(ctx context.Context, rev content.ArticleRevision) error {
	if rev.ID == uuid.Nil {
		rev.ID = uuid.New()
	}
	if rev.CreatedAt.IsZero() {
		rev.CreatedAt = time.Now().UTC()
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO article_revisions (
			id, article_id, version, editor_id, title, body, slug, category, excerpt,
			reading_time, difficulty, verified, status, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`,
		rev.ID, rev.ArticleID, rev.Version, rev.EditorID, rev.Title, rev.Body, rev.Slug,
		rev.Category, rev.Excerpt, rev.ReadingTime, rev.Difficulty, rev.Verified,
		string(rev.Status), rev.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert article revision: %w", err)
	}
	return nil
}

func (r *ArticleRepository) ListRevisions(ctx context.Context, articleID uuid.UUID, limit, offset int) ([]content.ArticleRevision, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
		SELECT `+revisionColumns+`
		FROM article_revisions
		WHERE article_id = $1
		ORDER BY version DESC
		LIMIT $2 OFFSET $3
	`, articleID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list article revisions: %w", err)
	}
	defer rows.Close()

	var out []content.ArticleRevision
	for rows.Next() {
		rev, err := scanRevision(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rev)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list article revisions: %w", err)
	}
	if out == nil {
		out = []content.ArticleRevision{}
	}
	return out, nil
}

func (r *ArticleRepository) GetRevision(ctx context.Context, articleID uuid.UUID, version int) (*content.ArticleRevision, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+revisionColumns+`
		FROM article_revisions
		WHERE article_id = $1 AND version = $2
	`, articleID, version)

	rev, err := scanRevision(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get article revision: %w", err)
	}
	return &rev, nil
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
		&article.Version,
		&article.CreatedAt,
		&article.UpdatedAt,
	)
	if err != nil {
		return content.Article{}, err
	}
	article.Status = content.ArticleStatus(status)
	return article, nil
}

func scanRevision(row scannable) (content.ArticleRevision, error) {
	var rev content.ArticleRevision
	var status string
	err := row.Scan(
		&rev.ID,
		&rev.ArticleID,
		&rev.Version,
		&rev.EditorID,
		&rev.Title,
		&rev.Body,
		&rev.Slug,
		&rev.Category,
		&rev.Excerpt,
		&rev.ReadingTime,
		&rev.Difficulty,
		&rev.Verified,
		&status,
		&rev.CreatedAt,
	)
	if err != nil {
		return content.ArticleRevision{}, err
	}
	rev.Status = content.ArticleStatus(status)
	return rev, nil
}
