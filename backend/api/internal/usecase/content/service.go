package content

import (
	"context"
	"time"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

type Service struct {
	articles outbound.ArticleRepository
}

func NewService(articles outbound.ArticleRepository) inbound.ContentService {
	return &Service{articles: articles}
}

func (s *Service) ListPublished(ctx context.Context, filter domain.ListFilter) ([]domain.Article, error) {
	return s.articles.ListPublished(ctx, filter)
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*domain.Article, error) {
	return s.articles.GetBySlug(ctx, slug)
}

func (s *Service) CreateDraft(ctx context.Context, authorID uuid.UUID, title, body string) (*domain.Article, error) {
	if err := requireTitle(title); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	article := domain.Article{
		ID:        uuid.New(),
		Slug:      slugify(title),
		Title:     title,
		Body:      body,
		AuthorID:  authorID,
		Status:    domain.ArticleStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return s.articles.Create(ctx, article)
}
