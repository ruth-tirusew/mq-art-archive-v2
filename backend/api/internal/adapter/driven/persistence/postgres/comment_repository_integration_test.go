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

func TestCommentRepository_integration(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	ctx := context.Background()

	articles := postgres.NewArticleRepository(pool)
	comments := postgres.NewCommentRepository(pool)

	authorID := uuid.New()
	now := time.Now().UTC()
	article, err := articles.Create(ctx, content.Article{
		ID: uuid.New(), Slug: "comment-test", Title: "Comment Test", Body: "one two three",
		AuthorID: authorID, Status: content.ArticleStatusPublished, Version: 1,
		CreatedAt: now, UpdatedAt: now,
	})
	assist.NoError(t, err)

	commenter := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, email, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, commenter, commenter.String()+"@example.com", string(identity.RoleArtist), now, now)
	assist.NoError(t, err)

	anchored, err := comments.Create(ctx, engagement.Comment{
		ID: uuid.New(), ArticleID: article.ID, UserID: commenter, Body: "interesting point",
		Anchor:    engagement.Anchor{QuotedText: "two", TextOffsetStart: 4, TextOffsetEnd: 7, ArticleVersionAtAnchor: 1},
		CreatedAt: now,
	})
	assist.NoError(t, err)
	assist.Equal(t, false, anchored.IsGeneral)
	assist.Equal(t, "two", anchored.QuotedText)

	general, err := comments.Create(ctx, engagement.Comment{
		ID: uuid.New(), ArticleID: article.ID, UserID: commenter, Body: "great overall",
		IsGeneral: true, CreatedAt: now,
	})
	assist.NoError(t, err)
	assist.Equal(t, true, general.IsGeneral)

	all, err := comments.ListByArticle(ctx, article.ID)
	assist.NoError(t, err)
	assist.Equal(t, 2, len(all))

	err = comments.Delete(ctx, anchored.ID, uuid.New())
	assist.ErrorIs(t, err, postgres.ErrNotFound)

	err = comments.Delete(ctx, anchored.ID, commenter)
	assist.NoError(t, err)

	all, err = comments.ListByArticle(ctx, article.ID)
	assist.NoError(t, err)
	assist.Equal(t, 1, len(all))
}
