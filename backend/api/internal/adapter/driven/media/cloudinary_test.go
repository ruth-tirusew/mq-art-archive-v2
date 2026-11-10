package media

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strings"
	"testing"
	"time"

	domain "github.com/mq/api/internal/domain/media"
)

func newTestCloudinary(t *testing.T) *Cloudinary {
	t.Helper()
	c, err := NewCloudinary("cloud", "key", "secret", "mq", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCloudinarySignerUsesExpectedSHA1Payload(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	signer := newTestCloudinary(t)
	signer.now = func() time.Time { return now }
	got, err := signer.SignUpload(context.Background(), domain.UploadOptions{PublicID: "mq/abc", Folder: "mq"})
	if err != nil {
		t.Fatal(err)
	}
	payload := fmt.Sprintf("folder=mq&overwrite=false&public_id=mq/abc&timestamp=%dsecret", now.Unix())
	want := fmt.Sprintf("%x", sha1.Sum([]byte(payload)))
	if got.Signature != want || !got.ExpireAt.Equal(now.Add(time.Minute)) {
		t.Fatalf("unexpected signature: %#v want %s", got, want)
	}
	if got.CloudName != "cloud" || got.APIKey != "key" {
		t.Fatalf("signature should carry the account: %#v", got)
	}
}

func TestCloudinaryRefusesAFolderItDidNotIssue(t *testing.T) {
	if _, err := newTestCloudinary(t).SignUpload(context.Background(), domain.UploadOptions{PublicID: "other/abc", Folder: "other"}); err == nil {
		t.Fatal("a client must not be able to pick its own folder")
	}
}

func TestCloudinaryNeedsCredentials(t *testing.T) {
	if _, err := NewCloudinary("", "", "", "mq", time.Minute); err == nil {
		t.Fatal("missing credentials should be an error, not a silently broken uploader")
	}
}

func TestDeliveryURLIsSecureAndLossless(t *testing.T) {
	c := newTestCloudinary(t)
	got, err := c.DeliveryURL("mq/abc", 1200)
	if err != nil {
		t.Fatal(err)
	}
	want := "https://res.cloudinary.com/cloud/image/upload/c_limit,w_1200,f_webp,q_100/mq/abc"
	if got != want {
		t.Fatalf("got %s\nwant %s", got, want)
	}
	full, _ := c.DeliveryURL("mq/abc", 0)
	if !strings.Contains(full, "/upload/f_webp,q_100/mq/abc") || strings.Contains(full, "w_") {
		t.Fatalf("no width should keep the original size, got %s", full)
	}
}

func TestThumbnailURLIsATinyJPEG(t *testing.T) {
	got, err := newTestCloudinary(t).ThumbnailURL("mq/abc")
	if err != nil {
		t.Fatal(err)
	}
	for _, part := range []string{"w_64", "f_jpg", "res.cloudinary.com/cloud/"} {
		if !strings.Contains(got, part) {
			t.Fatalf("%s missing from %s", part, got)
		}
	}
}
