package engagement

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/domain/engagement"
	"github.com/mq/api/internal/domain/identity"
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
	svc := NewService(repo, nil, nil, newFakeContent(content.Article{ID: articleID}), nil)

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
	svc := NewService(repo, nil, nil, newFakeContent(), nil)

	_, _, err := svc.ToggleFavorite(context.Background(), uuid.New(), uuid.New())
	if err != apperrors.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestFavoriteStatus_unauthenticatedViewer_seesCountNotOwnState(t *testing.T) {
	articleID := uuid.New()
	repo := newFakeFavoriteRepo()
	_ = repo.Add(context.Background(), uuid.New(), articleID)
	svc := NewService(repo, nil, nil, newFakeContent(content.Article{ID: articleID}), nil)

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
	svc := NewService(repo, nil, nil, newFakeContent(content.Article{ID: articleID}), nil)

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
	svc := NewService(repo, nil, nil, newFakeContent(kept), nil)

	articles, err := svc.ListMyFavorites(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 1 || articles[0].ID != kept.ID {
		t.Fatalf("expected only the surviving article, got %#v", articles)
	}
}

type fakeHighlightRepo struct {
	byArticle map[uuid.UUID][]engagement.Highlight
}

func newFakeHighlightRepo() *fakeHighlightRepo {
	return &fakeHighlightRepo{byArticle: map[uuid.UUID][]engagement.Highlight{}}
}

func (f *fakeHighlightRepo) Create(_ context.Context, h engagement.Highlight) (*engagement.Highlight, error) {
	f.byArticle[h.ArticleID] = append(f.byArticle[h.ArticleID], h)
	return &h, nil
}

func (f *fakeHighlightRepo) Delete(_ context.Context, id, userID uuid.UUID) error {
	for articleID, list := range f.byArticle {
		for i, h := range list {
			if h.ID == id && h.UserID == userID {
				f.byArticle[articleID] = append(list[:i], list[i+1:]...)
				return nil
			}
		}
	}
	return apperrors.ErrNotFound
}

func (f *fakeHighlightRepo) ListByArticle(_ context.Context, articleID uuid.UUID) ([]engagement.Highlight, error) {
	return f.byArticle[articleID], nil
}

func (f *fakeHighlightRepo) ListMineByArticle(_ context.Context, articleID, userID uuid.UUID) ([]engagement.Highlight, error) {
	out := []engagement.Highlight{}
	for _, h := range f.byArticle[articleID] {
		if h.UserID == userID {
			out = append(out, h)
		}
	}
	return out, nil
}

type fakeCommentRepo struct {
	byArticle map[uuid.UUID][]engagement.Comment
}

func newFakeCommentRepo() *fakeCommentRepo {
	return &fakeCommentRepo{byArticle: map[uuid.UUID][]engagement.Comment{}}
}

func (f *fakeCommentRepo) Create(_ context.Context, c engagement.Comment) (*engagement.Comment, error) {
	f.byArticle[c.ArticleID] = append(f.byArticle[c.ArticleID], c)
	return &c, nil
}

func (f *fakeCommentRepo) Delete(_ context.Context, id, userID uuid.UUID) error {
	for articleID, list := range f.byArticle {
		for i, c := range list {
			if c.ID == id && c.UserID == userID {
				f.byArticle[articleID] = append(list[:i], list[i+1:]...)
				return nil
			}
		}
	}
	return apperrors.ErrNotFound
}

func (f *fakeCommentRepo) ListByArticle(_ context.Context, articleID uuid.UUID) ([]engagement.Comment, error) {
	return f.byArticle[articleID], nil
}

type fakeIdentity struct{ users map[uuid.UUID]identity.User }

func newFakeIdentity(users ...identity.User) *fakeIdentity {
	m := map[uuid.UUID]identity.User{}
	for _, u := range users {
		m[u.ID] = u
	}
	return &fakeIdentity{users: m}
}

func (f *fakeIdentity) GetUser(_ context.Context, id uuid.UUID) (*identity.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return &u, nil
}

func TestCreateHighlight_validatesSpanAgainstBody(t *testing.T) {
	articleID := uuid.New()
	userID := uuid.New()
	article := content.Article{ID: articleID, Body: "The quick brown fox", Version: 1}
	svc := NewService(nil, newFakeHighlightRepo(), nil, newFakeContent(article), nil)

	got, err := svc.CreateHighlight(context.Background(), userID, articleID, "brown fox", 10, 19)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Matched || got.Start != 10 || got.End != 19 {
		t.Fatalf("unexpected result: %#v", got)
	}

	_, err = svc.CreateHighlight(context.Background(), userID, articleID, "brown fox", 0, 5)
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("expected ErrValidation for a span that doesn't match the quoted text, got %v", err)
	}
}

func TestDeleteHighlight_onlyOwner(t *testing.T) {
	articleID := uuid.New()
	owner := uuid.New()
	other := uuid.New()
	repo := newFakeHighlightRepo()
	svc := NewService(nil, repo, nil, newFakeContent(content.Article{ID: articleID, Body: "hello world", Version: 1}), nil)

	created, err := svc.CreateHighlight(context.Background(), owner, articleID, "hello", 0, 5)
	if err != nil {
		t.Fatal(err)
	}

	if err := svc.DeleteHighlight(context.Background(), other, created.ID); err != apperrors.ErrNotFound {
		t.Fatalf("expected ErrNotFound when a non-owner deletes, got %v", err)
	}
	if err := svc.DeleteHighlight(context.Background(), owner, created.ID); err != nil {
		t.Fatalf("expected owner delete to succeed, got %v", err)
	}
}

