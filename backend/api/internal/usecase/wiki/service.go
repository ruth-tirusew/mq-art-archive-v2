package wiki

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/content"
	domain "github.com/mq/api/internal/domain/wiki"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

type Service struct {
	submissions outbound.WikiSubmissionRepository
	content     inbound.ContentService
}

func NewService(submissions outbound.WikiSubmissionRepository, contentService inbound.ContentService) inbound.WikiSubmissionService {
	return &Service{submissions: submissions, content: contentService}
}

func (s *Service) Submit(ctx context.Context, submitterID uuid.UUID, articleID *uuid.UUID, title, body string) (*domain.Submission, error) {
	title = strings.TrimSpace(title)
	if title == "" || strings.TrimSpace(body) == "" {
		return nil, fmt.Errorf("%w: title and body are required", apperrors.ErrValidation)
	}

	kind := domain.KindNew
	var basedOnVersion *int
	if articleID != nil {
		kind = domain.KindEdit
		article, err := s.content.AdminGet(ctx, *articleID)
		if err != nil {
			return nil, err
		}
		version := article.Version
		basedOnVersion = &version
	}

	now := time.Now().UTC()
	return s.submissions.Create(ctx, domain.Submission{
		ID: uuid.New(), SubmitterID: submitterID, ArticleID: articleID, Title: title, Body: body,
		Kind: kind, BasedOnVersion: basedOnVersion,
		Status: domain.StatusPending, CreatedAt: now, UpdatedAt: now,
	})
}

func (s *Service) ListMine(ctx context.Context, id uuid.UUID) ([]domain.Submission, error) {
	return s.submissions.ListBySubmitter(ctx, id)
}

func (s *Service) ListPending(ctx context.Context) ([]domain.Submission, error) {
	return s.submissions.ListPending(ctx)
}

// ListOwnPending is the artist-facing counterpart to ListPending: submissions awaiting
// review against an article ownerID authors, rather than the admin-wide queue.
func (s *Service) ListOwnPending(ctx context.Context, ownerID uuid.UUID) ([]domain.Submission, error) {
	return s.submissions.ListPendingForOwner(ctx, ownerID)
}

func (s *Service) Approve(ctx context.Context, id, reviewerID uuid.UUID, notes string) (*domain.Submission, error) {
	return s.approve(ctx, id, reviewerID, notes, uuid.Nil)
}

// ApproveOwn is Approve restricted to an artist reviewing a submission against an article
// they own: it fails with apperrors.ErrForbidden unless the submission is an edit
// (ownerID has nothing pre-existing to own for a new-article submission) targeting an
// article whose AuthorID matches ownerID.
func (s *Service) ApproveOwn(ctx context.Context, id, ownerID uuid.UUID, notes string) (*domain.Submission, error) {
	return s.approve(ctx, id, ownerID, notes, ownerID)
}

// approve is shared by Approve (admin, unrestricted) and ApproveOwn (article-owning
// artist, restricted). requireOwner is uuid.Nil for the unrestricted admin path.
func (s *Service) approve(ctx context.Context, id, reviewerID uuid.UUID, notes string, requireOwner uuid.UUID) (*domain.Submission, error) {
	submission, err := s.submissions.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if submission.Status != domain.StatusPending {
		return nil, apperrors.ErrConflict
	}

	published := content.ArticleStatusPublished
	write := content.ArticleWrite{Title: submission.Title, Body: submission.Body, Status: &published}

	if submission.ArticleID == nil {
		if requireOwner != uuid.Nil {
			return nil, apperrors.ErrForbidden
		}
		article, err := s.content.AdminCreate(ctx, submission.SubmitterID, write)
		if err != nil {
			return nil, err
		}
		submission.ArticleID = &article.ID
	} else {
		if requireOwner != uuid.Nil {
			article, err := s.content.AdminGet(ctx, *submission.ArticleID)
			if err != nil {
				return nil, err
			}
			if article.AuthorID != requireOwner {
				return nil, apperrors.ErrForbidden
			}
		}
		if submission.BasedOnVersion != nil {
			if _, err := s.content.ApproveEdit(ctx, *submission.ArticleID, submission.SubmitterID, *submission.BasedOnVersion, submission.ID, write); err != nil {
				return nil, err
			}
		} else {
			// Legacy submission created before migration 00034 tracked based_on_version —
			// no version to check the article against, so fall back to the old
			// unconditional overwrite rather than refusing to approve it at all.
			if _, err := s.content.AdminUpdate(ctx, *submission.ArticleID, submission.SubmitterID, write); err != nil {
				return nil, err
			}
		}
	}
	return s.review(ctx, submission, reviewerID, domain.StatusApproved, notes)
}

func (s *Service) Reject(ctx context.Context, id, reviewerID uuid.UUID, notes string) (*domain.Submission, error) {
	return s.reject(ctx, id, reviewerID, notes, uuid.Nil)
}

// RejectOwn mirrors ApproveOwn's ownership restriction for rejecting a submission.
func (s *Service) RejectOwn(ctx context.Context, id, ownerID uuid.UUID, notes string) (*domain.Submission, error) {
	return s.reject(ctx, id, ownerID, notes, ownerID)
}

func (s *Service) reject(ctx context.Context, id, reviewerID uuid.UUID, notes string, requireOwner uuid.UUID) (*domain.Submission, error) {
	submission, err := s.submissions.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if submission.Status != domain.StatusPending {
		return nil, apperrors.ErrConflict
	}
	if requireOwner != uuid.Nil {
		if submission.ArticleID == nil {
			return nil, apperrors.ErrForbidden
		}
		article, err := s.content.AdminGet(ctx, *submission.ArticleID)
		if err != nil {
			return nil, err
		}
		if article.AuthorID != requireOwner {
			return nil, apperrors.ErrForbidden
		}
	}
	return s.review(ctx, submission, reviewerID, domain.StatusRejected, notes)
}

func (s *Service) GetForReview(ctx context.Context, id uuid.UUID) (*domain.Submission, *content.Article, error) {
	return s.getForReview(ctx, id, uuid.Nil)
}

func (s *Service) GetOwnForReview(ctx context.Context, id, ownerID uuid.UUID) (*domain.Submission, *content.Article, error) {
	return s.getForReview(ctx, id, ownerID)
}

// getForReview loads a submission and, for an edit submission, the current state of its
// target article — the pair a reviewer needs to render a diff. requireOwner is uuid.Nil
// for the unrestricted admin path, mirroring approve/reject above.
func (s *Service) getForReview(ctx context.Context, id, requireOwner uuid.UUID) (*domain.Submission, *content.Article, error) {
	submission, err := s.submissions.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if submission.ArticleID == nil {
		if requireOwner != uuid.Nil {
			return nil, nil, apperrors.ErrForbidden
		}
		return submission, nil, nil
	}
	article, err := s.content.AdminGet(ctx, *submission.ArticleID)
	if err != nil {
		return nil, nil, err
	}
	if requireOwner != uuid.Nil && article.AuthorID != requireOwner {
		return nil, nil, apperrors.ErrForbidden
	}
	return submission, article, nil
}

func (s *Service) review(ctx context.Context, submission *domain.Submission, reviewerID uuid.UUID, status domain.Status, notes string) (*domain.Submission, error) {
	now := time.Now().UTC()
	submission.Status, submission.ReviewNotes = status, notes
	submission.ReviewedBy, submission.ReviewedAt = &reviewerID, &now
	submission.UpdatedAt = now
	return s.submissions.Update(ctx, *submission)
}
