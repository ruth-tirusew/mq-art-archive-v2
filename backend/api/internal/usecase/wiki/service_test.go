package wiki

import (
	"context"
	"testing"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/wiki"
)

type submissionRepoStub struct{ saved *domain.Submission }

func (r *submissionRepoStub) Create(_ context.Context, item domain.Submission) (*domain.Submission, error) {
	r.saved = &item
	return &item, nil
}
func (*submissionRepoStub) GetByID(context.Context, uuid.UUID) (*domain.Submission, error) {
	return nil, nil
}
func (*submissionRepoStub) ListBySubmitter(context.Context, uuid.UUID) ([]domain.Submission, error) {
	return nil, nil
}
func (*submissionRepoStub) ListPending(context.Context) ([]domain.Submission, error) { return nil, nil }
func (*submissionRepoStub) Update(_ context.Context, item domain.Submission) (*domain.Submission, error) {
	return &item, nil
}

func TestSubmitCreatesPendingSubmission(t *testing.T) {
	repo := &submissionRepoStub{}
	svc := NewService(repo, nil)
	submitter := uuid.New()
	got, err := svc.Submit(context.Background(), submitter, nil, "  A title  ", "body")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.StatusPending || got.Title != "A title" || repo.saved == nil {
		t.Fatalf("unexpected submission: %#v", got)
	}
}
