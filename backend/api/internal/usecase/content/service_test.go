package content

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	domain "github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/testutil/assist"
)

type mockArticleRepo struct {
	listPublished  func(ctx context.Context, filter domain.ListFilter) ([]domain.Article, error)
	listAdmin      func(ctx context.Context, status *domain.ArticleStatus, limit, offset int) ([]domain.Article, error)
	getBySlug      func(ctx context.Context, slug string) (*domain.Article, error)
	getByID        func(ctx context.Context, id uuid.UUID) (*domain.Article, error)
	create         func(ctx context.Context, article domain.Article) (*domain.Article, error)
	update         func(ctx context.Context, article domain.Article) (*domain.Article, error)
	insertRevision func(ctx context.Context, rev domain.ArticleRevision) error
	listRevisions  func(ctx context.Context, articleID uuid.UUID, limit, offset int) ([]domain.ArticleRevision, error)
	getRevision    func(ctx context.Context, articleID uuid.UUID, version int) (*domain.ArticleRevision, error)
}

func (m *mockArticleRepo) ListPublished(ctx context.Context, filter domain.ListFilter) ([]domain.Article, error) {
	return m.listPublished(ctx, filter)
}

func (m *mockArticleRepo) ListAdmin(ctx context.Context, status *domain.ArticleStatus, limit, offset int) ([]domain.Article, error) {
	if m.listAdmin != nil {
		return m.listAdmin(ctx, status, limit, offset)
	}
	return nil, nil
}

func (m *mockArticleRepo) GetBySlug(ctx context.Context, slug string) (*domain.Article, error) {
	return m.getBySlug(ctx, slug)
}

func (m *mockArticleRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Article, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return nil, apperrors.ErrNotFound
}

func (m *mockArticleRepo) Search(ctx context.Context, query string, limit int) ([]domain.Article, error) {
	return nil, nil
}

func (m *mockArticleRepo) Create(ctx context.Context, article domain.Article) (*domain.Article, error) {
	return m.create(ctx, article)
}

func (m *mockArticleRepo) Update(ctx context.Context, article domain.Article) (*domain.Article, error) {
	if m.update != nil {
		return m.update(ctx, article)
	}
	return &article, nil
}

func (m *mockArticleRepo) InsertRevision(ctx context.Context, rev domain.ArticleRevision) error {
	if m.insertRevision != nil {
		return m.insertRevision(ctx, rev)
	}
	return nil
}

func (m *mockArticleRepo) ListRevisions(ctx context.Context, articleID uuid.UUID, limit, offset int) ([]domain.ArticleRevision, error) {
	if m.listRevisions != nil {
		return m.listRevisions(ctx, articleID, limit, offset)
	}
	return nil, nil
}

func (m *mockArticleRepo) GetRevision(ctx context.Context, articleID uuid.UUID, version int) (*domain.ArticleRevision, error) {
	if m.getRevision != nil {
		return m.getRevision(ctx, articleID, version)
	}
	return nil, apperrors.ErrNotFound
}

