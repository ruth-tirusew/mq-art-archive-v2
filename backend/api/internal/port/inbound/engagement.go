package inbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/content"
)

type EngagementService interface {
	// ToggleFavorite adds the caller's favorite if it doesn't exist, or removes it if it
	// does. Returns the resulting favorited state and the article's new total count.
	// Fails with apperrors.ErrNotFound if the article doesn't exist.
	ToggleFavorite(ctx context.Context, userID, articleID uuid.UUID) (favorited bool, count int, err error)
	// FavoriteStatus returns an article's total favorite count, and whether viewerID (may
	// be uuid.Nil for an unauthenticated viewer, in which case favorited is always false)
	// has favorited it.
	FavoriteStatus(ctx context.Context, viewerID, articleID uuid.UUID) (favorited bool, count int, err error)
	// ListMyFavorites returns the articles the caller has favorited, most recent first.
	ListMyFavorites(ctx context.Context, userID uuid.UUID) ([]content.Article, error)
}
