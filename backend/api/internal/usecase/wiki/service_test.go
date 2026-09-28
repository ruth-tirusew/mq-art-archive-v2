package wiki

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/content"
	domain "github.com/mq/api/internal/domain/wiki"
)

type submissionRepoStub struct {
	saved               *domain.Submission
	byID                map[uuid.UUID]domain.Submission
	update              func(context.Context, domain.Submission) (*domain.Submission, error)
	listPendingForOwner func(context.Context, uuid.UUID) ([]domain.Submission, error)
}

func (r *submissionRepoStub) Create(_ context.Context, item domain.Submission) (*domain.Submission, error) {
	r.saved = &item
	return &item, nil
}
func (r *submissionRepoStub) GetByID(_ context.Context, id uuid.UUID) (*domain.Submission, error) {
	if s, ok := r.byID[id]; ok {
		return &s, nil
	}
	return nil, apperrors.ErrNotFound
}
func (*submissionRepoStub) ListBySubmitter(context.Context, uuid.UUID) ([]domain.Submission, error) {
	return nil, nil
}
func (*submissionRepoStub) ListPending(context.Context) ([]domain.Submission, error) { return nil, nil }
func (r *submissionRepoStub) ListPendingForOwner(ctx context.Context, ownerID uuid.UUID) ([]domain.Submission, error) {
	if r.listPendingForOwner != nil {
		return r.listPendingForOwner(ctx, ownerID)
	}
	return nil, nil
}
func (r *submissionRepoStub) Update(ctx context.Context, item domain.Submission) (*domain.Submission, error) {
	if r.update != nil {
		return r.update(ctx, item)
	}
	return &item, nil
}

