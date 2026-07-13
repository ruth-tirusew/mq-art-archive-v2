package outbound

import (
	"context"

	"github.com/mq/api/internal/domain/content"
)

type ArticleRepository interface {
	ListPublished(ctx context.Context, filter content.ListFilter) ([]content.Article, error)
	GetBySlug(ctx context.Context, slug string) (*content.Article, error)
	Create(ctx context.Context, article content.Article) (*content.Article, error)
}
