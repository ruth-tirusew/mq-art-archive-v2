package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/wiki"
)

type WikiSubmissionRepository interface {
	Create(ctx context.Context, submission wiki.Submission) (*wiki.Submission, error)
	GetByID(ctx context.Context, id uuid.UUID) (*wiki.Submission, error)
	ListBySubmitter(ctx context.Context, submitterID uuid.UUID) ([]wiki.Submission, error)
	ListPending(ctx context.Context) ([]wiki.Submission, error)
	// ListPendingForOwner lists pending edit submissions targeting an article authored by
	// ownerID — the queue an artist reviews, as opposed to ListPending's admin-wide queue.
	// New-article submissions never appear here (nothing pre-existing to own).
	ListPendingForOwner(ctx context.Context, ownerID uuid.UUID) ([]wiki.Submission, error)
	Update(ctx context.Context, submission wiki.Submission) (*wiki.Submission, error)
}
