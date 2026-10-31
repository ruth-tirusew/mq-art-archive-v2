// Package questionnaire holds the use cases for the admin-editable artist questionnaire:
// the admin editor, the applicant/artist form, public answers and optional nudges.
package questionnaire

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	domain "github.com/mq/api/internal/domain/questionnaire"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

// maxRichHTMLBytes caps stored rich-text markup regardless of the question's visible-text limit,
// so a tiny text limit cannot hide a huge amount of markup.
const maxRichHTMLBytes = 40_000

// maxAnswersPerRequest bounds one autosave; no realistic questionnaire is near it.
const maxAnswersPerRequest = 100

type Service struct {
	revisions outbound.QuestionnaireRepository
	answers   outbound.QuestionnaireAnswerRepository
	nudges    outbound.QuestionnaireNudgeRepository
	sanitizer outbound.RichTextSanitizer
	users     outbound.UserRepository
	mailer    outbound.Mailer
	webAppURL string
	now       func() time.Time
}

type Option func(*Service)

// WithMailer enables best-effort nudge emails.
func WithMailer(mailer outbound.Mailer, webAppURL string) Option {
	return func(s *Service) {
		s.mailer = mailer
		s.webAppURL = strings.TrimRight(webAppURL, "/")
	}
}

// WithClock overrides the time source (tests).
func WithClock(now func() time.Time) Option {
	return func(s *Service) { s.now = now }
}

