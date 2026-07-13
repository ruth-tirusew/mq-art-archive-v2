package art

import (
	"context"
	"time"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

type Service struct {
	posts outbound.ArtPostRepository
}

func NewService(posts outbound.ArtPostRepository) inbound.ArtService {
	return &Service{posts: posts}
}

func (s *Service) ListByArtist(ctx context.Context, artistID uuid.UUID) ([]domain.ArtPost, error) {
	return s.posts.ListByArtist(ctx, artistID)
}

func (s *Service) ListPublished(ctx context.Context, filter domain.ListFilter) ([]domain.ArtPostWithArtist, error) {
	return s.posts.ListPublished(ctx, filter)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.ArtPost, error) {
	return s.posts.GetByID(ctx, id)
}

func (s *Service) CreateDraft(ctx context.Context, artistID uuid.UUID, title, description, medium string) (*domain.ArtPost, error) {
	if err := requireTitle(title); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	post := domain.ArtPost{
		ID:          uuid.New(),
		ArtistID:    artistID,
		Title:       title,
		Description: description,
		Medium:      medium,
		Status:      domain.ArtStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return s.posts.Create(ctx, post)
}
