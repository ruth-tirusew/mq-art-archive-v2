package media

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	domain "github.com/mq/api/internal/domain/media"
)

type Cloudinary struct {
	cloudName string
	apiKey    string
	apiSecret string
	folder    string
	expiry    time.Duration
	client    *http.Client
	now       func() time.Time
}

func NewCloudinary(cloudName, apiKey, apiSecret, folder string, expiry time.Duration) *Cloudinary {
	return &Cloudinary{
		cloudName: cloudName, apiKey: apiKey, apiSecret: apiSecret, folder: folder,
		expiry: expiry, client: &http.Client{Timeout: 10 * time.Second}, now: time.Now,
	}
}

func (c *Cloudinary) SignUpload(_ context.Context, opts domain.UploadOptions) (*domain.UploadSignature, error) {
	if opts.Folder != c.folder {
		return nil, fmt.Errorf("upload folder is not allowed")
	}
	timestamp := c.now().UTC()
	payload := fmt.Sprintf("folder=%s&overwrite=false&public_id=%s&timestamp=%d%s", opts.Folder, opts.PublicID, timestamp.Unix(), c.apiSecret)
	sum := sha1.Sum([]byte(payload))
	return &domain.UploadSignature{
		Timestamp: timestamp.Unix(), Signature: hex.EncodeToString(sum[:]), CloudName: c.cloudName,
		APIKey: c.apiKey, Folder: c.folder, PublicID: opts.PublicID, ExpireAt: timestamp.Add(c.expiry),
	}, nil
}

func (c *Cloudinary) Delete(ctx context.Context, publicID string) error {
	timestamp := c.now().UTC().Unix()
	payload := fmt.Sprintf("public_id=%s&timestamp=%d%s", publicID, timestamp, c.apiSecret)
	sum := sha1.Sum([]byte(payload))
	form := url.Values{
		"public_id": {publicID}, "timestamp": {fmt.Sprint(timestamp)}, "api_key": {c.apiKey},
		"signature": {hex.EncodeToString(sum[:])}, "resource_type": {"image"},
	}
	endpoint := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/destroy", c.cloudName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("cloudinary delete returned %s", resp.Status)
	}
	return nil
}
