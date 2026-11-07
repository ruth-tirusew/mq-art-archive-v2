package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/media"
)

type MediaSigner interface {
	SignUpload(ctx context.Context, opts media.UploadOptions) (*media.UploadSignature, error)
}

type MediaStorage interface {
	Delete(ctx context.Context, publicID string) error
}

// MediaDelivery builds the URLs an image is read from.
type MediaDelivery interface {
	// ThumbnailURL is a tiny version of the image, used only to compute its blur hash.
	ThumbnailURL(publicID string) (string, error)
}

// BlurhashEncoder computes the blur hash of the image at a URL.
type BlurhashEncoder interface {
	EncodeURL(ctx context.Context, url string) (string, error)
}

type MediaAssetRepository interface {
	Create(ctx context.Context, asset media.Asset) (*media.Asset, error)
	GetByID(ctx context.Context, id uuid.UUID) (*media.Asset, error)
	GetByPublicID(ctx context.Context, publicID string) (*media.Asset, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
