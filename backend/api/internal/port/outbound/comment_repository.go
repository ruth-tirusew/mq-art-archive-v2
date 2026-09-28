package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/engagement"
)

type CommentRepository interface {
	Create(ctx context.Context, c engagement.Comment) (*engagement.Comment, error)
	// Delete removes a comment, scoped to userID so a reader can only delete their own.
	Delete(ctx context.Context, id, userID uuid.UUID) error
	ListByArticle(ctx context.Context, articleID uuid.UUID) ([]engagement.Comment, error)
}
