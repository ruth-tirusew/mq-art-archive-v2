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

type ContributionStatus string

const (
	ContributionPending  ContributionStatus = "pending"
	ContributionAccepted ContributionStatus = "accepted"
	ContributionRejected ContributionStatus = "rejected"
)

type Revision struct {
	ID        uuid.UUID
	ArticleID uuid.UUID
	AuthorID  uuid.UUID
	Body      string
	Status    ContributionStatus
	CreatedAt time.Time
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
	Revisions    []Revision
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
