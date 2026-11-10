package media

import (
	"context"
	"errors"
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
		PublicID: "mq/abc", SecureURL: "https://res.cloudinary.com/cloud/image/upload/v1/mq/abc.webp",
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
		// Host must be exactly res.cloudinary.com, not merely end in cloudinary.com...
		{PublicID: "mq/a", SecureURL: "https://evilcloudinary.com/mq/a.jpg", ResourceType: "image", Format: "jpeg", Width: 1, Height: 1, Bytes: 1},
		{PublicID: "mq/a", SecureURL: "https://res.cloudinary.com.evil.example/mq/a.jpg", ResourceType: "image", Format: "jpeg", Width: 1, Height: 1, Bytes: 1},
		// ...and the URL has to point at the asset that was actually signed.
		{PublicID: "mq/a", SecureURL: "https://res.cloudinary.com/cloud/image/upload/v1/someone/else.jpg", ResourceType: "image", Format: "jpeg", Width: 1, Height: 1, Bytes: 1},
		{PublicID: "mq/a", SecureURL: "https://res.cloudinary.com/x", ResourceType: "image", Format: "jpeg", Bytes: domain.MaxImageBytes + 1},
	}
	for i, completion := range cases {
		if _, err := svc.CompleteUpload(context.Background(), uuid.New(), completion); err == nil {
			t.Errorf("case %d: expected error", i)
		}
	}
}

type deliveryStub struct{}

func (deliveryStub) ThumbnailURL(publicID string) (string, error) {
	return "https://res.cloudinary.com/cloud/image/upload/w_64/" + publicID, nil
}

type hasherStub struct {
	hash string
	err  error
	got  string
}

func (h *hasherStub) EncodeURL(_ context.Context, url string) (string, error) {
	h.got = url
	return h.hash, h.err
}

var completion = domain.Completion{
	PublicID: "mq/abc", SecureURL: "https://res.cloudinary.com/cloud/image/upload/v1/mq/abc.webp",
	ResourceType: "image", Format: "webp", Width: 100, Height: 200, Bytes: 1024,
}

func TestCompleteUploadStoresTheBlurhashOfTheThumbnail(t *testing.T) {
	hasher := &hasherStub{hash: "LEHV6nWB2yk8pyo0adR*.7kCMdnj"}
	repo := &repoStub{}
	svc := NewService(signerStub{}, storageStub{}, repo, "mq", WithBlurhash(deliveryStub{}, hasher))
	asset, err := svc.CompleteUpload(context.Background(), uuid.New(), completion)
	if err != nil {
		t.Fatal(err)
	}
	if asset.Blurhash != "LEHV6nWB2yk8pyo0adR*.7kCMdnj" || repo.saved.Blurhash != asset.Blurhash {
		t.Fatalf("blur hash not stored: %q", asset.Blurhash)
	}
	if hasher.got != "https://res.cloudinary.com/cloud/image/upload/w_64/mq/abc" {
		t.Fatalf("the hash should come from the tiny rendition, not the original: %s", hasher.got)
	}
}

func TestCompleteUploadSucceedsWithoutABlurhashWhenItCannotBeComputed(t *testing.T) {
	repo := &repoStub{}
	svc := NewService(signerStub{}, storageStub{}, repo, "mq", WithBlurhash(deliveryStub{}, &hasherStub{err: errors.New("cloudinary is slow")}))
	asset, err := svc.CompleteUpload(context.Background(), uuid.New(), completion)
	if err != nil {
		t.Fatalf("an upload must not fail over a placeholder: %v", err)
	}
	if asset.Blurhash != "" || repo.saved == nil {
		t.Fatalf("expected an asset saved with an empty hash for the backfill, got %q", asset.Blurhash)
	}
}
