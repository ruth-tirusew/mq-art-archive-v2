package media

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bbrks/go-blurhash"
)

func solid(c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 32, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 32; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func TestEncodeProducesAHashThatDecodesBackToTheSameColour(t *testing.T) {
	hash, err := Encode(solid(color.RGBA{R: 200, G: 40, B: 30, A: 255}))
	if err != nil {
		t.Fatal(err)
	}
	if len(hash) < 6 || len(hash) > 40 {
		t.Fatalf("a 4x3 blur hash is short, got %d chars: %s", len(hash), hash)
	}
	decoded, err := blurhash.Decode(hash, 8, 8, 1)
	if err != nil {
		t.Fatal(err)
	}
	r, g, b, _ := decoded.At(4, 4).RGBA()
	if r>>8 < 170 || g>>8 > 80 || b>>8 > 70 {
		t.Fatalf("decoded colour drifted: %d %d %d", r>>8, g>>8, b>>8)
	}
}

func TestBlurhasherRefusesAnythingButCloudinaryOverHTTPS(t *testing.T) {
	b := NewBlurhasher()
	for _, bad := range []string{
		"http://res.cloudinary.com/cloud/image/upload/a.jpg",
		"https://evil.example/a.jpg",
		"https://res.cloudinary.com.evil.example/a.jpg",
		"file:///etc/passwd",
		"http://169.254.169.254/latest/meta-data",
	} {
		if _, err := b.EncodeURL(context.Background(), bad); err == nil {
			t.Errorf("%s should be refused", bad)
		}
	}
}

func TestBlurhasherFetchesDecodesAndEncodes(t *testing.T) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, solid(color.RGBA{R: 20, G: 60, B: 200, A: 255}), nil); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write(buf.Bytes())
	}))
	defer srv.Close()

	b := NewBlurhasher()
	b.limit = maxThumbnailBytes
	hash, err := b.encodeFrom(context.Background(), srv.URL)
	if err != nil || hash == "" {
		t.Fatalf("hash=%q err=%v", hash, err)
	}

	notFound := httptest.NewServer(http.NotFoundHandler())
	defer notFound.Close()
	if _, err := b.encodeFrom(context.Background(), notFound.URL); err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("a missing image should be an error, got %v", err)
	}

	garbage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("not an image")) }))
	defer garbage.Close()
	if _, err := b.encodeFrom(context.Background(), garbage.URL); err == nil {
		t.Fatal("undecodable bytes should be an error")
	}
}

func TestRemoteBlurhasherAllowsAnyHTTPSHostButStillOnlyHTTPS(t *testing.T) {
	b := NewRemoteBlurhasher()
	if _, err := b.EncodeURL(context.Background(), "http://images.example/a.jpg"); err == nil {
		t.Error("plain http must still be refused")
	}
	if _, err := b.EncodeURL(context.Background(), "file:///etc/passwd"); err == nil {
		t.Error("non-http schemes must be refused")
	}
	// A different host passes the URL check (the request then fails to connect, which is fine).
	_, err := b.EncodeURL(context.Background(), "https://invalid.invalid/a.jpg")
	if err == nil || strings.Contains(err.Error(), "refusing") {
		t.Errorf("an https URL on another host should be attempted, got %v", err)
	}
}

func TestThumbnailFromURL(t *testing.T) {
	got, ok := ThumbnailFromURL("https://res.cloudinary.com/cloud/image/upload/v1699/mq/abc.webp?x=1")
	want := "https://res.cloudinary.com/cloud/image/upload/c_limit,w_64,h_64,f_jpg,q_70/v1699/mq/abc.webp"
	if !ok || got != want {
		t.Fatalf("got %q %v\nwant %q", got, ok, want)
	}
	for _, bad := range []string{
		"https://images.unsplash.com/photo-1?w=1200",
		"http://res.cloudinary.com/cloud/image/upload/mq/abc",
		"https://res.cloudinary.com/cloud/raw/upload/mq/abc",
		"not a url",
	} {
		if _, ok := ThumbnailFromURL(bad); ok {
			t.Errorf("%s is not a Cloudinary image upload URL", bad)
		}
	}
}
