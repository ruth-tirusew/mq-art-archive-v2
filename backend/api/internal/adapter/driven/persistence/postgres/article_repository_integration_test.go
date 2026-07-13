//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	"github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

func TestArticleRepository_integration(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewArticleRepository(pool)
	ctx := context.Background()

	authorID := uuid.New()
	draft, err := repo.Create(ctx, content.Article{
		ID:        uuid.New(),
		Slug:      "integration-draft",
		Title:     "Integration Draft",
		Body:      "body",
		AuthorID:  authorID,
		Status:    content.ArticleStatusDraft,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	assist.NoError(t, err)
	assist.Equal(t, content.ArticleStatusDraft, draft.Status)

	_, err = repo.GetBySlug(ctx, "integration-draft")
	assist.ErrorIs(t, err, postgres.ErrNotFound)

	publishedID := uuid.New()
	_, err = repo.Create(ctx, content.Article{
		ID:        publishedID,
		Slug:      "integration-published",
		Title:     "Integration Published",
		Body:      "published body",
		AuthorID:  authorID,
		Status:    content.ArticleStatusPublished,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	assist.NoError(t, err)

	got, err := repo.GetBySlug(ctx, "integration-published")
	assist.NoError(t, err)
	assist.Equal(t, "Integration Published", got.Title)

	list, err := repo.ListPublished(ctx, content.PublicListFilter())
	assist.NoError(t, err)
	assist.GreaterOrEqual(t, len(list), 1)
}
