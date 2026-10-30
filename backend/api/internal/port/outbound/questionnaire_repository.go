package outbound

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/questionnaire"
)

// QuestionnaireRepository stores questionnaire revisions and their questions.
// Not-found cases return apperrors.ErrNotFound; violated invariants return apperrors.ErrConflict.
type QuestionnaireRepository interface {
	// GetDraft returns the single working draft, or ErrNotFound when there is none.
	GetDraft(ctx context.Context) (*questionnaire.Revision, error)
	// GetLatestPublished returns the newest published revision (with questions), or ErrNotFound
	// before the first publish.
	GetLatestPublished(ctx context.Context) (*questionnaire.Revision, error)
	// GetRevision returns any revision by id, with its questions.
	GetRevision(ctx context.Context, id uuid.UUID) (*questionnaire.Revision, error)
	// ListPublished returns published revisions newest first, WITHOUT their questions.
	ListPublished(ctx context.Context) ([]questionnaire.Revision, error)
	// SaveDraft inserts or replaces the draft, replacing its whole question list. ErrConflict if
	// the revision is already published (published revisions are immutable) or a different draft
	// already exists.
	SaveDraft(ctx context.Context, draft questionnaire.Revision) (*questionnaire.Revision, error)
	// DiscardDraft deletes a draft. ErrNotFound if no draft has that id.
	DiscardDraft(ctx context.Context, id uuid.UUID) error
	// Publish validates the draft, assigns the next revision number and freezes it, atomically.
	// ErrConflict if it is not a draft; apperrors.ErrValidation if it is not a valid revision.
	Publish(ctx context.Context, id uuid.UUID, now time.Time) (*questionnaire.Revision, error)
}

// QuestionnaireAnswerRepository stores artists' answers, one current answer per question key.
type QuestionnaireAnswerRepository interface {
	ListByUser(ctx context.Context, userID uuid.UUID) ([]questionnaire.Answer, error)
	ListByUsers(ctx context.Context, userIDs []uuid.UUID) ([]questionnaire.Answer, error)
	// Upsert writes the answers atomically (all or none), keyed by user and question key.
	Upsert(ctx context.Context, answers []questionnaire.Answer) error
}

// QuestionnaireNudgeRepository tracks optional, admin-triggered nudges.
type QuestionnaireNudgeRepository interface {
	// ListUsersMissingAnswers returns approved artists who have no non-blank answer for at least
	// one question in the revision: the people a nudge for it would be about.
	ListUsersMissingAnswers(ctx context.Context, revisionID uuid.UUID) ([]uuid.UUID, error)
	// Create records nudges and returns only the user ids newly nudged, so a repeat call never
	// re-notifies someone who already has a nudge for that revision.
	Create(ctx context.Context, revisionID uuid.UUID, sentBy *uuid.UUID, userIDs []uuid.UUID, now time.Time) ([]uuid.UUID, error)
	// GetActive returns the user's most recent undismissed nudge, or ErrNotFound.
	GetActive(ctx context.Context, userID uuid.UUID) (*questionnaire.Nudge, error)
	// Dismiss marks a nudge dismissed. ErrNotFound if the user has no such nudge.
	Dismiss(ctx context.Context, revisionID, userID uuid.UUID, now time.Time) error
}
