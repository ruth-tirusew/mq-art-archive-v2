package content

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	domain "github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

type Service struct {
	articles outbound.ArticleRepository
}

func NewService(articles outbound.ArticleRepository) inbound.ContentService {
	return &Service{articles: articles}
}

func (s *Service) ListPublished(ctx context.Context, filter domain.ListFilter) ([]domain.Article, error) {
	return s.articles.ListPublished(ctx, filter)
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*domain.Article, error) {
	return s.articles.GetBySlug(ctx, slug)
}

func (s *Service) CreateDraft(ctx context.Context, authorID uuid.UUID, title, body string) (*domain.Article, error) {
	return s.AdminCreate(ctx, authorID, domain.ArticleWrite{
		Title: title,
		Body:  body,
	})
}

func (s *Service) AdminList(ctx context.Context, status *domain.ArticleStatus, limit, offset int) ([]domain.Article, error) {
	return s.articles.ListAdmin(ctx, status, limit, offset)
}

func (s *Service) AdminGet(ctx context.Context, id uuid.UUID) (*domain.Article, error) {
	return s.articles.GetByID(ctx, id)
}

func (s *Service) AdminCreate(ctx context.Context, authorID uuid.UUID, write domain.ArticleWrite) (*domain.Article, error) {
	if err := requireTitle(write.Title); err != nil {
		return nil, fmt.Errorf("%w: %s", apperrors.ErrValidation, err.Error())
	}

	status := domain.ArticleStatusDraft
	if write.Status != nil {
		if !validStatus(*write.Status) {
			return nil, fmt.Errorf("%w: invalid status", apperrors.ErrValidation)
		}
		status = *write.Status
	}

	now := time.Now().UTC()
	article := domain.Article{
		ID:           uuid.New(),
		Slug:         slugify(write.Title),
		Title:        strings.TrimSpace(write.Title),
		Body:         write.Body,
		Category:     resolveCategory(write.Category),
		Excerpt:      write.Excerpt,
		ReadingTime:  estimateReadingTime(write.Body),
		Difficulty:   resolveDifficulty(write.Difficulty),
		Verified:     write.Verified,
		Contributors: 1,
		AuthorID:     authorID,
		Status:       status,
		Version:      1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if status == domain.ArticleStatusPublished {
		article.PublishedAt = &now
	}

	return s.articles.Create(ctx, article)
}

func (s *Service) AdminUpdate(ctx context.Context, id, editorID uuid.UUID, write domain.ArticleWrite) (*domain.Article, error) {
	if err := requireTitle(write.Title); err != nil {
		return nil, fmt.Errorf("%w: %s", apperrors.ErrValidation, err.Error())
	}

	article, err := s.articles.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.snapshotCurrent(ctx, article, editorID); err != nil {
		return nil, err
	}

	applyWrite(article, write)

	return s.articles.Update(ctx, *article)
}

// ApproveEdit is AdminUpdate with an optimistic-concurrency check for the wiki submission
// approval path — see inbound.ContentService for why this needs to be a separate method
// rather than an extra AdminUpdate parameter (AdminUpdate is also used by direct admin
// edits and revision restores, neither of which have a submission version to check against).
func (s *Service) ApproveEdit(ctx context.Context, id, editorID uuid.UUID, expectedVersion int, submissionID uuid.UUID, write domain.ArticleWrite) (*domain.Article, error) {
	if err := requireTitle(write.Title); err != nil {
		return nil, fmt.Errorf("%w: %s", apperrors.ErrValidation, err.Error())
	}

	article, err := s.articles.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if article.Version != expectedVersion {
		// Fails fast on the common case without attempting a write; the atomic check in
		// UpdateForApproval is what actually prevents the race (this article could still
		// change between this read and that write).
		return nil, apperrors.ErrConflict
	}

	revision := buildRevision(article, editorID)
	revision.SubmissionID = &submissionID

	applyWrite(article, write)

	return s.articles.UpdateForApproval(ctx, *article, expectedVersion, revision)
}

// applyWrite mutates article in place with the fields from write, matching what
// AdminUpdate and ApproveEdit both need to do before persisting — slug regeneration only
// applies to drafts (a published article's URL shouldn't move out from under existing
// links), version/timestamps/publish-date bookkeeping is otherwise identical either way.
func applyWrite(article *domain.Article, write domain.ArticleWrite) {
	title := strings.TrimSpace(write.Title)
	if article.Status == domain.ArticleStatusDraft {
		if strings.TrimSpace(write.Slug) != "" {
			article.Slug = slugify(write.Slug)
		} else if title != article.Title {
			article.Slug = slugify(title)
		}
	}

	article.Title = title
	article.Body = write.Body
	article.Category = resolveCategory(write.Category)
	article.Excerpt = write.Excerpt
	article.ReadingTime = estimateReadingTime(write.Body)
	article.Difficulty = resolveDifficulty(write.Difficulty)
	article.Verified = write.Verified
	if write.Status != nil && validStatus(*write.Status) {
		article.Status = *write.Status
	}
	article.Version++
	article.UpdatedAt = time.Now().UTC()
	if article.Status == domain.ArticleStatusPublished && article.PublishedAt == nil {
		article.PublishedAt = &article.UpdatedAt
	}
}

func (s *Service) AdminSetStatus(ctx context.Context, id uuid.UUID, status *domain.ArticleStatus, verified *bool) (*domain.Article, error) {
	article, err := s.articles.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if status != nil {
		if !validStatus(*status) {
			return nil, fmt.Errorf("%w: invalid status", apperrors.ErrValidation)
		}
		article.Status = *status
	}
	if verified != nil {
		article.Verified = *verified
	}
	article.UpdatedAt = time.Now().UTC()
	if article.Status == domain.ArticleStatusPublished && article.PublishedAt == nil {
		article.PublishedAt = &article.UpdatedAt
	}
	return s.articles.Update(ctx, *article)
}

func (s *Service) AdminListRevisions(ctx context.Context, articleID uuid.UUID, limit, offset int) ([]domain.ArticleRevision, error) {
	if _, err := s.articles.GetByID(ctx, articleID); err != nil {
		return nil, err
	}
	return s.articles.ListRevisions(ctx, articleID, limit, offset)
}

func (s *Service) AdminGetRevision(ctx context.Context, articleID uuid.UUID, version int) (*domain.ArticleRevision, error) {
	if _, err := s.articles.GetByID(ctx, articleID); err != nil {
		return nil, err
	}
	return s.articles.GetRevision(ctx, articleID, version)
}

func (s *Service) AdminRestoreRevision(ctx context.Context, articleID uuid.UUID, version int, editorID uuid.UUID) (*domain.Article, error) {
	rev, err := s.AdminGetRevision(ctx, articleID, version)
	if err != nil {
		return nil, err
	}
	status := rev.Status
	return s.AdminUpdate(ctx, articleID, editorID, domain.ArticleWrite{
		Title:      rev.Title,
		Body:       rev.Body,
		Slug:       rev.Slug,
		Category:   rev.Category,
		Excerpt:    rev.Excerpt,
		Difficulty: rev.Difficulty,
		Verified:   rev.Verified,
		Status:     &status,
	})
}

func (s *Service) snapshotCurrent(ctx context.Context, article *domain.Article, editorID uuid.UUID) error {
	return s.articles.InsertRevision(ctx, buildRevision(article, editorID))
}

// buildRevision captures article's current (pre-mutation) state as a revision. Callers
// apply it either immediately (snapshotCurrent, for AdminUpdate) or as part of a larger
// transaction (ApproveEdit, via UpdateForApproval) — hence returning the value rather than
// inserting it directly.
func buildRevision(article *domain.Article, editorID uuid.UUID) domain.ArticleRevision {
	version := article.Version
	if version <= 0 {
		version = 1
	}
	return domain.ArticleRevision{
		ID:          uuid.New(),
		ArticleID:   article.ID,
		Version:     version,
		EditorID:    editorID,
		Title:       article.Title,
		Body:        article.Body,
		Slug:        article.Slug,
		Category:    article.Category,
		Excerpt:     article.Excerpt,
		ReadingTime: article.ReadingTime,
		Difficulty:  article.Difficulty,
		Verified:    article.Verified,
		Status:      article.Status,
		CreatedAt:   time.Now().UTC(),
	}
}
