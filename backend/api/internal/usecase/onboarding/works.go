package onboarding

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	domain "github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/port/inbound"
)

func (s *Service) requireWorks() error {
	if s.works == nil {
		return apperrors.ErrNotImplemented
	}
	return nil
}

// lockedForReview is true while the applicant's latest application is pending: what they sent is
// what the reviewer sees, so it must not change underneath them.
func (s *Service) lockedForReview(ctx context.Context, userID uuid.UUID) error {
	latest, err := s.applications.GetLatestByApplicantID(ctx, userID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if latest.Status == domain.ApprovalStatusPending {
		return fmt.Errorf("%w: your application is being reviewed, so your works can't change until it is decided", apperrors.ErrConflict)
	}
	return nil
}

func (s *Service) ListWorks(ctx context.Context, userID uuid.UUID) ([]domain.ApplicantWork, error) {
	if err := s.requireWorks(); err != nil {
		return nil, err
	}
	return s.works.ListByUser(ctx, userID)
}

func (s *Service) AddWork(ctx context.Context, userID uuid.UUID, in inbound.AddWorkInput) (*domain.ApplicantWork, error) {
	if err := s.requireWorks(); err != nil {
		return nil, err
	}
	work := domain.ApplicantWork{
		ID: uuid.New(), UserID: userID, MediaAssetID: in.MediaAssetID,
		Title: strings.TrimSpace(in.Title), Medium: strings.TrimSpace(in.Medium), Year: in.Year, CreatedAt: s.now(),
	}
	if err := work.Validate(s.now()); err != nil {
		return nil, fmt.Errorf("%w: %s", apperrors.ErrValidation, err.Error())
	}
	if err := s.lockedForReview(ctx, userID); err != nil {
		return nil, err
	}
	existing, err := s.works.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(existing) >= domain.MaxApplicationWorks {
		return nil, fmt.Errorf("%w: you can add up to %d works", apperrors.ErrValidation, domain.MaxApplicationWorks)
	}
	return s.works.Add(ctx, work)
}

func (s *Service) RemoveWork(ctx context.Context, userID, workID uuid.UUID) error {
	if err := s.requireWorks(); err != nil {
		return err
	}
	if err := s.lockedForReview(ctx, userID); err != nil {
		return err
	}
	return s.works.Delete(ctx, userID, workID)
}
