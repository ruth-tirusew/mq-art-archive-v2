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

	return s.articles.Update(ctx, *article)
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
	version := article.Version
	if version <= 0 {
		version = 1
	}
	return s.articles.InsertRevision(ctx, domain.ArticleRevision{
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
	})
}
