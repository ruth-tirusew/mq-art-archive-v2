package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mq/api/internal/port/outbound"
)

type FavoriteRepository struct{ pool *Pool }

func NewFavoriteRepository(pool *Pool) outbound.FavoriteRepository {
	return &FavoriteRepository{pool: pool}
}

func (r *FavoriteRepository) Add(ctx context.Context, userID, articleID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO article_favorites (user_id, article_id) VALUES ($1, $2)
		ON CONFLICT (user_id, article_id) DO NOTHING
	`, userID, articleID)
	if err != nil {
		return fmt.Errorf("add favorite: %w", err)
	}
	return nil
}

func (r *FavoriteRepository) Remove(ctx context.Context, userID, articleID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM article_favorites WHERE user_id = $1 AND article_id = $2
	`, userID, articleID)
	if err != nil {
		return fmt.Errorf("remove favorite: %w", err)
	}
	return nil
}

func (r *FavoriteRepository) IsFavorited(ctx context.Context, userID, articleID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM article_favorites WHERE user_id = $1 AND article_id = $2)
	`, userID, articleID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check favorite: %w", err)
	}
	return exists, nil
}

func (r *FavoriteRepository) CountForArticle(ctx context.Context, articleID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT count(*) FROM article_favorites WHERE article_id = $1
	`, articleID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count favorites: %w", err)
	}
	return count, nil
}

func (r *FavoriteRepository) ListArticleIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT article_id FROM article_favorites WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list favorite article ids: %w", err)
	}
	defer rows.Close()

	ids := []uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan favorite article id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
