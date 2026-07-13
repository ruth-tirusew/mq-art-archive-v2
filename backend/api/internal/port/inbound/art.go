package inbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/art"
)

type ArtService interface {
	ListByArtist(ctx context.Context, artistID uuid.UUID) ([]art.ArtPost, error)
	ListPublished(ctx context.Context, filter art.ListFilter) ([]art.ArtPostWithArtist, error)
	GetByID(ctx context.Context, id uuid.UUID) (*art.ArtPost, error)
	CreateDraft(ctx context.Context, artistID uuid.UUID, title, description, medium string) (*art.ArtPost, error)
}
