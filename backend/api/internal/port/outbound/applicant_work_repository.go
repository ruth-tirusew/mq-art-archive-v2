package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/onboarding"
)

// ApplicantWorkRepository stores the work an applicant has uploaded for their application.
type ApplicantWorkRepository interface {
	// ListByUser returns the user's works in order, each with its image's URL and blur hash.
	ListByUser(ctx context.Context, userID uuid.UUID) ([]onboarding.ApplicantWork, error)
	// Add stores a work for an image the user uploaded. apperrors.ErrNotFound if the image does not
	// exist or is not theirs (the two are indistinguishable on purpose); ErrConflict if that image is
	// already one of their works.
	Add(ctx context.Context, work onboarding.ApplicantWork) (*onboarding.ApplicantWork, error)
	// Delete removes one of the user's works. ErrNotFound if it isn't theirs.
	Delete(ctx context.Context, userID, workID uuid.UUID) error

	// GetPortrait returns the user's portrait, or ErrNotFound if they have none. Portraits are
	// kept apart from works: ListByUser never returns one.
	GetPortrait(ctx context.Context, userID uuid.UUID) (*onboarding.ApplicantPortrait, error)
	// SetPortrait makes an image the user uploaded their portrait, replacing any earlier one.
	// ErrNotFound if the image does not exist or is not theirs; ErrConflict if it is already one
	// of their works.
	SetPortrait(ctx context.Context, userID, mediaAssetID uuid.UUID) (*onboarding.ApplicantPortrait, error)
	// DeletePortrait removes the user's portrait. ErrNotFound if they have none.
	DeletePortrait(ctx context.Context, userID uuid.UUID) error
}
