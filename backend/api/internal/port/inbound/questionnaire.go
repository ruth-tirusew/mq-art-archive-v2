package inbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/questionnaire"
)

type QuestionInput struct {
	Key       string
	Type      questionnaire.QuestionType
	PromptEN  string
	PromptAM  string
	HelpEN    string
	HelpAM    string
	Required  bool
	IsPublic  bool
	MaxLength int // 0 means the default for the type
	Options   []questionnaire.Option
}

// DraftInput is the whole draft: questions are positioned by their order in the slice.
type DraftInput struct {
	ChangeKind questionnaire.ChangeKind
	Note       string
	Questions  []QuestionInput
}

type AnswerInput struct {
	Key    string
	Body   string
	Format questionnaire.Format
	Lang   questionnaire.Lang
}

// Form is what an applicant or artist fills in: the current revision and their answers so far.
type Form struct {
	Revision questionnaire.Revision
	Answers  []questionnaire.Answer
	// Locked is true while an application is under review; answers cannot change until it is decided.
	Locked bool
}

type PublicAnswer struct {
	Question questionnaire.Question
	Answer   questionnaire.Answer
}

type ReviewItem struct {
	Question questionnaire.Question
	Answer   *questionnaire.Answer // nil when unanswered
}

// ReviewView is what an admin sees next to Approve / Reject. Private answers are included.
type ReviewView struct {
	Revision questionnaire.Revision
	Items    []ReviewItem
	// Other holds answers to questions that are not in Revision (added later, or since removed).
	Other []questionnaire.Answer
}

type NudgePrompt struct {
	Nudge     questionnaire.Nudge
	Questions []questionnaire.Question // still-unanswered questions in the latest revision
}

type NudgePreview struct {
	Revision questionnaire.Revision // latest published, without questions
	Eligible int
	// CanSend is false for a minor revision, which never nudges.
	CanSend bool
}

// QuestionnaireService is the applicant/artist/public side.
type QuestionnaireService interface {
	// GetForm returns the latest published revision with the user's answers, or ErrNotFound
	// before the first revision is published (callers then fall back to the plain application form).
	GetForm(ctx context.Context, userID uuid.UUID) (*Form, error)
	// SaveAnswers autosaves a partial form. It never enforces required answers.
	SaveAnswers(ctx context.Context, userID uuid.UUID, answers []AnswerInput) error
	// PublicAnswers returns the answers a visitor may see, in question order. The caller MUST
	// first confirm the artist's profile is approved; this method does not check profile state.
	PublicAnswers(ctx context.Context, userID uuid.UUID) ([]PublicAnswer, error)
	ActiveNudge(ctx context.Context, userID uuid.UUID) (*NudgePrompt, error)
	DismissNudge(ctx context.Context, userID, revisionID uuid.UUID) error
	// ReviewView shows an application's answers against the revision it was submitted on
	// (revisionID), or the latest published one when nil.
	ReviewView(ctx context.Context, userID uuid.UUID, revisionID *uuid.UUID) (*ReviewView, error)
}

// QuestionnaireAdminService is the admin editor.
type QuestionnaireAdminService interface {
	GetDraft(ctx context.Context) (*questionnaire.Revision, error)
	// StartDraft returns the existing draft, or creates one cloned from the latest published revision.
	StartDraft(ctx context.Context, adminID uuid.UUID) (*questionnaire.Revision, error)
	SaveDraft(ctx context.Context, adminID uuid.UUID, in DraftInput) (*questionnaire.Revision, error)
	DiscardDraft(ctx context.Context) error
	Publish(ctx context.Context) (*questionnaire.Revision, error)
	ListRevisions(ctx context.Context) ([]questionnaire.Revision, error)
	GetRevision(ctx context.Context, id uuid.UUID) (*questionnaire.Revision, error)
	PreviewNudge(ctx context.Context) (*NudgePreview, error)
	// SendNudge emails approved artists with unanswered questions in the latest substantive
	// revision, once per artist per revision. It returns how many were newly nudged.
	SendNudge(ctx context.Context, adminID uuid.UUID) (int, error)
}

// ApplicationQuestionnaireGate lets onboarding require a complete questionnaire without
// depending on the questionnaire context's types.
type ApplicationQuestionnaireGate interface {
	// CheckComplete returns the revision id the applicant has fully answered. It returns
	// (nil, nil) when no questionnaire has been published, and apperrors.ErrValidation naming the
	// missing questions when a required one is unanswered.
	CheckComplete(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error)
}
