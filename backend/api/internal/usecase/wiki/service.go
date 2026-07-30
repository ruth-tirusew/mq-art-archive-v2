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
	now := time.Now().UTC()
	return s.submissions.Create(ctx, domain.Submission{
		ID: uuid.New(), SubmitterID: submitterID, ArticleID: articleID, Title: title, Body: body,
		Status: domain.StatusPending, CreatedAt: now, UpdatedAt: now,
	})
}

func (s *Service) ListMine(ctx context.Context, id uuid.UUID) ([]domain.Submission, error) {
	return s.submissions.ListBySubmitter(ctx, id)
}

func (s *Service) ListPending(ctx context.Context) ([]domain.Submission, error) {
	return s.submissions.ListPending(ctx)
}

func (s *Service) Approve(ctx context.Context, id, reviewerID uuid.UUID, notes string) (*domain.Submission, error) {
	submission, err := s.submissions.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if submission.Status != domain.StatusPending {
		return nil, apperrors.ErrConflict
	}
	published := content.ArticleStatusPublished
	if submission.ArticleID == nil {
		article, err := s.content.AdminCreate(ctx, reviewerID, content.ArticleWrite{
			Title: submission.Title, Body: submission.Body, Status: &published,
		})
		if err != nil {
			return nil, err
		}
		submission.ArticleID = &article.ID
	} else {
		if _, err := s.content.AdminUpdate(ctx, *submission.ArticleID, reviewerID, content.ArticleWrite{
			Title: submission.Title, Body: submission.Body, Status: &published,
		}); err != nil {
			return nil, err
		}
	}
	return s.review(ctx, submission, reviewerID, domain.StatusApproved, notes)
}

func (s *Service) Reject(ctx context.Context, id, reviewerID uuid.UUID, notes string) (*domain.Submission, error) {
	submission, err := s.submissions.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if submission.Status != domain.StatusPending {
		return nil, apperrors.ErrConflict
	}
	return s.review(ctx, submission, reviewerID, domain.StatusRejected, notes)
}

func (s *Service) review(ctx context.Context, submission *domain.Submission, reviewerID uuid.UUID, status domain.Status, notes string) (*domain.Submission, error) {
	now := time.Now().UTC()
	submission.Status, submission.ReviewNotes = status, notes
	submission.ReviewedBy, submission.ReviewedAt = &reviewerID, &now
	submission.UpdatedAt = now
	return s.submissions.Update(ctx, *submission)
}
