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
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateArticleRequest struct {
	Title string `json:"title" binding:"required"`
	Body  string `json:"body"`
}

func ToArticleResponse(a domain.Article) ArticleResponse {
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

type ErrorResponse struct {
	Error string `json:"error"`
}

type HealthResponse struct {
	Status string `json:"status"`
}
