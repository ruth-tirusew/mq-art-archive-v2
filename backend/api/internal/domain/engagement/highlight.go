package engagement

import (
	"time"

	"github.com/google/uuid"
)

// Highlight is a reader's personal mark on a span of an article, anchored per Anchor.
type Highlight struct {
	ID        uuid.UUID
	ArticleID uuid.UUID
	UserID    uuid.UUID
	Anchor
	CreatedAt time.Time
}
