package engagement

import (
	"time"

	"github.com/google/uuid"
)

// Comment is either anchored to a span of an article (a Medium-style margin comment) or
// general — an unanchored "response" at the bottom of the article. IsGeneral is the
// discriminator; Anchor is the zero value when IsGeneral is true.
type Comment struct {
	ID        uuid.UUID
	ArticleID uuid.UUID
	UserID    uuid.UUID
	Body      string
	IsGeneral bool
	Anchor
	CreatedAt time.Time
}
