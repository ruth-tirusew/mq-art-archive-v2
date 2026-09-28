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
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

func TestFavoriteRepository_integration(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	ctx := context.Background()

	articles := postgres.NewArticleRepository(pool)
	favorites := postgres.NewFavoriteRepository(pool)

	authorID := uuid.New()
	now := time.Now().UTC()
	article, err := articles.Create(ctx, content.Article{
		ID: uuid.New(), Slug: "fav-test", Title: "Fav Test", Body: "body",
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

	// Starts unfavorited.
	favorited, err := favorites.IsFavorited(ctx, userA, article.ID)
	assist.NoError(t, err)
	assist.Equal(t, false, favorited)

	// Add is idempotent.
	assist.NoError(t, favorites.Add(ctx, userA, article.ID))
	assist.NoError(t, favorites.Add(ctx, userA, article.ID))
	favorited, err = favorites.IsFavorited(ctx, userA, article.ID)
	assist.NoError(t, err)
	assist.Equal(t, true, favorited)

	assist.NoError(t, favorites.Add(ctx, userB, article.ID))

	count, err := favorites.CountForArticle(ctx, article.ID)
	assist.NoError(t, err)
	assist.Equal(t, 2, count)

	ids, err := favorites.ListArticleIDsByUser(ctx, userA)
	assist.NoError(t, err)
	assist.Equal(t, 1, len(ids))
	assist.Equal(t, article.ID, ids[0])

	// Removing one user's favorite doesn't affect the other's.
	assist.NoError(t, favorites.Remove(ctx, userA, article.ID))
	count, err = favorites.CountForArticle(ctx, article.ID)
	assist.NoError(t, err)
	assist.Equal(t, 1, count)

	favorited, err = favorites.IsFavorited(ctx, userB, article.ID)
	assist.NoError(t, err)
	assist.Equal(t, true, favorited)
}
