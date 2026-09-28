package inbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/domain/wiki"
)

type WikiSubmissionService interface {
	Submit(ctx context.Context, submitterID uuid.UUID, articleID *uuid.UUID, title, body string) (*wiki.Submission, error)
	ListMine(ctx context.Context, submitterID uuid.UUID) ([]wiki.Submission, error)
	ListPending(ctx context.Context) ([]wiki.Submission, error)
	// ListOwnPending is the artist-facing counterpart to ListPending, scoped to submissions
	// awaiting review against an article ownerID authors.
	ListOwnPending(ctx context.Context, ownerID uuid.UUID) ([]wiki.Submission, error)
	Approve(ctx context.Context, id, reviewerID uuid.UUID, notes string) (*wiki.Submission, error)
	Reject(ctx context.Context, id, reviewerID uuid.UUID, notes string) (*wiki.Submission, error)
	// ApproveOwn and RejectOwn let an artist review submissions against articles they
	// author, without the admin role: they fail with apperrors.ErrForbidden for a
	// new-article submission (nothing pre-existing to own) or one whose article isn't
	// authored by ownerID.
	ApproveOwn(ctx context.Context, id, ownerID uuid.UUID, notes string) (*wiki.Submission, error)
	RejectOwn(ctx context.Context, id, ownerID uuid.UUID, notes string) (*wiki.Submission, error)
	// GetForReview and GetOwnForReview fetch a submission alongside the current state of
	// its target article (nil for a new-article submission), so a caller can render a diff
	// between what's proposed and what's live. GetOwnForReview enforces the same ownership
	// check as ApproveOwn/RejectOwn; GetForReview (admin) does not.
	GetForReview(ctx context.Context, id uuid.UUID) (*wiki.Submission, *content.Article, error)
	GetOwnForReview(ctx context.Context, id, ownerID uuid.UUID) (*wiki.Submission, *content.Article, error)
}
