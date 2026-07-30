package inbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/media"
)

type MediaService interface {
	SignForUser(ctx context.Context, userID uuid.UUID) (*media.UploadSignature, error)
	CompleteUpload(ctx context.Context, userID uuid.UUID, completion media.Completion) (*media.Asset, error)
	DeleteOwned(ctx context.Context, userID, assetID uuid.UUID) error
}
