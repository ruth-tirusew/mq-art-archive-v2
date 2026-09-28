package outbound

import (
	"context"

	"github.com/google/uuid"
)

type FavoriteRepository interface {
	Add(ctx context.Context, userID, articleID uuid.UUID) error
	Remove(ctx context.Context, userID, articleID uuid.UUID) error
	IsFavorited(ctx context.Context, userID, articleID uuid.UUID) (bool, error)
	CountForArticle(ctx context.Context, articleID uuid.UUID) (int, error)
	ListArticleIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}
