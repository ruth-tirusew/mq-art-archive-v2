package inbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/art"
)

type ArtService interface {
	ListByArtist(ctx context.Context, artistID uuid.UUID) ([]art.ArtPost, error)
	ListOwnedByArtist(ctx context.Context, artistID uuid.UUID) ([]art.ArtPost, error)
	ListPublished(ctx context.Context, filter art.ListFilter) ([]art.ArtPostWithArtist, error)
	ListAll(ctx context.Context, status *art.ArtStatus, limit, offset int) ([]art.ArtPostWithArtist, error)
	GetByID(ctx context.Context, id uuid.UUID) (*art.ArtPost, error)
	GetOwned(ctx context.Context, artistID, postID uuid.UUID) (*art.ArtPost, error)
	CreateDraft(ctx context.Context, artistID uuid.UUID, write art.ArtPostWrite) (*art.ArtPost, error)
	UpdateOwned(ctx context.Context, artistID, postID uuid.UUID, write art.ArtPostWrite) (*art.ArtPost, error)
	PublishOwned(ctx context.Context, artistID, postID uuid.UUID) (*art.ArtPost, error)
	UnpublishOwned(ctx context.Context, artistID, postID uuid.UUID) (*art.ArtPost, error)
	ArchiveOwned(ctx context.Context, artistID, postID uuid.UUID) (*art.ArtPost, error)
	DeleteOwned(ctx context.Context, artistID, postID uuid.UUID) error
	// Admin
	AdminCreate(ctx context.Context, artistID uuid.UUID, write art.ArtPostWrite, status *art.ArtStatus) (*art.ArtPost, error)
	AdminUpdateContent(ctx context.Context, postID uuid.UUID, write art.ArtPostWrite) (*art.ArtPost, error)
	AdminDelete(ctx context.Context, postID uuid.UUID) error
	AdminUpdate(ctx context.Context, postID uuid.UUID, status *art.ArtStatus, featured *bool) (*art.ArtPost, error)
}
