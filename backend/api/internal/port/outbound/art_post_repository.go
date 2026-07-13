package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/art"
)

type ArtPostRepository interface {
	ListByArtist(ctx context.Context, artistID uuid.UUID) ([]art.ArtPost, error)
	ListPublished(ctx context.Context, filter art.ListFilter) ([]art.ArtPostWithArtist, error)
	GetByID(ctx context.Context, id uuid.UUID) (*art.ArtPost, error)
	Create(ctx context.Context, post art.ArtPost) (*art.ArtPost, error)
}
