package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/engagement"
)

type HighlightRepository interface {
	Create(ctx context.Context, h engagement.Highlight) (*engagement.Highlight, error)
	// Delete removes a highlight, scoped to userID so a reader can only delete their own.
	Delete(ctx context.Context, id, userID uuid.UUID) error
	ListByArticle(ctx context.Context, articleID uuid.UUID) ([]engagement.Highlight, error)
	ListMineByArticle(ctx context.Context, articleID, userID uuid.UUID) ([]engagement.Highlight, error)
	// ListByUser returns a user's highlights across every article, most recent first —
	// what a personal activity feed needs.
	ListByUser(ctx context.Context, userID uuid.UUID) ([]engagement.Highlight, error)
}
