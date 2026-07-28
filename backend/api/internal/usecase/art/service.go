package art

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
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

func (s *Service) ListOwnedByArtist(ctx context.Context, artistID uuid.UUID) ([]domain.ArtPost, error) {
	return s.posts.ListOwnedByArtist(ctx, artistID)
}

func (s *Service) ListPublished(ctx context.Context, filter domain.ListFilter) ([]domain.ArtPostWithArtist, error) {
	return s.posts.ListPublished(ctx, filter)
}

func (s *Service) ListAll(ctx context.Context, status *domain.ArtStatus, limit, offset int) ([]domain.ArtPostWithArtist, error) {
	return s.posts.ListAll(ctx, status, limit, offset)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.ArtPost, error) {
	return s.posts.GetByID(ctx, id)
}

func (s *Service) GetOwned(ctx context.Context, artistID, postID uuid.UUID) (*domain.ArtPost, error) {
	post, err := s.posts.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if post.ArtistID != artistID {
		return nil, apperrors.ErrForbidden
	}
	return post, nil
}

func (s *Service) CreateDraft(ctx context.Context, artistID uuid.UUID, write domain.ArtPostWrite) (*domain.ArtPost, error) {
	if err := requireTitle(write.Title); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	post := domain.ArtPost{
		ID:          uuid.New(),
		ArtistID:    artistID,
		Title:       write.Title,
		Description: write.Description,
		Medium:      write.Medium,
		Year:        write.Year,
		Dimensions:  write.Dimensions,
		City:        write.City,
		Style:       write.Style,
		Residency:   write.Residency,
		Exhibition:  write.Exhibition,
		Palette:     write.Palette,
		Media:       mediaFromURLs(write.MediaURLs),
		Status:      domain.ArtStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return s.posts.Create(ctx, post)
}

func (s *Service) UpdateOwned(ctx context.Context, artistID, postID uuid.UUID, write domain.ArtPostWrite) (*domain.ArtPost, error) {
	post, err := s.GetOwned(ctx, artistID, postID)
	if err != nil {
		return nil, err
	}
	if err := requireTitle(write.Title); err != nil {
		return nil, err
	}

	post.Title = write.Title
	post.Description = write.Description
	post.Medium = write.Medium
	post.Year = write.Year
	post.Dimensions = write.Dimensions
	post.City = write.City
	post.Style = write.Style
	post.Residency = write.Residency
	post.Exhibition = write.Exhibition
	post.Palette = write.Palette
	post.Media = mediaFromURLs(write.MediaURLs)
	post.UpdatedAt = time.Now().UTC()

	return s.posts.Update(ctx, *post)
}

func (s *Service) PublishOwned(ctx context.Context, artistID, postID uuid.UUID) (*domain.ArtPost, error) {
	post, err := s.GetOwned(ctx, artistID, postID)
	if err != nil {
		return nil, err
	}
	if post.Status != domain.ArtStatusDraft && post.Status != domain.ArtStatusArchived {
		return nil, apperrors.ErrConflict
	}
	now := time.Now().UTC()
	post.Status = domain.ArtStatusPublished
	post.PublishedAt = &now
	post.UpdatedAt = now
	return s.posts.Update(ctx, *post)
}

func (s *Service) UnpublishOwned(ctx context.Context, artistID, postID uuid.UUID) (*domain.ArtPost, error) {
	post, err := s.GetOwned(ctx, artistID, postID)
	if err != nil {
		return nil, err
	}
	if post.Status != domain.ArtStatusPublished {
		return nil, apperrors.ErrConflict
	}
	post.Status = domain.ArtStatusDraft
	post.PublishedAt = nil
	post.UpdatedAt = time.Now().UTC()
	return s.posts.Update(ctx, *post)
}

func (s *Service) ArchiveOwned(ctx context.Context, artistID, postID uuid.UUID) (*domain.ArtPost, error) {
	post, err := s.GetOwned(ctx, artistID, postID)
	if err != nil {
		return nil, err
	}
	post.Status = domain.ArtStatusArchived
	post.UpdatedAt = time.Now().UTC()
	return s.posts.Update(ctx, *post)
}

func (s *Service) DeleteOwned(ctx context.Context, artistID, postID uuid.UUID) error {
	post, err := s.GetOwned(ctx, artistID, postID)
	if err != nil {
		return err
	}
	if post.Status == domain.ArtStatusPublished {
		return apperrors.ErrConflict
	}
	return s.posts.Delete(ctx, postID)
}

func (s *Service) AdminCreate(ctx context.Context, artistID uuid.UUID, write domain.ArtPostWrite, status *domain.ArtStatus) (*domain.ArtPost, error) {
	if err := requireTitle(write.Title); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	st := domain.ArtStatusDraft
	if status != nil && *status != "" {
		st = *status
	}
	post := domain.ArtPost{
		ID:          uuid.New(),
		ArtistID:    artistID,
		Title:       write.Title,
		Description: write.Description,
		Medium:      write.Medium,
		Year:        write.Year,
		Dimensions:  write.Dimensions,
		City:        write.City,
		Style:       write.Style,
		Residency:   write.Residency,
		Exhibition:  write.Exhibition,
		Palette:     write.Palette,
		Media:       mediaFromURLs(write.MediaURLs),
		Status:      st,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if st == domain.ArtStatusPublished {
		post.PublishedAt = &now
	}

	return s.posts.Create(ctx, post)
}

func (s *Service) AdminUpdateContent(ctx context.Context, postID uuid.UUID, write domain.ArtPostWrite) (*domain.ArtPost, error) {
	post, err := s.posts.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if err := requireTitle(write.Title); err != nil {
		return nil, err
	}

	post.Title = write.Title
	post.Description = write.Description
	post.Medium = write.Medium
	post.Year = write.Year
	post.Dimensions = write.Dimensions
	post.City = write.City
	post.Style = write.Style
	post.Residency = write.Residency
	post.Exhibition = write.Exhibition
	post.Palette = write.Palette
	post.Media = mediaFromURLs(write.MediaURLs)
	post.UpdatedAt = time.Now().UTC()

	return s.posts.Update(ctx, *post)
}

func (s *Service) AdminDelete(ctx context.Context, postID uuid.UUID) error {
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		return err
	}
	return s.posts.Delete(ctx, postID)
}

func (s *Service) AdminUpdate(ctx context.Context, postID uuid.UUID, status *domain.ArtStatus, featured *bool) (*domain.ArtPost, error) {
	post, err := s.posts.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if status != nil {
		post.Status = *status
		if *status == domain.ArtStatusPublished && post.PublishedAt == nil {
			now := time.Now().UTC()
			post.PublishedAt = &now
		}
		if *status == domain.ArtStatusDraft {
			post.PublishedAt = nil
		}
	}
	if featured != nil {
		post.FeaturedAcquisition = *featured
	}
	post.UpdatedAt = time.Now().UTC()
	return s.posts.Update(ctx, *post)
}

func mediaFromURLs(urls []string) []domain.MediaAsset {
	out := make([]domain.MediaAsset, 0, len(urls))
	for i, u := range urls {
		if u == "" {
			continue
		}
		out = append(out, domain.MediaAsset{
			ID:        uuid.New(),
			URL:       u,
			SortOrder: i,
		})
	}
	return out
}
