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
}
