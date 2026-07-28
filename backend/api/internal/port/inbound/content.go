package inbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/content"
)

type ContentService interface {
	ListPublished(ctx context.Context, filter content.ListFilter) ([]content.Article, error)
	GetBySlug(ctx context.Context, slug string) (*content.Article, error)
	CreateDraft(ctx context.Context, authorID uuid.UUID, title, body string) (*content.Article, error)

	AdminList(ctx context.Context, status *content.ArticleStatus, limit, offset int) ([]content.Article, error)
	AdminGet(ctx context.Context, id uuid.UUID) (*content.Article, error)
	AdminCreate(ctx context.Context, authorID uuid.UUID, write content.ArticleWrite) (*content.Article, error)
	AdminUpdate(ctx context.Context, id, editorID uuid.UUID, write content.ArticleWrite) (*content.Article, error)
	AdminSetStatus(ctx context.Context, id uuid.UUID, status *content.ArticleStatus, verified *bool) (*content.Article, error)
	AdminListRevisions(ctx context.Context, articleID uuid.UUID, limit, offset int) ([]content.ArticleRevision, error)
	AdminGetRevision(ctx context.Context, articleID uuid.UUID, version int) (*content.ArticleRevision, error)
	AdminRestoreRevision(ctx context.Context, articleID uuid.UUID, version int, editorID uuid.UUID) (*content.Article, error)
}
