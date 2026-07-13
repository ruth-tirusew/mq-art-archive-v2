package profile

import (
	"context"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/profile"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

type Service struct {
	profiles outbound.ProfileRepository
}

func NewService(profiles outbound.ProfileRepository) inbound.ProfileService {
	return &Service{profiles: profiles}
}

func (s *Service) GetArtistBySlug(ctx context.Context, slug string) (*domain.ArtistProfile, error) {
	return s.profiles.GetArtistBySlug(ctx, slug)
}

func (s *Service) GetArtistByHandle(ctx context.Context, handle string) (*domain.ArtistProfile, error) {
	return s.profiles.GetArtistByHandle(ctx, handle)
}

func (s *Service) ListApproved(ctx context.Context, filter domain.ListFilter) ([]domain.ArtistProfile, error) {
	return s.profiles.ListApproved(ctx, filter)
}

func (s *Service) GetArtistByID(ctx context.Context, id uuid.UUID) (*domain.ArtistProfile, error) {
	return s.profiles.GetArtistByID(ctx, id)
}

func (s *Service) GetArtistByUserID(ctx context.Context, userID uuid.UUID) (*domain.ArtistProfile, error) {
	return s.profiles.GetArtistByUserID(ctx, userID)
}

func (s *Service) UpdateArtist(ctx context.Context, p domain.ArtistProfile) (*domain.ArtistProfile, error) {
	return s.profiles.SaveArtist(ctx, p)
}