func TestPopularHighlight_requiresMultipleReadersAndResolvesAgainstCurrentBody(t *testing.T) {
	articleID := uuid.New()
	article := content.Article{ID: articleID, Body: "one two three four five", Version: 1}
	repo := newFakeHighlightRepo()
	svc := NewService(nil, repo, nil, newFakeContent(article), nil)

	// Only one reader highlighted "two" — shouldn't count as popular.
	if _, err := svc.CreateHighlight(context.Background(), uuid.New(), articleID, "two", 4, 7); err != nil {
		t.Fatal(err)
	}
	popular, err := svc.PopularHighlight(context.Background(), articleID)
	if err != nil {
		t.Fatal(err)
	}
	if popular != nil {
		t.Fatalf("expected no popular highlight with only one reader, got %#v", popular)
	}

	// A second reader highlights the same span — now it should surface.
	if _, err := svc.CreateHighlight(context.Background(), uuid.New(), articleID, "two", 4, 7); err != nil {
		t.Fatal(err)
	}
	popular, err = svc.PopularHighlight(context.Background(), articleID)
	if err != nil {
		t.Fatal(err)
	}
	if popular == nil || popular.Start != 4 || popular.End != 7 || popular.Count != 2 {
		t.Fatalf("expected popular highlight at (4,7,count=2), got %#v", popular)
	}
}

func TestCreateComment_generalResponseSkipsSpanValidation(t *testing.T) {
	articleID := uuid.New()
	userID := uuid.New()
	article := content.Article{ID: articleID, Body: "some body", Version: 1}
	svc := NewService(nil, nil, newFakeCommentRepo(), newFakeContent(article), newFakeIdentity())

	got, err := svc.CreateComment(context.Background(), userID, articleID, "Great article!", "", 0, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	if !got.IsGeneral || got.Matched {
		t.Fatalf("expected a general, unanchored comment, got %#v", got)
	}
}

func TestCreateComment_anchoredRequiresValidSpan(t *testing.T) {
	articleID := uuid.New()
	userID := uuid.New()
	article := content.Article{ID: articleID, Body: "some body", Version: 1}
	svc := NewService(nil, nil, newFakeCommentRepo(), newFakeContent(article), newFakeIdentity())

	_, err := svc.CreateComment(context.Background(), userID, articleID, "huh?", "wrong quote", 0, 4, false)
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("expected ErrValidation for a mismatched anchor, got %v", err)
	}
}

func TestCreateComment_emptyBody_rejected(t *testing.T) {
	articleID := uuid.New()
	article := content.Article{ID: articleID, Body: "some body", Version: 1}
	svc := NewService(nil, nil, newFakeCommentRepo(), newFakeContent(article), newFakeIdentity())

	_, err := svc.CreateComment(context.Background(), uuid.New(), articleID, "   ", "", 0, 0, true)
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("expected ErrValidation for an empty body, got %v", err)
	}
}

func TestListComments_resolvesAuthorNameFallsBackToEmailPrefix(t *testing.T) {
	articleID := uuid.New()
	article := content.Article{ID: articleID, Body: "some body", Version: 1}
	userWithName := identity.User{ID: uuid.New(), Email: "named@example.com", DisplayName: "Selam"}
	userNoName := identity.User{ID: uuid.New(), Email: "anon@example.com"}
	svc := NewService(nil, nil, newFakeCommentRepo(), newFakeContent(article), newFakeIdentity(userWithName, userNoName))

	_, err := svc.CreateComment(context.Background(), userWithName.ID, articleID, "Nice!", "", 0, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.CreateComment(context.Background(), userNoName.ID, articleID, "Same!", "", 0, 0, true)
	if err != nil {
		t.Fatal(err)
	}

	comments, err := svc.ListComments(context.Background(), articleID)
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}
	if comments[0].AuthorName != "Selam" {
		t.Fatalf("expected display name 'Selam', got %q", comments[0].AuthorName)
	}
	if comments[1].AuthorName != "anon" {
		t.Fatalf("expected email-prefix fallback 'anon', got %q", comments[1].AuthorName)
	}
}

func TestDeleteComment_onlyOwner(t *testing.T) {
	articleID := uuid.New()
	owner := uuid.New()
	other := uuid.New()
	article := content.Article{ID: articleID, Body: "some body", Version: 1}
	svc := NewService(nil, nil, newFakeCommentRepo(), newFakeContent(article), newFakeIdentity())

	created, err := svc.CreateComment(context.Background(), owner, articleID, "hi", "", 0, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteComment(context.Background(), other, created.ID); err != apperrors.ErrNotFound {
		t.Fatalf("expected ErrNotFound for a non-owner delete, got %v", err)
	}
	if err := svc.DeleteComment(context.Background(), owner, created.ID); err != nil {
		t.Fatalf("expected owner delete to succeed, got %v", err)
	}
}
