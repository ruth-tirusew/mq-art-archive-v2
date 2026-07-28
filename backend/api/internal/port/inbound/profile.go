package inbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/profile"
)

type ProfileService interface {
	GetArtistBySlug(ctx context.Context, slug string) (*profile.ArtistProfile, error)
	GetArtistByHandle(ctx context.Context, handle string) (*profile.ArtistProfile, error)
	ListApproved(ctx context.Context, filter profile.ListFilter) ([]profile.ArtistProfile, error)
	ListAll(ctx context.Context, status *profile.ProfileStatus, limit, offset int) ([]profile.ArtistProfile, error)
	GetArtistByID(ctx context.Context, id uuid.UUID) (*profile.ArtistProfile, error)
	GetArtistByUserID(ctx context.Context, userID uuid.UUID) (*profile.ArtistProfile, error)
	UpdateArtist(ctx context.Context, profile profile.ArtistProfile) (*profile.ArtistProfile, error)
	UpdateOwnProfile(ctx context.Context, userID uuid.UUID, update profile.OwnProfileUpdate) (*profile.ArtistProfile, error)
	AdminCreate(ctx context.Context, write profile.AdminArtistWrite) (*profile.ArtistProfile, error)
	AdminUpdateContent(ctx context.Context, id uuid.UUID, write profile.AdminArtistWrite) (*profile.ArtistProfile, error)
	AdminDelete(ctx context.Context, id uuid.UUID) error
	AdminUpdate(ctx context.Context, id uuid.UUID, status *profile.ProfileStatus, featured *bool) (*profile.ArtistProfile, error)
}
