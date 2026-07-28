package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/content"
)

type ArticleRepository interface {
	ListPublished(ctx context.Context, filter content.ListFilter) ([]content.Article, error)
	ListAdmin(ctx context.Context, status *content.ArticleStatus, limit, offset int) ([]content.Article, error)
	GetBySlug(ctx context.Context, slug string) (*content.Article, error)
	GetByID(ctx context.Context, id uuid.UUID) (*content.Article, error)
	Create(ctx context.Context, article content.Article) (*content.Article, error)
	Update(ctx context.Context, article content.Article) (*content.Article, error)
	Search(ctx context.Context, query string, limit int) ([]content.Article, error)
	InsertRevision(ctx context.Context, rev content.ArticleRevision) error
	ListRevisions(ctx context.Context, articleID uuid.UUID, limit, offset int) ([]content.ArticleRevision, error)
	GetRevision(ctx context.Context, articleID uuid.UUID, version int) (*content.ArticleRevision, error)
}
