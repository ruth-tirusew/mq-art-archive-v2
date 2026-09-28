// Package engagement models how readers interact with published content — starting with
// favorites, and later comments/highlights (see the wiki-collaboration plan's Phase 5).
// It is deliberately separate from domain/content: favoriting is about a reader's
// relationship to an article, not the article itself.
package engagement

import (
	"time"

	"github.com/google/uuid"
)

// Favorite records that a user has favorited an article. The pair is unique — favoriting
// twice is a no-op at the repository level (upsert), and unfavoriting removes the row.
type Favorite struct {
	UserID    uuid.UUID
	ArticleID uuid.UUID
	CreatedAt time.Time
}
