package profile

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/identity"
	domain "github.com/mq/api/internal/domain/profile"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

type Service struct {
	profiles outbound.ProfileRepository
	users    outbound.UserRepository
	posts    outbound.ArtPostRepository
}

func NewService(
	profiles outbound.ProfileRepository,
	users outbound.UserRepository,
	posts outbound.ArtPostRepository,
) inbound.ProfileService {
	return &Service{profiles: profiles, users: users, posts: posts}
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

func (s *Service) CountApproved(ctx context.Context, filter domain.ListFilter) (int, error) {
	counter, ok := s.profiles.(interface {
		CountApproved(context.Context, domain.ListFilter) (int, error)
	})
	if !ok {
		items, err := s.profiles.ListApproved(ctx, filter)
		return len(items), err
	}
	return counter.CountApproved(ctx, filter)
}

func (s *Service) ListAll(ctx context.Context, status *domain.ProfileStatus, limit, offset int) ([]domain.ArtistProfile, error) {
	return s.profiles.ListAll(ctx, status, limit, offset)
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

func (s *Service) UpdateOwnProfile(ctx context.Context, userID uuid.UUID, update domain.OwnProfileUpdate) (*domain.ArtistProfile, error) {
	current, err := s.profiles.GetArtistByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	updated := *current
	if strings.TrimSpace(update.DisplayName) != "" {
		updated.DisplayName = strings.TrimSpace(update.DisplayName)
	}
	if strings.TrimSpace(update.Slug) != "" {
		updated.Slug = slugify(update.Slug)
	}
	if strings.TrimSpace(update.Handle) != "" {
		updated.Handle = slugify(update.Handle)
	}
	updated.Bio = update.Bio
	updated.Born = update.Born
	updated.Discipline = update.Discipline
	updated.Tagline = update.Tagline
	updated.YearsActive = update.YearsActive
	updated.PortraitURL = update.PortraitURL
	if update.Influences != nil {
		updated.Influences = update.Influences
	}
	updated.InResidence = update.InResidence
	updated.ResidencyPlace = update.ResidencyPlace
	updated.OpenForCommission = update.OpenForCommission
	updated.Contact = update.Contact
	updated.Social = update.Social

	if update.Status != "" {
		if update.Status != domain.ProfileStatusDraft && update.Status != domain.ProfileStatusPending {
			return nil, apperrors.ErrForbidden
		}
		updated.Status = update.Status
	}

	if updated.Slug == "" {
		updated.Slug = slugify(updated.DisplayName)
	}
	if updated.Handle == "" {
		updated.Handle = updated.Slug
	}

	if err := s.ensureUniqueSlug(ctx, updated.ID, updated.Slug); err != nil {
		return nil, err
	}
	if err := s.ensureUniqueHandle(ctx, updated.ID, updated.Handle); err != nil {
		return nil, err
	}

	updated.UpdatedAt = time.Now().UTC()
	return s.profiles.SaveArtist(ctx, updated)
}

func (s *Service) AdminCreate(ctx context.Context, write domain.AdminArtistWrite) (*domain.ArtistProfile, error) {
	email := strings.TrimSpace(strings.ToLower(write.Email))
	displayName := strings.TrimSpace(write.DisplayName)
	if email == "" || displayName == "" {
		return nil, apperrors.ErrValidation
	}

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}
	if user != nil {
		existing, getErr := s.profiles.GetArtistByUserID(ctx, user.ID)
		if getErr == nil && existing != nil {
			return nil, apperrors.ErrConflict
		}
		if getErr != nil && !errors.Is(getErr, apperrors.ErrNotFound) {
			return nil, getErr
		}
	} else {
		now := time.Now().UTC()
		user = &identity.User{
			ID:        uuid.New(),
			Email:     email,
			Role:      identity.RoleArtist,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := s.users.Create(ctx, *user); err != nil {
			return nil, err
		}
	}

	draft, err := s.profiles.CreateDraftForUser(ctx, user.ID, displayName)
	if err != nil {
		return nil, err
	}
	return s.applyAdminWrite(ctx, draft, write, true)
}

func (s *Service) AdminUpdateContent(ctx context.Context, id uuid.UUID, write domain.AdminArtistWrite) (*domain.ArtistProfile, error) {
	current, err := s.profiles.GetArtistByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.applyAdminWrite(ctx, current, write, false)
}

func (s *Service) AdminDelete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.profiles.GetArtistByID(ctx, id); err != nil {
		return err
	}
	posts, err := s.posts.ListOwnedByArtist(ctx, id)
	if err != nil {
		return err
	}
	if len(posts) > 0 {
		return apperrors.ErrConflict
	}
	return s.profiles.Delete(ctx, id)
}

func (s *Service) AdminUpdate(ctx context.Context, id uuid.UUID, status *domain.ProfileStatus, featured *bool) (*domain.ArtistProfile, error) {
	current, err := s.profiles.GetArtistByID(ctx, id)
	if err != nil {
		return nil, err
	}
	updated := *current
	if status != nil {
		updated.Status = *status
		now := time.Now().UTC()
		if *status == domain.ProfileStatusApproved {
			if updated.ApprovedAt == nil {
				updated.ApprovedAt = &now
			}
		} else {
			updated.ApprovedAt = nil
		}
	}
	if featured != nil {
		updated.Featured = *featured
	}
	updated.UpdatedAt = time.Now().UTC()
	return s.profiles.SaveArtist(ctx, updated)
}

func (s *Service) applyAdminWrite(ctx context.Context, current *domain.ArtistProfile, write domain.AdminArtistWrite, isCreate bool) (*domain.ArtistProfile, error) {
	updated := *current
	if strings.TrimSpace(write.DisplayName) != "" {
		updated.DisplayName = strings.TrimSpace(write.DisplayName)
	}
	if strings.TrimSpace(write.Slug) != "" {
		updated.Slug = slugify(write.Slug)
	} else if isCreate || updated.Slug == "" {
		updated.Slug = slugify(updated.DisplayName)
	}
	if strings.TrimSpace(write.Handle) != "" {
		updated.Handle = slugify(write.Handle)
	} else if updated.Handle == "" {
		updated.Handle = updated.Slug
	}
	updated.Bio = write.Bio
	updated.Born = write.Born
	updated.Discipline = write.Discipline
	updated.Tagline = write.Tagline
	updated.YearsActive = write.YearsActive
	updated.PortraitURL = write.PortraitURL
	if write.Influences != nil {
		updated.Influences = write.Influences
	}
	updated.InResidence = write.InResidence
	updated.ResidencyPlace = write.ResidencyPlace
	updated.OpenForCommission = write.OpenForCommission
	updated.Contact = write.Contact
	if updated.Contact.Email == "" && write.Email != "" {
		updated.Contact.Email = strings.TrimSpace(strings.ToLower(write.Email))
	}
	updated.Social = write.Social
	if write.Status != "" {
		updated.Status = write.Status
		now := time.Now().UTC()
		if write.Status == domain.ProfileStatusApproved {
			if updated.ApprovedAt == nil {
				updated.ApprovedAt = &now
			}
		} else if write.Status != domain.ProfileStatusApproved {
			updated.ApprovedAt = nil
		}
	}

	if err := s.ensureUniqueSlug(ctx, updated.ID, updated.Slug); err != nil {
		return nil, err
	}
	if err := s.ensureUniqueHandle(ctx, updated.ID, updated.Handle); err != nil {
		return nil, err
	}
	updated.UpdatedAt = time.Now().UTC()
	return s.profiles.SaveArtist(ctx, updated)
}

func (s *Service) ensureUniqueSlug(ctx context.Context, profileID uuid.UUID, slug string) error {
	existing, err := s.profiles.GetArtistBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil
		}
		return err
	}
	if existing.ID != profileID {
		return apperrors.ErrConflict
	}
	return nil
}

func (s *Service) ensureUniqueHandle(ctx context.Context, profileID uuid.UUID, handle string) error {
	existing, err := s.profiles.GetArtistByHandle(ctx, handle)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil
		}
		return err
	}
	if existing.ID != profileID {
		return apperrors.ErrConflict
	}
	return nil
}
