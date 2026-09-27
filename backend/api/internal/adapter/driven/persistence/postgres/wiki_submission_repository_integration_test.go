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

func TestWikiSubmissionRepository_ListPendingForOwner_integration(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	ctx := context.Background()

	articles := postgres.NewArticleRepository(pool)
	submissions := postgres.NewWikiSubmissionRepository(pool)

	owner := uuid.New()
	other := uuid.New()
	submitter := uuid.New()
	now := time.Now().UTC()
	for _, id := range []uuid.UUID{owner, other, submitter} {
		_, err := pool.Exec(ctx, `
			INSERT INTO users (id, email, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
		`, id, id.String()+"@example.com", string(identity.RoleArtist), now, now)
		assist.NoError(t, err)
	}

	ownedArticle, err := articles.Create(ctx, content.Article{
		ID: uuid.New(), Slug: "owned-article", Title: "Owned", Body: "body",
		AuthorID: owner, Status: content.ArticleStatusPublished, Version: 1,
		CreatedAt: now, UpdatedAt: now,
	})
	assist.NoError(t, err)

	otherArticle, err := articles.Create(ctx, content.Article{
		ID: uuid.New(), Slug: "other-article", Title: "Other", Body: "body",
		AuthorID: other, Status: content.ArticleStatusPublished, Version: 1,
		CreatedAt: now, UpdatedAt: now,
	})
	assist.NoError(t, err)

	// Pending edit against the owner's article — should be returned.
	_, err = submissions.Create(ctx, wiki.Submission{
		ID: uuid.New(), SubmitterID: submitter, ArticleID: &ownedArticle.ID,
		Title: "Edit to owned", Body: "proposed", Kind: wiki.KindEdit,
		Status: wiki.StatusPending, CreatedAt: now, UpdatedAt: now,
	})
	assist.NoError(t, err)

	// Pending edit against someone else's article — must not leak into owner's queue.
	_, err = submissions.Create(ctx, wiki.Submission{
		ID: uuid.New(), SubmitterID: submitter, ArticleID: &otherArticle.ID,
		Title: "Edit to other", Body: "proposed", Kind: wiki.KindEdit,
		Status: wiki.StatusPending, CreatedAt: now, UpdatedAt: now,
	})
	assist.NoError(t, err)

	// A new-article submission has no ArticleID, so it can never appear in an owner queue.
	_, err = submissions.Create(ctx, wiki.Submission{
		ID: uuid.New(), SubmitterID: submitter, ArticleID: nil,
		Title: "Brand new", Body: "proposed", Kind: wiki.KindNew,
		Status: wiki.StatusPending, CreatedAt: now, UpdatedAt: now,
	})
	assist.NoError(t, err)

	got, err := submissions.ListPendingForOwner(ctx, owner)
	assist.NoError(t, err)
	assist.Equal(t, 1, len(got))
	assist.Equal(t, "Edit to owned", got[0].Title)
}