// fakeContent is a minimal stand-in for inbound.ContentService covering only what the
// wiki usecase actually calls; every method is overridable via a func field.
type fakeContent struct {
	adminGet    func(ctx context.Context, id uuid.UUID) (*content.Article, error)
	adminCreate func(ctx context.Context, authorID uuid.UUID, write content.ArticleWrite) (*content.Article, error)
	adminUpdate func(ctx context.Context, id, editorID uuid.UUID, write content.ArticleWrite) (*content.Article, error)
	approveEdit func(ctx context.Context, id, editorID uuid.UUID, expectedVersion int, submissionID uuid.UUID, write content.ArticleWrite) (*content.Article, error)
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
func (f *fakeContent) AdminGet(ctx context.Context, id uuid.UUID) (*content.Article, error) {
	if f.adminGet != nil {
		return f.adminGet(ctx, id)
	}
	return nil, apperrors.ErrNotFound
}
func (f *fakeContent) AdminCreate(ctx context.Context, authorID uuid.UUID, write content.ArticleWrite) (*content.Article, error) {
	if f.adminCreate != nil {
		return f.adminCreate(ctx, authorID, write)
	}
	return &content.Article{ID: uuid.New(), AuthorID: authorID, Title: write.Title}, nil
}
func (f *fakeContent) AdminUpdate(ctx context.Context, id, editorID uuid.UUID, write content.ArticleWrite) (*content.Article, error) {
	if f.adminUpdate != nil {
		return f.adminUpdate(ctx, id, editorID, write)
	}
	return &content.Article{ID: id, Title: write.Title}, nil
}
func (f *fakeContent) ApproveEdit(ctx context.Context, id, editorID uuid.UUID, expectedVersion int, submissionID uuid.UUID, write content.ArticleWrite) (*content.Article, error) {
	if f.approveEdit != nil {
		return f.approveEdit(ctx, id, editorID, expectedVersion, submissionID, write)
	}
	return &content.Article{ID: id, Title: write.Title, Version: expectedVersion + 1}, nil
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

func TestSubmitCreatesPendingSubmission(t *testing.T) {
	repo := &submissionRepoStub{}
	svc := NewService(repo, &fakeContent{})
	submitter := uuid.New()
	got, err := svc.Submit(context.Background(), submitter, nil, "  A title  ", "body")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.StatusPending || got.Title != "A title" || repo.saved == nil {
		t.Fatalf("unexpected submission: %#v", got)
	}
	if got.Kind != domain.KindNew || got.BasedOnVersion != nil {
		t.Fatalf("expected new-article submission with no based-on version, got %#v", got)
	}
}

func TestSubmitEdit_capturesBasedOnVersion(t *testing.T) {
	articleID := uuid.New()
	content := &fakeContent{
		adminGet: func(ctx context.Context, id uuid.UUID) (*content.Article, error) {
			return &content.Article{ID: id, Version: 4}, nil
		},
	}
	repo := &submissionRepoStub{}
	svc := NewService(repo, content)

	got, err := svc.Submit(context.Background(), uuid.New(), &articleID, "Edit", "body")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != domain.KindEdit || got.BasedOnVersion == nil || *got.BasedOnVersion != 4 {
		t.Fatalf("expected edit submission based on version 4, got %#v", got)
	}
}

func TestApprove_newArticle_usesSubmitterAsAuthor(t *testing.T) {
	submitter := uuid.New()
	reviewer := uuid.New()
	id := uuid.New()
	var capturedAuthor uuid.UUID

	repo := &submissionRepoStub{byID: map[uuid.UUID]domain.Submission{
		id: {ID: id, SubmitterID: submitter, Status: domain.StatusPending, Kind: domain.KindNew, Title: "T", Body: "B"},
	}}
	fake := &fakeContent{
		adminCreate: func(ctx context.Context, authorID uuid.UUID, write content.ArticleWrite) (*content.Article, error) {
			capturedAuthor = authorID
			return &content.Article{ID: uuid.New(), AuthorID: authorID}, nil
		},
	}
	svc := NewService(repo, fake)

	got, err := svc.Approve(context.Background(), id, reviewer, "")
	if err != nil {
		t.Fatal(err)
	}
	if capturedAuthor != submitter {
		t.Fatalf("expected article authored by submitter %s, got %s", submitter, capturedAuthor)
	}
	if got.Status != domain.StatusApproved || *got.ReviewedBy != reviewer {
		t.Fatalf("unexpected review result: %#v", got)
	}
}

func TestApprove_edit_delegatesConflictToApproveEdit(t *testing.T) {
	submitter := uuid.New()
	reviewer := uuid.New()
	articleID := uuid.New()
	id := uuid.New()
	version := 3

	repo := &submissionRepoStub{byID: map[uuid.UUID]domain.Submission{
		id: {ID: id, SubmitterID: submitter, ArticleID: &articleID, Status: domain.StatusPending, Kind: domain.KindEdit, BasedOnVersion: &version, Title: "T", Body: "B"},
	}}
	fake := &fakeContent{
		approveEdit: func(ctx context.Context, gotID, editorID uuid.UUID, expectedVersion int, submissionID uuid.UUID, write content.ArticleWrite) (*content.Article, error) {
			if gotID != articleID || editorID != submitter || expectedVersion != version || submissionID != id {
				t.Fatalf("unexpected ApproveEdit args: %v %v %v %v", gotID, editorID, expectedVersion, submissionID)
			}
			return nil, apperrors.ErrConflict
		},
	}
	svc := NewService(repo, fake)

	_, err := svc.Approve(context.Background(), id, reviewer, "")
	if err != apperrors.ErrConflict {
		t.Fatalf("expected conflict error to propagate, got %v", err)
	}
}

func TestApproveOwn_forbidsNewArticleSubmission(t *testing.T) {
	owner := uuid.New()
	id := uuid.New()
	repo := &submissionRepoStub{byID: map[uuid.UUID]domain.Submission{
		id: {ID: id, SubmitterID: uuid.New(), Status: domain.StatusPending, Kind: domain.KindNew, Title: "T", Body: "B"},
	}}
	svc := NewService(repo, &fakeContent{})

	_, err := svc.ApproveOwn(context.Background(), id, owner, "")
	if err != apperrors.ErrForbidden {
		t.Fatalf("expected forbidden for new-article submission, got %v", err)
	}
}

func TestApproveOwn_forbidsWhenNotOwner(t *testing.T) {
	owner := uuid.New()
	someoneElse := uuid.New()
	articleID := uuid.New()
	id := uuid.New()
	version := 1
	repo := &submissionRepoStub{byID: map[uuid.UUID]domain.Submission{
		id: {ID: id, SubmitterID: uuid.New(), ArticleID: &articleID, Status: domain.StatusPending, Kind: domain.KindEdit, BasedOnVersion: &version, Title: "T", Body: "B"},
	}}
	fake := &fakeContent{
		adminGet: func(ctx context.Context, gotID uuid.UUID) (*content.Article, error) {
			return &content.Article{ID: gotID, AuthorID: someoneElse, Version: version}, nil
		},
	}
	svc := NewService(repo, fake)

	_, err := svc.ApproveOwn(context.Background(), id, owner, "")
	if err != apperrors.ErrForbidden {
		t.Fatalf("expected forbidden when caller doesn't own the article, got %v", err)
	}
}

func TestApproveOwn_allowsOwner(t *testing.T) {
	owner := uuid.New()
	articleID := uuid.New()
	id := uuid.New()
	version := 1
	var approveEditCalled bool

	repo := &submissionRepoStub{byID: map[uuid.UUID]domain.Submission{
		id: {ID: id, SubmitterID: uuid.New(), ArticleID: &articleID, Status: domain.StatusPending, Kind: domain.KindEdit, BasedOnVersion: &version, Title: "T", Body: "B"},
	}}
	fake := &fakeContent{
		adminGet: func(ctx context.Context, gotID uuid.UUID) (*content.Article, error) {
			return &content.Article{ID: gotID, AuthorID: owner, Version: version}, nil
		},
		approveEdit: func(ctx context.Context, gotID, editorID uuid.UUID, expectedVersion int, submissionID uuid.UUID, write content.ArticleWrite) (*content.Article, error) {
			approveEditCalled = true
			return &content.Article{ID: gotID, Version: expectedVersion + 1}, nil
		},
	}
	svc := NewService(repo, fake)

	got, err := svc.ApproveOwn(context.Background(), id, owner, "")
	if err != nil {
		t.Fatal(err)
	}
	if !approveEditCalled {
		t.Fatal("expected ApproveEdit to be called")
	}
	if got.Status != domain.StatusApproved {
		t.Fatalf("expected approved submission, got %#v", got)
	}
}

func TestGetForReview_newArticleSubmission_hasNoArticle(t *testing.T) {
	id := uuid.New()
	repo := &submissionRepoStub{byID: map[uuid.UUID]domain.Submission{
		id: {ID: id, SubmitterID: uuid.New(), Status: domain.StatusPending, Kind: domain.KindNew, Title: "T", Body: "B"},
	}}
	svc := NewService(repo, &fakeContent{})

	submission, article, err := svc.GetForReview(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if submission.ID != id || article != nil {
		t.Fatalf("expected new-article submission with no article, got %#v / %#v", submission, article)
	}
}

func TestGetForReview_edit_returnsCurrentArticle(t *testing.T) {
	articleID := uuid.New()
	id := uuid.New()
	version := 2
	repo := &submissionRepoStub{byID: map[uuid.UUID]domain.Submission{
		id: {ID: id, SubmitterID: uuid.New(), ArticleID: &articleID, Status: domain.StatusPending, Kind: domain.KindEdit, BasedOnVersion: &version, Title: "T", Body: "B"},
	}}
	fake := &fakeContent{
		adminGet: func(ctx context.Context, gotID uuid.UUID) (*content.Article, error) {
			return &content.Article{ID: gotID, Body: "current body", Version: version}, nil
		},
	}
	svc := NewService(repo, fake)

	submission, article, err := svc.GetForReview(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if submission.ID != id || article == nil || article.Body != "current body" {
		t.Fatalf("expected edit submission with current article attached, got %#v / %#v", submission, article)
	}
}

func TestGetOwnForReview_forbidsNewArticleSubmission(t *testing.T) {
	owner := uuid.New()
	id := uuid.New()
	repo := &submissionRepoStub{byID: map[uuid.UUID]domain.Submission{
		id: {ID: id, SubmitterID: uuid.New(), Status: domain.StatusPending, Kind: domain.KindNew, Title: "T", Body: "B"},
	}}
	svc := NewService(repo, &fakeContent{})

	_, _, err := svc.GetOwnForReview(context.Background(), id, owner)
	if err != apperrors.ErrForbidden {
		t.Fatalf("expected forbidden for new-article submission, got %v", err)
	}
}

func TestGetOwnForReview_forbidsWhenNotOwner(t *testing.T) {
	owner := uuid.New()
	someoneElse := uuid.New()
	articleID := uuid.New()
	id := uuid.New()
	version := 1
	repo := &submissionRepoStub{byID: map[uuid.UUID]domain.Submission{
		id: {ID: id, SubmitterID: uuid.New(), ArticleID: &articleID, Status: domain.StatusPending, Kind: domain.KindEdit, BasedOnVersion: &version, Title: "T", Body: "B"},
	}}
	fake := &fakeContent{
		adminGet: func(ctx context.Context, gotID uuid.UUID) (*content.Article, error) {
			return &content.Article{ID: gotID, AuthorID: someoneElse, Version: version}, nil
		},
	}
	svc := NewService(repo, fake)

	_, _, err := svc.GetOwnForReview(context.Background(), id, owner)
	if err != apperrors.ErrForbidden {
		t.Fatalf("expected forbidden when caller doesn't own the article, got %v", err)
	}
}

func TestListOwnPending_delegatesToOwnerScopedQuery(t *testing.T) {
	owner := uuid.New()
	var capturedOwner uuid.UUID
	expected := []domain.Submission{{ID: uuid.New()}}

	repo := &submissionRepoStub{
		listPendingForOwner: func(ctx context.Context, ownerID uuid.UUID) ([]domain.Submission, error) {
			capturedOwner = ownerID
			return expected, nil
		},
	}
	svc := NewService(repo, &fakeContent{})

	got, err := svc.ListOwnPending(context.Background(), owner)
	if err != nil {
		t.Fatal(err)
	}
	if capturedOwner != owner || len(got) != 1 || got[0].ID != expected[0].ID {
		t.Fatalf("expected owner-scoped list to delegate correctly, got %#v (owner %s)", got, capturedOwner)
	}
}
