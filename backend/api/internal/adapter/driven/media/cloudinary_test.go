package media

import (
	"context"
	"crypto/sha1"
	"fmt"
	"testing"
	"time"

	domain "github.com/mq/api/internal/domain/media"
)

func TestCloudinarySignerUsesExpectedSHA1Payload(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	signer := NewCloudinary("cloud", "key", "secret", "mq", time.Minute)
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
}
