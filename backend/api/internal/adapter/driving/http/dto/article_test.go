package dto

import (
	"testing"
	"time"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/testutil/assist"
)

func TestToArticleResponse(t *testing.T) {
	now := time.Now().UTC()
	id := uuid.New()
	authorID := uuid.New()

	article := domain.Article{
		ID:        id,
		Slug:      "test",
		Title:     "Test",
		Body:      "Body",
		AuthorID:  authorID,
		Status:    domain.ArticleStatusPublished,
		CreatedAt: now,
		UpdatedAt: now,
	}

	got := ToArticleResponse(article)
	assist.Equal(t, id, got.ID)
	assist.Equal(t, "published", got.Status)
}

func TestToArticleResponses_emptySlice(t *testing.T) {
	got := ToArticleResponses(nil)
	assist.Len(t, 0, len(got))
}