func NewService(
	revisions outbound.QuestionnaireRepository,
	answers outbound.QuestionnaireAnswerRepository,
	nudges outbound.QuestionnaireNudgeRepository,
	sanitizer outbound.RichTextSanitizer,
	users outbound.UserRepository,
	opts ...Option,
) *Service {
	s := &Service{
		revisions: revisions, answers: answers, nudges: nudges, sanitizer: sanitizer, users: users,
		now: func() time.Time { return time.Now().UTC() },
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

var (
	_ inbound.QuestionnaireService         = (*Service)(nil)
	_ inbound.QuestionnaireAdminService    = (*Service)(nil)
	_ inbound.ApplicationQuestionnaireGate = (*Service)(nil)
)

// validation turns a domain validation failure into the shared apperrors.ErrValidation.
func validation(err error) error {
	var ve *domain.ValidationError
	if errors.As(err, &ve) {
		return fmt.Errorf("%w: %s", apperrors.ErrValidation, ve.Msg)
	}
	return err
}

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", apperrors.ErrValidation, fmt.Sprintf(format, args...))
}

// ---- Admin editor ----

func (s *Service) GetDraft(ctx context.Context) (*domain.Revision, error) {
	return s.revisions.GetDraft(ctx)
}

func (s *Service) StartDraft(ctx context.Context, adminID uuid.UUID) (*domain.Revision, error) {
	if draft, err := s.revisions.GetDraft(ctx); err == nil {
		return draft, nil
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}

	now := s.now()
	draft := domain.Revision{
		ID: uuid.New(), Status: domain.RevisionDraft, ChangeKind: domain.ChangeSubstantive,
		CreatedBy: &adminID, CreatedAt: now, UpdatedAt: now,
	}
	latest, err := s.revisions.GetLatestPublished(ctx)
	switch {
	case err == nil:
		for _, q := range latest.Questions {
			q.ID, q.RevisionID = uuid.Nil, draft.ID
			draft.Questions = append(draft.Questions, q)
		}
	case !errors.Is(err, apperrors.ErrNotFound):
		return nil, err
	}
	return s.revisions.SaveDraft(ctx, draft)
}

func (s *Service) SaveDraft(ctx context.Context, adminID uuid.UUID, in inbound.DraftInput) (*domain.Revision, error) {
	if in.ChangeKind != domain.ChangeMinor && in.ChangeKind != domain.ChangeSubstantive {
		return nil, invalid("change_kind must be minor or substantive")
	}
	draft, err := s.StartDraft(ctx, adminID)
	if err != nil {
		return nil, err
	}

	rev := domain.Revision{
		ID: draft.ID, Status: domain.RevisionDraft, ChangeKind: in.ChangeKind,
		Note: strings.TrimSpace(in.Note), CreatedBy: draft.CreatedBy, CreatedAt: draft.CreatedAt, UpdatedAt: s.now(),
	}
	for i, qi := range in.Questions {
		maxLen := qi.MaxLength
		if maxLen == 0 {
			maxLen = domain.DefaultMaxLength(qi.Type)
		}
		rev.Questions = append(rev.Questions, domain.Question{
			RevisionID: rev.ID, Key: strings.TrimSpace(qi.Key), Position: i + 1, Type: qi.Type,
			PromptEN: strings.TrimSpace(qi.PromptEN), PromptAM: strings.TrimSpace(qi.PromptAM),
			HelpEN: strings.TrimSpace(qi.HelpEN), HelpAM: strings.TrimSpace(qi.HelpAM),
			Required: qi.Required, IsPublic: qi.IsPublic, MaxLength: maxLen, Options: qi.Options,
		})
	}
	if err := rev.Validate(); err != nil {
		return nil, validation(err)
	}
	if err := s.checkKeyStability(ctx, rev); err != nil {
		return nil, err
	}
	return s.revisions.SaveDraft(ctx, rev)
}

// checkKeyStability enforces that a key keeps meaning the same thing: reusing a published key for
// a different type would silently reinterpret every artist's existing answer.
func (s *Service) checkKeyStability(ctx context.Context, rev domain.Revision) error {
	latest, err := s.revisions.GetLatestPublished(ctx)
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, q := range rev.Questions {
		if old, ok := latest.Question(q.Key); ok && old.Type != q.Type {
			return invalid("question %q was a %s question; changing its type needs a new key", q.Key, old.Type)
		}
	}
	return nil
}

func (s *Service) DiscardDraft(ctx context.Context) error {
	draft, err := s.revisions.GetDraft(ctx)
	if err != nil {
		return err
	}
	return s.revisions.DiscardDraft(ctx, draft.ID)
}

func (s *Service) Publish(ctx context.Context) (*domain.Revision, error) {
	draft, err := s.revisions.GetDraft(ctx)
	if err != nil {
		return nil, err
	}
	if err := draft.Validate(); err != nil {
		return nil, validation(err)
	}
	return s.revisions.Publish(ctx, draft.ID, s.now())
}

func (s *Service) ListRevisions(ctx context.Context) ([]domain.Revision, error) {
	return s.revisions.ListPublished(ctx)
}

func (s *Service) GetRevision(ctx context.Context, id uuid.UUID) (*domain.Revision, error) {
	return s.revisions.GetRevision(ctx, id)
}

// ---- Applicant / artist form ----

func (s *Service) GetForm(ctx context.Context, userID uuid.UUID) (*inbound.Form, error) {
	rev, err := s.revisions.GetLatestPublished(ctx)
	if err != nil {
		return nil, err
	}
	answers, err := s.answers.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	locked, err := s.answers.IsLockedForReview(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &inbound.Form{Revision: *rev, Answers: s.cleanAnswers(answers), Locked: locked}, nil
}

func (s *Service) SaveAnswers(ctx context.Context, userID uuid.UUID, in []inbound.AnswerInput) error {
	if len(in) == 0 {
		return nil
	}
	if len(in) > maxAnswersPerRequest {
		return invalid("too many answers in one request")
	}
	rev, err := s.revisions.GetLatestPublished(ctx)
	if err != nil {
		return err
	}
	locked, err := s.answers.IsLockedForReview(ctx, userID)
	if err != nil {
		return err
	}
	if locked {
		return fmt.Errorf("%w: your application is under review and can't be edited right now", apperrors.ErrConflict)
	}

	now := s.now()
	seen := make(map[string]bool, len(in))
	out := make([]domain.Answer, 0, len(in))
	for _, a := range in {
		if seen[a.Key] {
			return invalid("answer for %q appears twice", a.Key)
		}
		seen[a.Key] = true

		q, ok := rev.Question(a.Key)
		if !ok {
			return invalid("%q is not a question on the current form", a.Key)
		}
		format, lang := a.Format, a.Lang
		if format == "" {
			format = domain.FormatPlain
		}
		if lang == "" {
			lang = domain.LangEnglish
		}

		body := strings.TrimSpace(a.Body)
		measured := body
		if format == domain.FormatRich {
			var text string
			body, text = s.sanitizer.Sanitize(a.Body)
			if len(body) > maxRichHTMLBytes {
				return invalid("answer to %q is too long", a.Key)
			}
			measured = text
		}
		if err := q.CheckAnswer(measured, format, lang); err != nil {
			return validation(err)
		}
		// A link answer ends up as an href on a public page, so it must be a real http(s) address:
		// never javascript:, data: or the like.
		if q.Type == domain.TypeLink && body != "" {
			if err := apperrors.ValidateWebURL(body); err != nil {
				return err
			}
		}
		out = append(out, domain.Answer{
			UserID: userID, QuestionKey: a.Key, RevisionID: rev.ID, Body: body,
			Format: format, Lang: lang, CreatedAt: now, UpdatedAt: now,
		})
	}
	return s.answers.Upsert(ctx, out)
}

// CheckComplete implements inbound.ApplicationQuestionnaireGate.
func (s *Service) CheckComplete(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error) {
	rev, err := s.revisions.GetLatestPublished(ctx)
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	answers, err := s.answers.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if missing := rev.MissingRequired(answers); len(missing) > 0 {
		return nil, invalid("please answer the required questions: %s", strings.Join(missing, ", "))
	}
	return &rev.ID, nil
}

// ---- Public and review views ----

func (s *Service) PublicAnswers(ctx context.Context, userID uuid.UUID) ([]inbound.PublicAnswer, error) {
	rev, err := s.revisions.GetLatestPublished(ctx)
	if errors.Is(err, apperrors.ErrNotFound) {
		return []inbound.PublicAnswer{}, nil
	}
	if err != nil {
		return nil, err
	}
	answers, err := s.answers.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	byKey := indexAnswers(s.cleanAnswers(answers))

	// Only questions in the latest revision are shown, in its order: a question an admin has since
	// removed, or marked private, disappears from the public page without touching stored answers.
	out := []inbound.PublicAnswer{}
	for _, q := range rev.Questions {
		a, ok := byKey[q.Key]
		if !ok || !q.IsPublic || strings.TrimSpace(a.Body) == "" {
			continue
		}
		// Output-side guard, like cleanAnswers: a link row that predates validation or was written
		// by hand must not become a javascript: href.
		if q.Type == domain.TypeLink && !apperrors.IsWebURL(a.Body) {
			continue
		}
		out = append(out, inbound.PublicAnswer{Question: q, Answer: a})
	}
	return out, nil
}

func (s *Service) ReviewView(ctx context.Context, userID uuid.UUID, revisionID *uuid.UUID) (*inbound.ReviewView, error) {
	var rev *domain.Revision
	var err error
	if revisionID != nil {
		rev, err = s.revisions.GetRevision(ctx, *revisionID)
	} else {
		rev, err = s.revisions.GetLatestPublished(ctx)
	}
	if err != nil {
		return nil, err
	}
	answers, err := s.answers.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	answers = s.cleanAnswers(answers)
	byKey := indexAnswers(answers)

	view := &inbound.ReviewView{Revision: *rev, Items: []inbound.ReviewItem{}, Other: []domain.Answer{}}
	inRev := make(map[string]bool, len(rev.Questions))
	for _, q := range rev.Questions {
		inRev[q.Key] = true
		item := inbound.ReviewItem{Question: q}
		if a, ok := byKey[q.Key]; ok {
			a := a
			item.Answer = &a
		}
		view.Items = append(view.Items, item)
	}
	for _, a := range answers {
		if !inRev[a.QuestionKey] {
			view.Other = append(view.Other, a)
		}
	}
	return view, nil
}

// cleanAnswers re-sanitises rich answers on the way out. They were sanitised when written, but
// output is the last line of defence: a row edited by hand, loaded by a script, or saved under an
// older, looser policy must never reach a browser as unsanitised HTML.
func (s *Service) cleanAnswers(answers []domain.Answer) []domain.Answer {
	out := make([]domain.Answer, len(answers))
	for i, a := range answers {
		if a.Format == domain.FormatRich {
			a.Body, _ = s.sanitizer.Sanitize(a.Body)
		}
		out[i] = a
	}
	return out
}

func indexAnswers(answers []domain.Answer) map[string]domain.Answer {
	m := make(map[string]domain.Answer, len(answers))
	for _, a := range answers {
		m[a.QuestionKey] = a
	}
	return m
}

// ---- Nudges ----

func (s *Service) ActiveNudge(ctx context.Context, userID uuid.UUID) (*inbound.NudgePrompt, error) {
	nudge, err := s.nudges.GetActive(ctx, userID)
	if err != nil {
		return nil, err
	}
	rev, err := s.revisions.GetLatestPublished(ctx)
	if err != nil {
		return nil, err
	}
	answers, err := s.answers.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	var questions []domain.Question
	for _, key := range rev.Unanswered(answers) {
		if q, ok := rev.Question(key); ok {
			questions = append(questions, q)
		}
	}
	if len(questions) == 0 {
		// They answered everything on their own; there is nothing left to nudge about.
		_ = s.nudges.Dismiss(ctx, nudge.RevisionID, userID, s.now())
		return nil, apperrors.ErrNotFound
	}
	return &inbound.NudgePrompt{Nudge: *nudge, Questions: questions}, nil
}

func (s *Service) DismissNudge(ctx context.Context, userID, revisionID uuid.UUID) error {
	return s.nudges.Dismiss(ctx, revisionID, userID, s.now())
}

func (s *Service) PreviewNudge(ctx context.Context) (*inbound.NudgePreview, error) {
	rev, err := s.revisions.GetLatestPublished(ctx)
	if err != nil {
		return nil, err
	}
	missing, err := s.nudges.ListUsersMissingAnswers(ctx, rev.ID)
	if err != nil {
		return nil, err
	}
	summary := *rev
	summary.Questions = nil
	return &inbound.NudgePreview{
		Revision: summary, Eligible: len(missing), CanSend: rev.ChangeKind == domain.ChangeSubstantive,
	}, nil
}

func (s *Service) SendNudge(ctx context.Context, adminID uuid.UUID) (int, error) {
	rev, err := s.revisions.GetLatestPublished(ctx)
	if err != nil {
		return 0, err
	}
	if rev.ChangeKind != domain.ChangeSubstantive {
		return 0, invalid("revision %d is a minor change; minor revisions don't nudge artists", *rev.Number)
	}
	missing, err := s.nudges.ListUsersMissingAnswers(ctx, rev.ID)
	if err != nil {
		return 0, err
	}
	created, err := s.nudges.Create(ctx, rev.ID, &adminID, missing, s.now())
	if err != nil {
		return 0, err
	}
	for _, id := range created {
		s.sendNudgeEmail(ctx, id)
	}
	return len(created), nil
}

// sendNudgeEmail is best-effort: the nudge row is already saved and the in-app banner works
// without the email, so a mail outage must not fail the whole send.
func (s *Service) sendNudgeEmail(ctx context.Context, userID uuid.UUID) {
	if s.mailer == nil {
		return
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil || user == nil || user.Email == "" {
		return
	}
	body := fmt.Sprintf("Hi %s,\n\nWe've added a few new questions to the Artiv artist profile. "+
		"There's no deadline and your profile stays exactly as it is. If you'd like to answer them, you can do it here:\n%s/studio\n",
		user.DisplayName, s.webAppURL)
	_ = s.mailer.Send(ctx, user.Email, "New questions for your Artiv profile", body)
}
