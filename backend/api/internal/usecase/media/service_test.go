package media

import (
	"context"
	"testing"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/media"
)

type signerStub struct{}

func (signerStub) SignUpload(_ context.Context, o domain.UploadOptions) (*domain.UploadSignature, error) {
	return &domain.UploadSignature{PublicID: o.PublicID, Folder: o.Folder}, nil
}

type repoStub struct{ saved *domain.Asset }

func (r *repoStub) Create(_ context.Context, a domain.Asset) (*domain.Asset, error) {
	r.saved = &a
	return &a, nil
}
func (r *repoStub) GetByID(context.Context, uuid.UUID) (*domain.Asset, error)    { return r.saved, nil }
func (r *repoStub) GetByPublicID(context.Context, string) (*domain.Asset, error) { return r.saved, nil }
func (r *repoStub) Delete(context.Context, uuid.UUID) error                      { return nil }

type storageStub struct{}

func (storageStub) Delete(context.Context, string) error { return nil }

func TestCompleteUploadValidatesAndPersists(t *testing.T) {
	repo := &repoStub{}
	svc := NewService(signerStub{}, storageStub{}, repo, "mq")
	owner := uuid.New()
	asset, err := svc.CompleteUpload(context.Background(), owner, domain.Completion{
		PublicID: "mq/abc", SecureURL: "https://res.cloudinary.com/cloud/image/upload/abc.webp",
		ResourceType: "image", Format: "webp", Width: 100, Height: 200, Bytes: 1024,
	})
	if err != nil {
		t.Fatal(err)
	}
	if asset.OwnerUserID != owner || repo.saved == nil {
		t.Fatal("asset was not persisted for owner")
	}
}

func TestCompleteUploadRejectsUnsafeMetadata(t *testing.T) {
	svc := NewService(signerStub{}, storageStub{}, &repoStub{}, "mq")
	cases := []domain.Completion{
		{PublicID: "other/a", SecureURL: "https://res.cloudinary.com/x", ResourceType: "image", Format: "jpeg", Bytes: 1},
		{PublicID: "mq/a", SecureURL: "http://res.cloudinary.com/x", ResourceType: "image", Format: "jpeg", Bytes: 1},
		{PublicID: "mq/a", SecureURL: "https://res.cloudinary.com/x", ResourceType: "image", Format: "gif", Bytes: 1},
		{PublicID: "mq/a", SecureURL: "https://res.cloudinary.com/x", ResourceType: "image", Format: "jpeg", Bytes: domain.MaxImageBytes + 1},
	}
	for i, completion := range cases {
		if _, err := svc.CompleteUpload(context.Background(), uuid.New(), completion); err == nil {
			t.Errorf("case %d: expected error", i)
		}
	}
}
