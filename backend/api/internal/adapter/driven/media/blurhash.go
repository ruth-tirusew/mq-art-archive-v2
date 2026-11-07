package media

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg" // registered for image.Decode: the thumbnail is requested as a JPEG
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bbrks/go-blurhash"
)

// maxThumbnailBytes bounds what is read for a blur hash; the thumbnail is a few KB.
const maxThumbnailBytes = 512 * 1024

// maxRemoteImageBytes bounds a full-size image fetched by the backfill command.
const maxRemoteImageBytes = 8 * 1024 * 1024

// Blurhasher turns an image URL into a blur hash. It only fetches from Cloudinary's delivery host
// over https, so a crafted URL can't make the server request anything else.
type Blurhasher struct {
	client *http.Client
	host   string // "" allows any https host (backfill only)
	limit  int64
}

func NewBlurhasher() *Blurhasher {
	return &Blurhasher{client: &http.Client{Timeout: 10 * time.Second}, host: "res.cloudinary.com", limit: maxThumbnailBytes}
}

// NewRemoteBlurhasher fetches full images from any https host. It is for the backfill command,
// which an operator runs by hand over images already in the database (seed data, other CDNs). The
// API itself never uses it, because a request-driven fetch of arbitrary URLs would be an SSRF hole.
func NewRemoteBlurhasher() *Blurhasher {
	return &Blurhasher{client: &http.Client{Timeout: 20 * time.Second}, limit: maxRemoteImageBytes}
}

func (b *Blurhasher) EncodeURL(ctx context.Context, rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" || (b.host != "" && !strings.EqualFold(u.Hostname(), b.host)) {
		return "", fmt.Errorf("blurhash: refusing to fetch %q", rawURL)
	}
	return b.encodeFrom(ctx, u.String())
}

func (b *Blurhasher) encodeFrom(ctx context.Context, target string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return "", err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("blurhash: fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("blurhash: fetch returned %s", resp.Status)
	}
	img, _, err := image.Decode(io.LimitReader(resp.Body, b.limit))
	if err != nil {
		return "", fmt.Errorf("blurhash: decode: %w", err)
	}
	return Encode(img)
}

// Encode computes the blur hash of an already-decoded image. 4x3 components suits landscape art
// and is still a short string; portraits and squares decode fine at that ratio too.
func Encode(img image.Image) (string, error) {
	hash, err := blurhash.Encode(4, 3, img)
	if err != nil {
		return "", fmt.Errorf("blurhash: encode: %w", err)
	}
	return hash, nil
}

// ThumbnailFromURL derives the tiny-JPEG rendition URL from an uploaded image's delivery URL by
// inserting the transformation after "/upload/". It needs no credentials, which is what lets the
// backfill command run against assets it only knows by URL. ok is false for anything that isn't a
// Cloudinary upload URL.
func ThumbnailFromURL(secureURL string) (string, bool) {
	u, err := url.Parse(secureURL)
	if err != nil || u.Scheme != "https" || !strings.EqualFold(u.Hostname(), "res.cloudinary.com") {
		return "", false
	}
	const marker = "/image/upload/"
	i := strings.Index(u.Path, marker)
	if i < 0 {
		return "", false
	}
	u.Path = u.Path[:i+len(marker)] + "c_limit,w_64,h_64,f_jpg,q_70/" + u.Path[i+len(marker):]
	u.RawQuery = ""
	return u.String(), true
}
