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

type MediaAssetRepository interface {
	Create(ctx context.Context, asset media.Asset) (*media.Asset, error)
	GetByID(ctx context.Context, id uuid.UUID) (*media.Asset, error)
	GetByPublicID(ctx context.Context, publicID string) (*media.Asset, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
