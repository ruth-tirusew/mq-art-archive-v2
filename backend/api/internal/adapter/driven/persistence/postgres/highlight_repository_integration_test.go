//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	"github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/domain/engagement"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

func TestHighlightRepository_integration(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	ctx := context.Background()

	articles := postgres.NewArticleRepository(pool)
	highlights := postgres.NewHighlightRepository(pool)

	authorID := uuid.New()
	now := time.Now().UTC()
	article, err := articles.Create(ctx, content.Article{
		ID: uuid.New(), Slug: "highlight-test", Title: "Highlight Test", Body: "one two three",
		AuthorID: authorID, Status: content.ArticleStatusPublished, Version: 1,
		CreatedAt: now, UpdatedAt: now,
	})
	assist.NoError(t, err)

	userA := uuid.New()
	userB := uuid.New()
	for _, id := range []uuid.UUID{userA, userB} {
		_, err := pool.Exec(ctx, `
			INSERT INTO users (id, email, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
		`, id, id.String()+"@example.com", string(identity.RoleArtist), now, now)
		assist.NoError(t, err)
	}

	created, err := highlights.Create(ctx, engagement.Highlight{
		ID: uuid.New(), ArticleID: article.ID, UserID: userA,
		Anchor:    engagement.Anchor{QuotedText: "one", TextOffsetStart: 0, TextOffsetEnd: 3, ArticleVersionAtAnchor: 1},
		CreatedAt: now,
	})
	assist.NoError(t, err)

	_, err = highlights.Create(ctx, engagement.Highlight{
		ID: uuid.New(), ArticleID: article.ID, UserID: userB,
		Anchor:    engagement.Anchor{QuotedText: "two", TextOffsetStart: 4, TextOffsetEnd: 7, ArticleVersionAtAnchor: 1},
		CreatedAt: now,
	})
	assist.NoError(t, err)

	all, err := highlights.ListByArticle(ctx, article.ID)
	assist.NoError(t, err)
	assist.Equal(t, 2, len(all))

	mineA, err := highlights.ListMineByArticle(ctx, article.ID, userA)
	assist.NoError(t, err)
	assist.Equal(t, 1, len(mineA))
	assist.Equal(t, "one", mineA[0].QuotedText)

	// Deleting someone else's highlight is a no-op error, not a silent success.
	err = highlights.Delete(ctx, created.ID, userB)
	assist.ErrorIs(t, err, postgres.ErrNotFound)

	err = highlights.Delete(ctx, created.ID, userA)
	assist.NoError(t, err)

	mineA, err = highlights.ListMineByArticle(ctx, article.ID, userA)
	assist.NoError(t, err)
	assist.Equal(t, 0, len(mineA))
}
