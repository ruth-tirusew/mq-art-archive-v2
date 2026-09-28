package inbound

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/content"
)

// ResolvedHighlight is a Highlight re-anchored against the article's current body/version
// — see engagement.Anchor.Resolve for what "resolved" means and why Matched can be false.
type ResolvedHighlight struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	QuotedText string
	Matched    bool
	Start      int
	End        int
	CreatedAt  time.Time
}

// ResolvedComment is a Comment re-anchored the same way, plus the commenter's display name
// (resolved at read time, not snapshotted, so a later name change is reflected).
type ResolvedComment struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	AuthorName string
	Body       string
	IsGeneral  bool
	QuotedText string
	Matched    bool
	Start      int
	End        int
	CreatedAt  time.Time
}

// PopularHighlight is the article's single most-highlighted passage — see
// engagement.MostHighlighted. Nil (from EngagementService.PopularHighlight) when no span
// has reached the minimum reader count to count as "popular".
type PopularHighlight struct {
	Start int
	End   int
	Count int
}

// ActivityHighlight and ActivityComment are a user's own highlight/comment carried with
// enough of its target article's identity (slug/title) to link back to it — what a
// personal "my activity" feed across every article needs. Unlike ResolvedHighlight/
// ResolvedComment, these aren't re-anchored against the article's current body: a
// summary feed just needs to show what was said and where, not render it inline.
type ActivityHighlight struct {
	ID           uuid.UUID
	ArticleID    uuid.UUID
	ArticleSlug  string
	ArticleTitle string
	QuotedText   string
	CreatedAt    time.Time
}

type ActivityComment struct {
	ID           uuid.UUID
	ArticleID    uuid.UUID
	ArticleSlug  string
	ArticleTitle string
	Body         string
	IsGeneral    bool
	CreatedAt    time.Time
}

type EngagementService interface {
	// ToggleFavorite adds the caller's favorite if it doesn't exist, or removes it if it
	// does. Returns the resulting favorited state and the article's new total count.
	// Fails with apperrors.ErrNotFound if the article doesn't exist.
	ToggleFavorite(ctx context.Context, userID, articleID uuid.UUID) (favorited bool, count int, err error)
	// FavoriteStatus returns an article's total favorite count, and whether viewerID (may
	// be uuid.Nil for an unauthenticated viewer, in which case favorited is always false)
	// has favorited it.
	FavoriteStatus(ctx context.Context, viewerID, articleID uuid.UUID) (favorited bool, count int, err error)
	// ListMyFavorites returns the articles the caller has favorited, most recent first.
	ListMyFavorites(ctx context.Context, userID uuid.UUID) ([]content.Article, error)

	// CreateHighlight anchors a new personal highlight for userID on articleID. Fails with
	// apperrors.ErrNotFound if the article doesn't exist, apperrors.ErrValidation if
	// quotedText/start/end don't describe a real, in-bounds span of the article's current
	// body.
	CreateHighlight(ctx context.Context, userID, articleID uuid.UUID, quotedText string, start, end int) (*ResolvedHighlight, error)
	// DeleteHighlight removes userID's own highlight; fails with apperrors.ErrNotFound if
	// it doesn't exist or isn't theirs.
	DeleteHighlight(ctx context.Context, userID, highlightID uuid.UUID) error
	// ListMyHighlights returns userID's own highlights on articleID, resolved against its
	// current body/version.
	ListMyHighlights(ctx context.Context, userID, articleID uuid.UUID) ([]ResolvedHighlight, error)
	// PopularHighlight returns the article's most-highlighted passage across all readers,
	// or nil if none has reached the minimum reader count to surface as "popular".
	PopularHighlight(ctx context.Context, articleID uuid.UUID) (*PopularHighlight, error)

	// CreateComment adds a comment for userID on articleID — anchored to a span
	// (quotedText/start/end, isGeneral false) or a general "response" (isGeneral true,
	// quotedText/start/end ignored). Fails with apperrors.ErrNotFound if the article
	// doesn't exist, apperrors.ErrValidation for an empty body or an anchored comment
	// whose span doesn't describe real, in-bounds text.
	CreateComment(ctx context.Context, userID, articleID uuid.UUID, body, quotedText string, start, end int, isGeneral bool) (*ResolvedComment, error)
	// DeleteComment removes userID's own comment; fails with apperrors.ErrNotFound if it
	// doesn't exist or isn't theirs.
	DeleteComment(ctx context.Context, userID, commentID uuid.UUID) error
	// ListComments returns all of an article's comments (anchored and general), resolved
	// against its current body/version, oldest first.
	ListComments(ctx context.Context, articleID uuid.UUID) ([]ResolvedComment, error)

	// ListMyHighlightsAll and ListMyCommentsAll return userID's own highlights/comments
	// across every article, most recent first — the personal activity feed.
	ListMyHighlightsAll(ctx context.Context, userID uuid.UUID) ([]ActivityHighlight, error)
	ListMyCommentsAll(ctx context.Context, userID uuid.UUID) ([]ActivityComment, error)
}
