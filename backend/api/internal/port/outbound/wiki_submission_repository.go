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
	Update(ctx context.Context, submission wiki.Submission) (*wiki.Submission, error)
}
