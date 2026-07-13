package content

import (
	"context"
	"testing"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/testutil/assist"
)

type mockArticleRepo struct {
	listPublished func(ctx context.Context, filter domain.ListFilter) ([]domain.Article, error)
	getBySlug     func(ctx context.Context, slug string) (*domain.Article, error)
	create        func(ctx context.Context, article domain.Article) (*domain.Article, error)
}

func (m *mockArticleRepo) ListPublished(ctx context.Context, filter domain.ListFilter) ([]domain.Article, error) {
	return m.listPublished(ctx, filter)
}

func (m *mockArticleRepo) GetBySlug(ctx context.Context, slug string) (*domain.Article, error) {
	return m.getBySlug(ctx, slug)
}

func (m *mockArticleRepo) Create(ctx context.Context, article domain.Article) (*domain.Article, error) {
	return m.create(ctx, article)
}

func TestService_CreateDraft_requiresTitle(t *testing.T) {
	svc := NewService(&mockArticleRepo{})

	_, err := svc.CreateDraft(context.Background(), uuid.New(), "  ", "body")
	assist.Error(t, err)
	assist.Contains(t, err.Error(), "title is required")
}

func TestService_CreateDraft_delegatesToRepository(t *testing.T) {
	authorID := uuid.New()
	var captured domain.Article

	repo := &mockArticleRepo{
		create: func(ctx context.Context, article domain.Article) (*domain.Article, error) {
			captured = article
			return &article, nil
		},
	}
	svc := NewService(repo)

	got, err := svc.CreateDraft(context.Background(), authorID, "How to Paint", "Use brushes")
	assist.NoError(t, err)

	assist.Equal(t, authorID, captured.AuthorID)
	assist.Equal(t, "How to Paint", captured.Title)
	assist.Equal(t, domain.ArticleStatusDraft, captured.Status)
	assist.Equal(t, "how-to-paint", captured.Slug)
	assist.Equal(t, got.ID, captured.ID)
}

func TestService_GetBySlug_delegatesToRepository(t *testing.T) {
	expected := &domain.Article{Slug: "test-slug", Title: "Test"}
	repo := &mockArticleRepo{
		getBySlug: func(ctx context.Context, slug string) (*domain.Article, error) {
			assist.Equal(t, "test-slug", slug)
			return expected, nil
		},
	}
	svc := NewService(repo)

	got, err := svc.GetBySlug(context.Background(), "test-slug")
	assist.NoError(t, err)
	assist.Equal(t, expected, got)
}
