package content

import (
	"time"

	"github.com/google/uuid"
)

type ArticleStatus string

const (
	ArticleStatusDraft     ArticleStatus = "draft"
	ArticleStatusPublished ArticleStatus = "published"
	ArticleStatusArchived  ArticleStatus = "archived"
)

// ArticleRevision is an immutable snapshot of an article at a given version.
type ArticleRevision struct {
	ID          uuid.UUID
	ArticleID   uuid.UUID
	Version     int
	EditorID    uuid.UUID
	Title       string
	Body        string
	Slug        string
	Category    string
	Excerpt     string
	ReadingTime int
	Difficulty  string
	Verified    bool
	Status      ArticleStatus
	CreatedAt   time.Time
}

type Article struct {
	ID           uuid.UUID
	Slug         string
	Title        string
	Body         string
	Category     string
	Excerpt      string
	ReadingTime  int
	Difficulty   string
	Verified     bool
	Contributors int
	AuthorID     uuid.UUID
	Status       ArticleStatus
	Version      int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
