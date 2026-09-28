//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	"github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/domain/wiki"
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

func TestArticleRepository_UpdateForApproval_integration(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewArticleRepository(pool)
	ctx := context.Background()

	authorID := uuid.New()
	article, err := repo.Create(ctx, content.Article{
		ID:        uuid.New(),
		Slug:      "approval-flow",
		Title:     "Approval Flow",
		Body:      "original body",
		AuthorID:  authorID,
		Status:    content.ArticleStatusPublished,
		Version:   1,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	assist.NoError(t, err)

	submissions := postgres.NewWikiSubmissionRepository(pool)
	editorID := uuid.New()
	now := time.Now().UTC()
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, email, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, editorID, editorID.String()+"@example.com", string(identity.RoleArtist), now, now)
	assist.NoError(t, err)
	submission, err := submissions.Create(ctx, wiki.Submission{
		ID: uuid.New(), SubmitterID: editorID, ArticleID: &article.ID,
		Title: article.Title, Body: "approved body", Kind: wiki.KindEdit,
		Status: wiki.StatusPending, CreatedAt: now, UpdatedAt: now,
	})
	assist.NoError(t, err)
	submissionID := submission.ID

	updatedArticle := *article
	updatedArticle.Body = "approved body"

	revision := content.ArticleRevision{
		ArticleID: article.ID, Version: article.Version, EditorID: editorID,
		Title: article.Title, Body: article.Body, Slug: article.Slug, Status: article.Status,
		SubmissionID: &submissionID,
	}

	got, err := repo.UpdateForApproval(ctx, updatedArticle, article.Version, revision)
	assist.NoError(t, err)
	assist.Equal(t, 2, got.Version)

	fromDB, err := repo.GetByID(ctx, article.ID)
	assist.NoError(t, err)
	assist.Equal(t, "approved body", fromDB.Body)

	revs, err := repo.ListRevisions(ctx, article.ID, 10, 0)
	assist.NoError(t, err)
	assist.Equal(t, 1, len(revs))
	assist.Equal(t, "original body", revs[0].Body)

	// A stale expectedVersion (the article already moved to version 2 above) must be
	// rejected rather than silently overwriting the newer state.
	_, err = repo.UpdateForApproval(ctx, updatedArticle, article.Version, revision)
	assist.ErrorIs(t, err, postgres.ErrConflict)
}
