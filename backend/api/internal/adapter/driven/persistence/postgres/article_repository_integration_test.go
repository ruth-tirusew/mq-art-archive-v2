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

	byID, err := repo.GetByID(ctx, draft.ID)
	assist.NoError(t, err)
	assist.Equal(t, "Integration Draft", byID.Title)

	draftStatus := content.ArticleStatusDraft
	adminList, err := repo.ListAdmin(ctx, &draftStatus, 50, 0)
	assist.NoError(t, err)
	assist.GreaterOrEqual(t, len(adminList), 1)

	byID.Title = "Integration Draft Updated"
	byID.Body = "updated body"
	byID.Category = "Legal"
	byID.UpdatedAt = time.Now().UTC()
	updated, err := repo.Update(ctx, *byID)
	assist.NoError(t, err)
	assist.Equal(t, "Integration Draft Updated", updated.Title)
	assist.Equal(t, "Legal", updated.Category)
}