func TestService_CreateDraft_requiresTitle(t *testing.T) {
	svc := NewService(&mockArticleRepo{})

	_, err := svc.CreateDraft(context.Background(), uuid.New(), "  ", "body")
	assist.Error(t, err)
	assist.ErrorIs(t, err, apperrors.ErrValidation)
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
	assist.Equal(t, "General", captured.Category)
	assist.Equal(t, "Beginner", captured.Difficulty)
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

func TestService_AdminCreate_withMetadata(t *testing.T) {
	authorID := uuid.New()
	status := domain.ArticleStatusPublished
	var captured domain.Article

	repo := &mockArticleRepo{
		create: func(ctx context.Context, article domain.Article) (*domain.Article, error) {
			captured = article
			return &article, nil
		},
	}
	svc := NewService(repo)

	_, err := svc.AdminCreate(context.Background(), authorID, domain.ArticleWrite{
		Title:      "Legal Tips",
		Body:       "File the paperwork early.",
		Category:   "Legal",
		Excerpt:    "Short guide",
		Difficulty: "Intermediate",
		Verified:   true,
		Status:     &status,
	})
	assist.NoError(t, err)
	assist.Equal(t, "Legal", captured.Category)
	assist.Equal(t, 1, captured.ReadingTime)
	assist.Equal(t, true, captured.Verified)
	assist.Equal(t, domain.ArticleStatusPublished, captured.Status)
}

func TestService_AdminUpdate_reslugifiesDraft(t *testing.T) {
	id := uuid.New()
	editorID := uuid.New()
	var captured domain.Article
	var snap domain.ArticleRevision

	repo := &mockArticleRepo{
		getByID: func(ctx context.Context, gotID uuid.UUID) (*domain.Article, error) {
			assist.Equal(t, id, gotID)
			return &domain.Article{
				ID:      id,
				Slug:    "old-title",
				Title:   "Old Title",
				Status:  domain.ArticleStatusDraft,
				Version: 1,
			}, nil
		},
		insertRevision: func(ctx context.Context, rev domain.ArticleRevision) error {
			snap = rev
			return nil
		},
		update: func(ctx context.Context, article domain.Article) (*domain.Article, error) {
			captured = article
			return &article, nil
		},
	}
	svc := NewService(repo)

	_, err := svc.AdminUpdate(context.Background(), id, editorID, domain.ArticleWrite{
		Title:    "New Title",
		Body:     "Updated body",
		Category: "Materials",
	})
	assist.NoError(t, err)
	assist.Equal(t, "new-title", captured.Slug)
	assist.Equal(t, "New Title", captured.Title)
	assist.Equal(t, "Materials", captured.Category)
	assist.Equal(t, 2, captured.Version)
	assist.Equal(t, 1, snap.Version)
	assist.Equal(t, editorID, snap.EditorID)
}

func TestService_AdminUpdate_keepsSlugWhenPublished(t *testing.T) {
	id := uuid.New()
	editorID := uuid.New()
	var captured domain.Article

	repo := &mockArticleRepo{
		getByID: func(ctx context.Context, gotID uuid.UUID) (*domain.Article, error) {
			return &domain.Article{
				ID:      id,
				Slug:    "stable-slug",
				Title:   "Old Title",
				Status:  domain.ArticleStatusPublished,
				Version: 3,
			}, nil
		},
		update: func(ctx context.Context, article domain.Article) (*domain.Article, error) {
			captured = article
			return &article, nil
		},
	}
	svc := NewService(repo)

	_, err := svc.AdminUpdate(context.Background(), id, editorID, domain.ArticleWrite{
		Title: "Renamed Published",
		Body:  "Body",
	})
	assist.NoError(t, err)
	assist.Equal(t, "stable-slug", captured.Slug)
	assist.Equal(t, 4, captured.Version)
}

func TestService_AdminRestoreRevision(t *testing.T) {
	id := uuid.New()
	editorID := uuid.New()
	var captured domain.Article

	repo := &mockArticleRepo{
		getByID: func(ctx context.Context, gotID uuid.UUID) (*domain.Article, error) {
			return &domain.Article{
				ID: id, Title: "Current", Body: "now", Status: domain.ArticleStatusDraft, Version: 2, Slug: "current",
			}, nil
		},
		getRevision: func(ctx context.Context, articleID uuid.UUID, version int) (*domain.ArticleRevision, error) {
			assist.Equal(t, 1, version)
			return &domain.ArticleRevision{
				ArticleID: id, Version: 1, Title: "Original", Body: "then", Slug: "original",
				Status: domain.ArticleStatusDraft, ReadingTime: 2,
			}, nil
		},
		update: func(ctx context.Context, article domain.Article) (*domain.Article, error) {
			captured = article
			return &article, nil
		},
	}
	svc := NewService(repo)

	got, err := svc.AdminRestoreRevision(context.Background(), id, 1, editorID)
	assist.NoError(t, err)
	assist.Equal(t, "Original", got.Title)
	assist.Equal(t, "then", captured.Body)
	assist.Equal(t, 3, captured.Version)
}

func TestService_AdminGetRevision_notFound(t *testing.T) {
	id := uuid.New()
	repo := &mockArticleRepo{
		getByID: func(ctx context.Context, gotID uuid.UUID) (*domain.Article, error) {
			return &domain.Article{ID: id, Version: 1}, nil
		},
	}
	svc := NewService(repo)
	_, err := svc.AdminGetRevision(context.Background(), id, 99)
	assist.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestService_AdminSetStatus(t *testing.T) {
	id := uuid.New()
	published := domain.ArticleStatusPublished
	verified := true
	var captured domain.Article

	repo := &mockArticleRepo{
		getByID: func(ctx context.Context, gotID uuid.UUID) (*domain.Article, error) {
			return &domain.Article{ID: id, Status: domain.ArticleStatusDraft, Verified: false}, nil
		},
		update: func(ctx context.Context, article domain.Article) (*domain.Article, error) {
			captured = article
			return &article, nil
		},
	}
	svc := NewService(repo)

	_, err := svc.AdminSetStatus(context.Background(), id, &published, &verified)
	assist.NoError(t, err)
	assist.Equal(t, domain.ArticleStatusPublished, captured.Status)
	assist.Equal(t, true, captured.Verified)
}

func TestService_AdminSetStatus_invalidStatus(t *testing.T) {
	id := uuid.New()
	bad := domain.ArticleStatus("nope")
	repo := &mockArticleRepo{
		getByID: func(ctx context.Context, gotID uuid.UUID) (*domain.Article, error) {
			return &domain.Article{ID: id, Status: domain.ArticleStatusDraft}, nil
		},
	}
	svc := NewService(repo)

	_, err := svc.AdminSetStatus(context.Background(), id, &bad, nil)
	assist.Error(t, err)
	assist.ErrorIs(t, err, apperrors.ErrValidation)
}
