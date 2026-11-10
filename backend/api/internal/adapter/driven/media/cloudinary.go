package media

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	domain "github.com/mq/api/internal/domain/media"
)

// Cloudinary is the image store: signed direct uploads from the browser, deletion, and delivery
// URLs. It is built on Cloudinary's official Go SDK, so signing and URL building follow the
// vendor's own rules rather than a hand-copy of them.
type Cloudinary struct {
	cld       *cloudinary.Cloudinary
	apiSecret string
	folder    string
	expiry    time.Duration
	now       func() time.Time
}

func NewCloudinary(cloudName, apiKey, apiSecret, folder string, expiry time.Duration) (*Cloudinary, error) {
	if cloudName == "" || apiKey == "" || apiSecret == "" || folder == "" {
		return nil, fmt.Errorf("cloudinary: cloud name, API key, API secret and folder are all required")
	}
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("cloudinary: %w", err)
	}
	cld.Config.URL.Secure = true
	// Plain, cacheable URLs: no injected /v1/ segment and no SDK usage-analytics query string.
	cld.Config.URL.ForceVersion = false
	cld.Config.URL.Analytics = false
	return &Cloudinary{cld: cld, apiSecret: apiSecret, folder: folder, expiry: expiry, now: time.Now}, nil
}

// SignUpload signs the parameters the browser will send straight to Cloudinary. The public id and
// folder are fixed here, so a client cannot choose where its file lands or overwrite another's.
func (c *Cloudinary) SignUpload(_ context.Context, opts domain.UploadOptions) (*domain.UploadSignature, error) {
	if opts.Folder != c.folder {
		return nil, fmt.Errorf("upload folder is not allowed")
	}
	timestamp := c.now().UTC()
	signature, err := api.SignParameters(url.Values{
		"folder":    {opts.Folder},
		"overwrite": {"false"},
		"public_id": {opts.PublicID},
		"timestamp": {strconv.FormatInt(timestamp.Unix(), 10)},
	}, c.apiSecret)
	if err != nil {
		return nil, fmt.Errorf("sign upload: %w", err)
	}
	return &domain.UploadSignature{
		Timestamp: timestamp.Unix(), Signature: signature, CloudName: c.cld.Config.Cloud.CloudName,
		APIKey: c.cld.Config.Cloud.APIKey, Folder: c.folder, PublicID: opts.PublicID, ExpireAt: timestamp.Add(c.expiry),
	}, nil
}

func (c *Cloudinary) Delete(ctx context.Context, publicID string) error {
	res, err := c.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID, ResourceType: "image", Invalidate: api.Bool(true),
	})
	if err != nil {
		return err
	}
	if res.Error.Message != "" {
		return fmt.Errorf("cloudinary delete: %s", res.Error.Message)
	}
	return nil
}

// Lossless is the transformation for full-quality delivery: WebP at quality 100 is Cloudinary's
// lossless WebP, which is smaller than the original JPEG/PNG with no loss of image data. The
// original file is never modified; this only changes what is delivered.
const Lossless = "f_webp,q_100"

// DeliveryURL is the URL to show an image at (at most) width pixels wide, delivered losslessly.
// A width of 0 keeps the original size.
func (c *Cloudinary) DeliveryURL(publicID string, width int) (string, error) {
	return c.transformed(publicID, losslessTransformation(width))
}

// ThumbnailURL is a tiny JPEG used only to compute a blur hash. It is lossy on purpose: a blur
// hash needs a handful of colours, not detail, and standard-library decoding keeps the server light.
func (c *Cloudinary) ThumbnailURL(publicID string) (string, error) {
	return c.transformed(publicID, "c_limit,w_64,h_64,f_jpg,q_70")
}

func losslessTransformation(width int) string {
	if width <= 0 {
		return Lossless
	}
	return fmt.Sprintf("c_limit,w_%d,%s", width, Lossless)
}

func (c *Cloudinary) transformed(publicID, transformation string) (string, error) {
	img, err := c.cld.Image(publicID)
	if err != nil {
		return "", err
	}
	img.Transformation = transformationOf(transformation)
	return img.String()
}
