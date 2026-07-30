package inbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/wiki"
)

type WikiSubmissionService interface {
	Submit(ctx context.Context, submitterID uuid.UUID, articleID *uuid.UUID, title, body string) (*wiki.Submission, error)
	ListMine(ctx context.Context, submitterID uuid.UUID) ([]wiki.Submission, error)
	ListPending(ctx context.Context) ([]wiki.Submission, error)
	Approve(ctx context.Context, id, reviewerID uuid.UUID, notes string) (*wiki.Submission, error)
	Reject(ctx context.Context, id, reviewerID uuid.UUID, notes string) (*wiki.Submission, error)
}
