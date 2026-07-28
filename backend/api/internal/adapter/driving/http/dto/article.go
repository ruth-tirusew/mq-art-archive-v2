package dto

import (
	"time"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/content"
)

type ArticleResponse struct {
	ID           uuid.UUID `json:"id"`
	Slug         string    `json:"slug"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	Category     string    `json:"category"`
	Excerpt      string    `json:"excerpt,omitempty"`
	ReadingTime  int       `json:"reading_time"`
	Difficulty   string    `json:"difficulty"`
	Verified     bool      `json:"verified"`
	Contributors int       `json:"contributors"`
	AuthorID     uuid.UUID `json:"author_id"`
	Status       string    `json:"status"`
	Version      int       `json:"version"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ArticleRevisionResponse struct {
	ID          uuid.UUID `json:"id"`
	ArticleID   uuid.UUID `json:"article_id"`
	Version     int       `json:"version"`
	EditorID    uuid.UUID `json:"editor_id"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	Slug        string    `json:"slug"`
	Category    string    `json:"category"`
	Excerpt     string    `json:"excerpt,omitempty"`
	ReadingTime int       `json:"reading_time"`
	Difficulty  string    `json:"difficulty"`
	Verified    bool      `json:"verified"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateArticleRequest struct {
	Title string `json:"title" binding:"required"`
	Body  string `json:"body"`
}

type AdminArticleWriteRequest struct {
	Title      string  `json:"title" binding:"required"`
	Body       string  `json:"body"`
	Category   string  `json:"category"`
	Excerpt    string  `json:"excerpt"`
	Difficulty string  `json:"difficulty"`
	Verified   bool    `json:"verified"`
	Status     *string `json:"status"`
}

type AdminPatchArticleRequest struct {
	Status   *string `json:"status"`
	Verified *bool   `json:"verified"`
}

func (r AdminArticleWriteRequest) ToWrite() domain.ArticleWrite {
	write := domain.ArticleWrite{
		Title:      r.Title,
		Body:       r.Body,
		Category:   r.Category,
		Excerpt:    r.Excerpt,
		Difficulty: r.Difficulty,
		Verified:   r.Verified,
	}
	if r.Status != nil {
		s := domain.ArticleStatus(*r.Status)
		write.Status = &s
	}
	return write
}

func ToArticleResponse(a domain.Article) ArticleResponse {
	version := a.Version
	if version <= 0 {
		version = 1
	}
	return ArticleResponse{
		ID:           a.ID,
		Slug:         a.Slug,
		Title:        a.Title,
		Body:         a.Body,
		Category:     a.Category,
		Excerpt:      a.Excerpt,
		ReadingTime:  a.ReadingTime,
		Difficulty:   a.Difficulty,
		Verified:     a.Verified,
		Contributors: a.Contributors,
		AuthorID:     a.AuthorID,
		Status:       string(a.Status),
		Version:      version,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}

func ToArticleResponses(articles []domain.Article) []ArticleResponse {
	out := make([]ArticleResponse, len(articles))
	for i, a := range articles {
		out[i] = ToArticleResponse(a)
	}
	return out
}

func ToArticleRevisionResponse(r domain.ArticleRevision) ArticleRevisionResponse {
	return ArticleRevisionResponse{
		ID:          r.ID,
		ArticleID:   r.ArticleID,
		Version:     r.Version,
		EditorID:    r.EditorID,
		Title:       r.Title,
		Body:        r.Body,
		Slug:        r.Slug,
		Category:    r.Category,
		Excerpt:     r.Excerpt,
		ReadingTime: r.ReadingTime,
		Difficulty:  r.Difficulty,
		Verified:    r.Verified,
		Status:      string(r.Status),
		CreatedAt:   r.CreatedAt,
	}
}

func ToArticleRevisionResponses(revs []domain.ArticleRevision) []ArticleRevisionResponse {
	out := make([]ArticleRevisionResponse, len(revs))
	for i, r := range revs {
		out[i] = ToArticleRevisionResponse(r)
	}
	return out
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type HealthResponse struct {
	Status string `json:"status"`
}
