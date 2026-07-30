package wiki

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

type Submission struct {
	ID          uuid.UUID  `json:"id"`
	SubmitterID uuid.UUID  `json:"submitter_id"`
	ArticleID   *uuid.UUID `json:"article_id,omitempty"`
	Title       string     `json:"title"`
	Body        string     `json:"body"`
	Status      Status     `json:"status"`
	ReviewNotes string     `json:"review_notes,omitempty"`
	ReviewedBy  *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
