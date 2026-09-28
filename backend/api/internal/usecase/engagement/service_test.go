package engagement

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/content"
)

type fakeFavoriteRepo struct {
	favorited map[uuid.UUID]map[uuid.UUID]bool // userID -> articleID -> favorited
}

func newFakeFavoriteRepo() *fakeFavoriteRepo {
	return &fakeFavoriteRepo{favorited: map[uuid.UUID]map[uuid.UUID]bool{}}
}

func (f *fakeFavoriteRepo) Add(_ context.Context, userID, articleID uuid.UUID) error {
	if f.favorited[userID] == nil {
		f.favorited[userID] = map[uuid.UUID]bool{}
	}
	f.favorited[userID][articleID] = true
	return nil
}

func (f *fakeFavoriteRepo) Remove(_ context.Context, userID, articleID uuid.UUID) error {
	delete(f.favorited[userID], articleID)
	return nil
}

func (f *fakeFavoriteRepo) IsFavorited(_ context.Context, userID, articleID uuid.UUID) (bool, error) {
	return f.favorited[userID][articleID], nil
}

func (f *fakeFavoriteRepo) CountForArticle(_ context.Context, articleID uuid.UUID) (int, error) {
	count := 0
	for _, byArticle := range f.favorited {
		if byArticle[articleID] {
			count++
		}
	}
	return count, nil
}

func (f *fakeFavoriteRepo) ListArticleIDsByUser(_ context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	ids := []uuid.UUID{}
	for articleID, ok := range f.favorited[userID] {
		if ok {
			ids = append(ids, articleID)
		}
	}
	return ids, nil
}

// fakeContent is a minimal stand-in for inbound.ContentService covering only AdminGet,
// which is all this usecase calls.
type fakeContent struct {
	articles map[uuid.UUID]content.Article
}

func newFakeContent(articles ...content.Article) *fakeContent {
	m := map[uuid.UUID]content.Article{}
	for _, a := range articles {
		m[a.ID] = a
	}
	return &fakeContent{articles: m}
}

func (*fakeContent) ListPublished(context.Context, content.ListFilter) ([]content.Article, error) {
	return nil, nil
}
func (*fakeContent) GetBySlug(context.Context, string) (*content.Article, error) { return nil, nil }
func (*fakeContent) CreateDraft(context.Context, uuid.UUID, string, string) (*content.Article, error) {
	return nil, nil
}
func (*fakeContent) AdminList(context.Context, *content.ArticleStatus, int, int) ([]content.Article, error) {
	return nil, nil
}
func (f *fakeContent) AdminGet(_ context.Context, id uuid.UUID) (*content.Article, error) {
	a, ok := f.articles[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return &a, nil
}
func (*fakeContent) AdminCreate(context.Context, uuid.UUID, content.ArticleWrite) (*content.Article, error) {
	return nil, nil
}
func (*fakeContent) AdminUpdate(context.Context, uuid.UUID, uuid.UUID, content.ArticleWrite) (*content.Article, error) {
	return nil, nil
}
func (*fakeContent) ApproveEdit(context.Context, uuid.UUID, uuid.UUID, int, uuid.UUID, content.ArticleWrite) (*content.Article, error) {
	return nil, nil
}
func (*fakeContent) AdminSetStatus(context.Context, uuid.UUID, *content.ArticleStatus, *bool) (*content.Article, error) {
	return nil, nil
}
func (*fakeContent) AdminListRevisions(context.Context, uuid.UUID, int, int) ([]content.ArticleRevision, error) {
	return nil, nil
}
func (*fakeContent) AdminGetRevision(context.Context, uuid.UUID, int) (*content.ArticleRevision, error) {
	return nil, nil
}
func (*fakeContent) AdminRestoreRevision(context.Context, uuid.UUID, int, uuid.UUID) (*content.Article, error) {
	return nil, nil
}

func TestToggleFavorite_addsThenRemoves(t *testing.T) {
	articleID := uuid.New()
	userID := uuid.New()
	repo := newFakeFavoriteRepo()
	svc := NewService(repo, newFakeContent(content.Article{ID: articleID}))

	favorited, count, err := svc.ToggleFavorite(context.Background(), userID, articleID)
	if err != nil {
		t.Fatal(err)
	}
	if !favorited || count != 1 {
		t.Fatalf("expected favorited=true count=1, got favorited=%v count=%d", favorited, count)
	}

	favorited, count, err = svc.ToggleFavorite(context.Background(), userID, articleID)
	if err != nil {
		t.Fatal(err)
	}
	if favorited || count != 0 {
		t.Fatalf("expected favorited=false count=0 after toggling off, got favorited=%v count=%d", favorited, count)
	}
}

func TestToggleFavorite_unknownArticle_returnsNotFound(t *testing.T) {
	repo := newFakeFavoriteRepo()
	svc := NewService(repo, newFakeContent())

	_, _, err := svc.ToggleFavorite(context.Background(), uuid.New(), uuid.New())
	if err != apperrors.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestFavoriteStatus_unauthenticatedViewer_seesCountNotOwnState(t *testing.T) {
	articleID := uuid.New()
	repo := newFakeFavoriteRepo()
	_ = repo.Add(context.Background(), uuid.New(), articleID)
	svc := NewService(repo, newFakeContent(content.Article{ID: articleID}))

	favorited, count, err := svc.FavoriteStatus(context.Background(), uuid.Nil, articleID)
	if err != nil {
		t.Fatal(err)
	}
	if favorited || count != 1 {
		t.Fatalf("expected favorited=false (no viewer) count=1, got favorited=%v count=%d", favorited, count)
	}
}

func TestFavoriteStatus_authenticatedViewer_seesOwnState(t *testing.T) {
	articleID := uuid.New()
	userID := uuid.New()
	repo := newFakeFavoriteRepo()
	_ = repo.Add(context.Background(), userID, articleID)
	svc := NewService(repo, newFakeContent(content.Article{ID: articleID}))

	favorited, count, err := svc.FavoriteStatus(context.Background(), userID, articleID)
	if err != nil {
		t.Fatal(err)
	}
	if !favorited || count != 1 {
		t.Fatalf("expected favorited=true count=1, got favorited=%v count=%d", favorited, count)
	}
}

func TestListMyFavorites_returnsArticlesSkippingMissing(t *testing.T) {
	userID := uuid.New()
	kept := content.Article{ID: uuid.New(), Title: "Kept"}
	deletedID := uuid.New()
	repo := newFakeFavoriteRepo()
	_ = repo.Add(context.Background(), userID, kept.ID)
	_ = repo.Add(context.Background(), userID, deletedID) // no longer exists in content service
	svc := NewService(repo, newFakeContent(kept))

	articles, err := svc.ListMyFavorites(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 1 || articles[0].ID != kept.ID {
		t.Fatalf("expected only the surviving article, got %#v", articles)
	}
}
