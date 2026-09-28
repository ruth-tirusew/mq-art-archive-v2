package engagement

import (
	"context"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

type Service struct {
	favorites outbound.FavoriteRepository
	content   inbound.ContentService
}

func NewService(favorites outbound.FavoriteRepository, content inbound.ContentService) inbound.EngagementService {
	return &Service{favorites: favorites, content: content}
}

func (s *Service) ToggleFavorite(ctx context.Context, userID, articleID uuid.UUID) (bool, int, error) {
	if _, err := s.content.AdminGet(ctx, articleID); err != nil {
		return false, 0, err
	}

	already, err := s.favorites.IsFavorited(ctx, userID, articleID)
	if err != nil {
		return false, 0, err
	}

	if already {
		if err := s.favorites.Remove(ctx, userID, articleID); err != nil {
			return false, 0, err
		}
	} else if err := s.favorites.Add(ctx, userID, articleID); err != nil {
		return false, 0, err
	}

	count, err := s.favorites.CountForArticle(ctx, articleID)
	if err != nil {
		return false, 0, err
	}
	return !already, count, nil
}

func (s *Service) FavoriteStatus(ctx context.Context, viewerID, articleID uuid.UUID) (bool, int, error) {
	count, err := s.favorites.CountForArticle(ctx, articleID)
	if err != nil {
		return false, 0, err
	}
	if viewerID == uuid.Nil {
		return false, count, nil
	}
	favorited, err := s.favorites.IsFavorited(ctx, viewerID, articleID)
	if err != nil {
		return false, 0, err
	}
	return favorited, count, nil
}

func (s *Service) ListMyFavorites(ctx context.Context, userID uuid.UUID) ([]domain.Article, error) {
	ids, err := s.favorites.ListArticleIDsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	articles := make([]domain.Article, 0, len(ids))
	for _, id := range ids {
		article, err := s.content.AdminGet(ctx, id)
		if err != nil {
			continue // favorited article was since deleted — skip rather than fail the list
		}
		articles = append(articles, *article)
	}
	return articles, nil
}
