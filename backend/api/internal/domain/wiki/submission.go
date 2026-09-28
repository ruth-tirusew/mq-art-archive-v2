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

// Kind distinguishes a submission proposing a brand-new article from one editing an
// existing one. Captured once at submit time — see migration 00034 for why this can't be
// derived later from ArticleID alone.
type Kind string

const (
	KindNew  Kind = "new"
	KindEdit Kind = "edit"
)

type Submission struct {
	ID          uuid.UUID  `json:"id"`
	SubmitterID uuid.UUID  `json:"submitter_id"`
	ArticleID   *uuid.UUID `json:"article_id,omitempty"`
	Title       string     `json:"title"`
	Body        string     `json:"body"`
	Kind        Kind       `json:"kind"`
	// BasedOnVersion is the target article's version at submission time (nil for a
	// new-article submission). Checked against the article's current version on approval
	// so a stale edit can't silently overwrite changes made after the submission.
	BasedOnVersion *int       `json:"based_on_version,omitempty"`
	Status         Status     `json:"status"`
	ReviewNotes    string     `json:"review_notes,omitempty"`
	ReviewedBy     *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewedAt     *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
