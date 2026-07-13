package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/profile"
)

type ProfileRepository interface {
	GetArtistBySlug(ctx context.Context, slug string) (*profile.ArtistProfile, error)
	GetArtistByHandle(ctx context.Context, handle string) (*profile.ArtistProfile, error)
	ListApproved(ctx context.Context, filter profile.ListFilter) ([]profile.ArtistProfile, error)
	GetArtistByID(ctx context.Context, id uuid.UUID) (*profile.ArtistProfile, error)
	GetArtistByUserID(ctx context.Context, userID uuid.UUID) (*profile.ArtistProfile, error)
	SaveArtist(ctx context.Context, p profile.ArtistProfile) (*profile.ArtistProfile, error)
}
