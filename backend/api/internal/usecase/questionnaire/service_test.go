package questionnaire

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driven/richtext"
	"github.com/mq/api/internal/domain/apperrors"
	onboardingdomain "github.com/mq/api/internal/domain/onboarding"
	profiledomain "github.com/mq/api/internal/domain/profile"
	domain "github.com/mq/api/internal/domain/questionnaire"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/testutil/assist"
)

type fakeRevisions struct {
	draft     *domain.Revision
	published []*domain.Revision
}

func (f *fakeRevisions) GetDraft(ctx context.Context) (*domain.Revision, error) {
	if f.draft == nil {
		return nil, apperrors.ErrNotFound
	}
	c := *f.draft
	return &c, nil
}
func (f *fakeRevisions) GetLatestPublished(ctx context.Context) (*domain.Revision, error) {
	if len(f.published) == 0 {
		return nil, apperrors.ErrNotFound
	}
	c := *f.published[len(f.published)-1]
	return &c, nil
}
func (f *fakeRevisions) GetRevision(ctx context.Context, id uuid.UUID) (*domain.Revision, error) {
	for _, r := range append([]*domain.Revision{f.draft}, f.published...) {
		if r != nil && r.ID == id {
			c := *r
			return &c, nil
		}
	}
	return nil, apperrors.ErrNotFound
}
func (f *fakeRevisions) ListPublished(ctx context.Context) ([]domain.Revision, error) {
	out := []domain.Revision{}
	for i := len(f.published) - 1; i >= 0; i-- {
		r := *f.published[i]
		r.Questions = nil
		out = append(out, r)
	}
	return out, nil
}
func (f *fakeRevisions) SaveDraft(ctx context.Context, d domain.Revision, expected *time.Time) (*domain.Revision, error) {
	if expected != nil && f.draft != nil && !f.draft.UpdatedAt.Equal(*expected) {
		return nil, apperrors.ErrConflict
	}
	f.draft = &d
	c := d
	return &c, nil
}
func (f *fakeRevisions) DiscardDraft(ctx context.Context, id uuid.UUID) error {
	f.draft = nil
	return nil
}
func (f *fakeRevisions) Publish(ctx context.Context, id uuid.UUID, now time.Time) (*domain.Revision, error) {
	r := *f.draft
	n := len(f.published) + 1
	r.Number, r.Status, r.PublishedAt = &n, domain.RevisionPublished, &now
	f.published = append(f.published, &r)
	f.draft = nil
	return &r, nil
}

type fakeAnswers struct {
	byUser map[uuid.UUID]map[string]domain.Answer
	locked bool
	writes int
}

func (f *fakeAnswers) ListByUser(ctx context.Context, u uuid.UUID) ([]domain.Answer, error) {
	out := []domain.Answer{}
	for _, a := range f.byUser[u] {
		out = append(out, a)
	}
	return out, nil
}
func (f *fakeAnswers) ListByUsers(ctx context.Context, us []uuid.UUID) ([]domain.Answer, error) {
	return nil, nil
}
func (f *fakeAnswers) Upsert(ctx context.Context, as []domain.Answer) error {
	f.writes++
	if f.byUser == nil {
		f.byUser = map[uuid.UUID]map[string]domain.Answer{}
	}
	for _, a := range as {
		if f.byUser[a.UserID] == nil {
			f.byUser[a.UserID] = map[string]domain.Answer{}
		}
		f.byUser[a.UserID][a.QuestionKey] = a
	}
	return nil
}
func (f *fakeAnswers) IsLockedForReview(ctx context.Context, u uuid.UUID) (bool, error) {
	return f.locked, nil
}

type fakeNudges struct {
	missing   []uuid.UUID
	created   map[uuid.UUID]bool
	dismissed map[uuid.UUID]bool
}

func (f *fakeNudges) ListUsersMissingAnswers(ctx context.Context, r uuid.UUID) ([]uuid.UUID, error) {
	return f.missing, nil
}
func (f *fakeNudges) Create(ctx context.Context, r uuid.UUID, by *uuid.UUID, us []uuid.UUID, now time.Time) ([]uuid.UUID, error) {
	if f.created == nil {
		f.created = map[uuid.UUID]bool{}
	}
	var fresh []uuid.UUID
	for _, u := range us {
		if !f.created[u] {
			f.created[u] = true
			fresh = append(fresh, u)
		}
	}
	return fresh, nil
}
func (f *fakeNudges) GetActive(ctx context.Context, u uuid.UUID) (*domain.Nudge, error) {
	if f.created[u] && !f.dismissed[u] {
		return &domain.Nudge{UserID: u, RevisionID: uuid.New()}, nil
	}
	return nil, apperrors.ErrNotFound
}
func (f *fakeNudges) Dismiss(ctx context.Context, r, u uuid.UUID, now time.Time) error {
	if f.dismissed == nil {
		f.dismissed = map[uuid.UUID]bool{}
	}
	f.dismissed[u] = true
	return nil
}

type mailRecorder struct{ to []string }

func (m *mailRecorder) Send(ctx context.Context, to, subject, body string) error {
	m.to = append(m.to, to)
	return nil
}
func (m *mailRecorder) SendHTML(ctx context.Context, to, subject, h, t string) error {
	return m.Send(ctx, to, subject, t)
}

type env struct {
	svc     *Service
	revs    *fakeRevisions
	answers *fakeAnswers
	nudges  *fakeNudges
	mail    *mailRecorder
	admin   uuid.UUID
}

func newEnv(t *testing.T) *env {
	t.Helper()
	e := &env{revs: &fakeRevisions{}, answers: &fakeAnswers{}, nudges: &fakeNudges{}, mail: &mailRecorder{}, admin: uuid.New()}
	e.svc = NewService(e.revs, e.answers, e.nudges, richtext.NewSanitizer(), &userRepoStub{}, WithMailer(e.mail, "https://artiv.test"))
	return e
}

func q(key string, typ domain.QuestionType) inbound.QuestionInput {
	return inbound.QuestionInput{Key: key, Type: typ, PromptEN: "Prompt " + key, IsPublic: true}
}

// publish saves and publishes a revision made of the given questions.
func (e *env) publish(t *testing.T, kind domain.ChangeKind, qs ...inbound.QuestionInput) *domain.Revision {
	t.Helper()
	ctx := context.Background()
	_, err := e.svc.SaveDraft(ctx, e.admin, inbound.DraftInput{ChangeKind: kind, Questions: qs})
	assist.NoError(t, err)
	rev, err := e.svc.Publish(ctx)
	assist.NoError(t, err)
	return rev
}

func TestAdmin_SaveDraftAndPublish(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()

	bio := q("bio", domain.TypeLongText)
	bio.Required = true
	rev := e.publish(t, domain.ChangeSubstantive, bio, q("city", domain.TypeShortText))
	assist.Equal(t, 1, *rev.Number)
	assist.Equal(t, 5000, rev.Questions[0].MaxLength) // default by type
	assist.Equal(t, 1, rev.Questions[0].Position)
	assist.Equal(t, 2, rev.Questions[1].Position)

	// A new draft starts as a copy of the latest revision, so admins edit rather than retype.
	draft, err := e.svc.StartDraft(ctx, e.admin)
	assist.NoError(t, err)
	assist.Len(t, 2, len(draft.Questions))
	again, err := e.svc.StartDraft(ctx, e.admin)
	assist.NoError(t, err)
	assist.Equal(t, draft.ID, again.ID) // idempotent while a draft exists
}

func TestAdmin_SaveDraftValidation(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()

	for name, in := range map[string]inbound.DraftInput{
		"bad change kind": {ChangeKind: "huge", Questions: []inbound.QuestionInput{q("a", domain.TypeShortText)}},
		"duplicate key":   {ChangeKind: domain.ChangeMinor, Questions: []inbound.QuestionInput{q("a", domain.TypeShortText), q("a", domain.TypeShortText)}},
		"bad key":         {ChangeKind: domain.ChangeMinor, Questions: []inbound.QuestionInput{q("Bad Key", domain.TypeShortText)}},
		"choice no opts":  {ChangeKind: domain.ChangeMinor, Questions: []inbound.QuestionInput{q("c", domain.TypeChoice)}},
		"no questions":    {ChangeKind: domain.ChangeMinor},
	} {
		_, err := e.svc.SaveDraft(ctx, e.admin, in)
		if err == nil {
			t.Errorf("%s: expected an error", name)
			continue
		}
		assist.ErrorIs(t, err, apperrors.ErrValidation)
	}
}

func TestAdmin_keyCannotChangeType(t *testing.T) {
	e := newEnv(t)
	e.publish(t, domain.ChangeSubstantive, q("about", domain.TypeShortText))

	_, err := e.svc.SaveDraft(context.Background(), e.admin, inbound.DraftInput{
		ChangeKind: domain.ChangeSubstantive, Questions: []inbound.QuestionInput{q("about", domain.TypeLongText)},
	})
	assist.ErrorIs(t, err, apperrors.ErrValidation)
	assist.Contains(t, err.Error(), "new key")
}

func TestAnswers_saveValidateAndSanitize(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	user := uuid.New()
	short := q("city", domain.TypeShortText)
	short.MaxLength = 5
	long := q("bio", domain.TypeLongText)
	long.MaxLength = 20
	long.Required = true
	e.publish(t, domain.ChangeSubstantive, long, short)

	// Autosave of a partial form is fine: required questions aren't enforced here.
	assist.NoError(t, e.svc.SaveAnswers(ctx, user, []inbound.AnswerInput{{Key: "city", Body: "Addis"}}))

	// Rich text is sanitised before storage.
	assist.NoError(t, e.svc.SaveAnswers(ctx, user, []inbound.AnswerInput{
		{Key: "bio", Body: `<p onclick="x()">ሰላም</p><script>alert(1)</script>`, Format: domain.FormatRich, Lang: domain.LangAmharic},
	}))
	stored := e.answers.byUser[user]["bio"]
	assist.Equal(t, "<p>ሰላም</p>", stored.Body)
	assist.Equal(t, domain.LangAmharic, stored.Lang)

	// Short answers can be formatted too; their limit (5 here) also counts visible text only.
	assist.NoError(t, e.svc.SaveAnswers(ctx, user, []inbound.AnswerInput{
		{Key: "city", Body: `<p><strong>Addis</strong><img src=x onerror=alert(1)></p>`, Format: domain.FormatRich},
	}))
	assist.Equal(t, "<p><strong>Addis</strong></p>", e.answers.byUser[user]["city"].Body)

	// The length limit counts visible text, not markup.
	assist.NoError(t, e.svc.SaveAnswers(ctx, user, []inbound.AnswerInput{
		{Key: "bio", Body: "<p><strong>" + strings.Repeat("a", 20) + "</strong></p>", Format: domain.FormatRich},
	}))
	err := e.svc.SaveAnswers(ctx, user, []inbound.AnswerInput{
		{Key: "bio", Body: "<p>" + strings.Repeat("a", 21) + "</p>", Format: domain.FormatRich},
	})
	assist.ErrorIs(t, err, apperrors.ErrValidation)

	for name, in := range map[string][]inbound.AnswerInput{
		"unknown key":   {{Key: "nope", Body: "x"}},
		"duplicate key": {{Key: "city", Body: "a"}, {Key: "city", Body: "b"}},
		"too long":      {{Key: "city", Body: "Addis Ababa"}},
		"bad language":  {{Key: "city", Body: "x", Lang: "fr"}},
	} {
		err := e.svc.SaveAnswers(ctx, user, in)
		if err == nil {
			t.Errorf("%s: expected an error", name)
			continue
		}
		assist.ErrorIs(t, err, apperrors.ErrValidation)
	}
}

func TestAnswers_invalidBatchWritesNothing(t *testing.T) {
	e := newEnv(t)
	e.publish(t, domain.ChangeSubstantive, q("a", domain.TypeShortText))
	before := e.answers.writes
	err := e.svc.SaveAnswers(context.Background(), uuid.New(), []inbound.AnswerInput{{Key: "a", Body: "ok"}, {Key: "zzz", Body: "bad"}})
	assist.ErrorIs(t, err, apperrors.ErrValidation)
	assist.Equal(t, before, e.answers.writes)
}

func TestAnswers_lockedWhileUnderReview(t *testing.T) {
	e := newEnv(t)
	e.publish(t, domain.ChangeSubstantive, q("a", domain.TypeShortText))
	e.answers.locked = true

	err := e.svc.SaveAnswers(context.Background(), uuid.New(), []inbound.AnswerInput{{Key: "a", Body: "x"}})
	assist.ErrorIs(t, err, apperrors.ErrConflict)
	form, err := e.svc.GetForm(context.Background(), uuid.New())
	assist.NoError(t, err)
	assist.Equal(t, true, form.Locked)
}

func TestGetForm_beforeFirstPublish(t *testing.T) {
	_, err := newEnv(t).svc.GetForm(context.Background(), uuid.New())
	assist.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestCheckComplete(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	user := uuid.New()

	submitted, err := e.svc.CheckComplete(ctx, user)
	assist.NoError(t, err)
	if submitted != nil {
		t.Error("no published questionnaire should mean no requirement")
	}

	bio := q("bio", domain.TypeLongText)
	bio.Required = true
	published := e.publish(t, domain.ChangeSubstantive, bio, q("city", domain.TypeShortText))

	_, err = e.svc.CheckComplete(ctx, user)
	assist.ErrorIs(t, err, apperrors.ErrValidation)
	assist.Contains(t, err.Error(), "bio")

	assist.NoError(t, e.svc.SaveAnswers(ctx, user, []inbound.AnswerInput{{Key: "bio", Body: "hello"}}))
	submitted, err = e.svc.CheckComplete(ctx, user)
	assist.NoError(t, err)
	assist.Equal(t, published.ID, submitted.RevisionID) // optional "city" need not be answered
}

func TestPublicAnswers_onlyPublicNonBlankCurrentQuestions(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	user := uuid.New()
	private := q("income", domain.TypeShortText)
	private.IsPublic = false
	e.publish(t, domain.ChangeSubstantive, q("bio", domain.TypeShortText), private, q("city", domain.TypeShortText), q("old", domain.TypeShortText))

	assist.NoError(t, e.svc.SaveAnswers(ctx, user, []inbound.AnswerInput{
		{Key: "bio", Body: "hi"}, {Key: "income", Body: "secret"}, {Key: "city", Body: "  "}, {Key: "old", Body: "legacy"},
	}))
	// "old" is removed in revision 2.
	e.publish(t, domain.ChangeSubstantive, q("bio", domain.TypeShortText), private, q("city", domain.TypeShortText))

	got, err := e.svc.PublicAnswers(ctx, user)
	assist.NoError(t, err)
	assist.Len(t, 1, len(got))
	assist.Equal(t, "bio", got[0].Question.Key)

	// Marking a question private later hides it without deleting the answer.
	e.publish(t, domain.ChangeMinor, func() inbound.QuestionInput { x := q("bio", domain.TypeShortText); x.IsPublic = false; return x }(), private, q("city", domain.TypeShortText))
	got, _ = e.svc.PublicAnswers(ctx, user)
	assist.Len(t, 0, len(got))
	assist.Equal(t, "hi", e.answers.byUser[user]["bio"].Body)
}

func TestReviewView_includesPrivateAndOtherAnswers(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	user := uuid.New()
	private := q("income", domain.TypeShortText)
	private.IsPublic = false
	r1 := e.publish(t, domain.ChangeSubstantive, q("bio", domain.TypeShortText), private)
	assist.NoError(t, e.svc.SaveAnswers(ctx, user, []inbound.AnswerInput{{Key: "bio", Body: "hi"}, {Key: "income", Body: "secret"}}))
	e.publish(t, domain.ChangeSubstantive, q("bio", domain.TypeShortText), q("new_q", domain.TypeShortText))

	// Reviewed against the revision the applicant answered, not the newest one.
	view, err := e.svc.ReviewView(ctx, user, &r1.ID)
	assist.NoError(t, err)
	assist.Len(t, 2, len(view.Items))
	assist.Equal(t, "secret", view.Items[1].Answer.Body)
	assist.Len(t, 0, len(view.Other))

	// Against the latest revision, "income" is no longer in the form, so it moves to Other.
	latest, err := e.svc.ReviewView(ctx, user, nil)
	assist.NoError(t, err)
	assist.Len(t, 1, len(latest.Other))
	assist.Equal(t, "income", latest.Other[0].QuestionKey)
	if latest.Items[1].Answer != nil {
		t.Error("unanswered new question should have a nil answer")
	}
}

func TestNudge_minorRevisionNeverSends(t *testing.T) {
	e := newEnv(t)
	e.nudges.missing = []uuid.UUID{uuid.New()}
	e.publish(t, domain.ChangeMinor, q("a", domain.TypeShortText))

	preview, err := e.svc.PreviewNudge(context.Background())
	assist.NoError(t, err)
	assist.Equal(t, false, preview.CanSend)
	_, err = e.svc.SendNudge(context.Background(), e.admin)
	assist.ErrorIs(t, err, apperrors.ErrValidation)
	assist.Len(t, 0, len(e.mail.to))
}

func TestNudge_sendsOncePerArtist(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	a, b := uuid.New(), uuid.New()
	e.nudges.missing = []uuid.UUID{a, b}
	e.publish(t, domain.ChangeSubstantive, q("a", domain.TypeShortText))

	preview, _ := e.svc.PreviewNudge(ctx)
	assist.Equal(t, true, preview.CanSend)
	assist.Equal(t, 2, preview.Eligible)

	n, err := e.svc.SendNudge(ctx, e.admin)
	assist.NoError(t, err)
	assist.Equal(t, 2, n)
	assist.Len(t, 2, len(e.mail.to))

	n, err = e.svc.SendNudge(ctx, e.admin)
	assist.NoError(t, err)
	assist.Equal(t, 0, n) // already nudged: no second email
	assist.Len(t, 2, len(e.mail.to))
}

func TestActiveNudge_listsUnansweredAndAutoDismisses(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	artist := uuid.New()
	e.nudges.missing = []uuid.UUID{artist}
	e.publish(t, domain.ChangeSubstantive, q("a", domain.TypeShortText), q("b", domain.TypeShortText))
	_, err := e.svc.SendNudge(ctx, e.admin)
	assist.NoError(t, err)

	assist.NoError(t, e.svc.SaveAnswers(ctx, artist, []inbound.AnswerInput{{Key: "a", Body: "x"}}))
	prompt, err := e.svc.ActiveNudge(ctx, artist)
	assist.NoError(t, err)
	assist.Len(t, 1, len(prompt.Questions))
	assist.Equal(t, "b", prompt.Questions[0].Key)

	// Answering the rest on their own clears the banner.
	assist.NoError(t, e.svc.SaveAnswers(ctx, artist, []inbound.AnswerInput{{Key: "b", Body: "y"}}))
	_, err = e.svc.ActiveNudge(ctx, artist)
	assist.ErrorIs(t, err, apperrors.ErrNotFound)
	assist.Equal(t, true, e.nudges.dismissed[artist])
}

func TestOutput_resanitisesStoredRichAnswers(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	user := uuid.New()
	other := q("note", domain.TypeLongText)
	e.publish(t, domain.ChangeSubstantive, q("bio", domain.TypeLongText), other)

	// Simulate a row that bypassed SaveAnswers (manual edit, script, older policy).
	dirty := `<p>hi</p><script>alert(1)</script><a href="javascript:alert(2)" onclick="x()">link</a>`
	rev, _ := e.revs.GetLatestPublished(ctx)
	e.answers.byUser = map[uuid.UUID]map[string]domain.Answer{user: {
		"bio":  {UserID: user, QuestionKey: "bio", RevisionID: rev.ID, Body: dirty, Format: domain.FormatRich, Lang: domain.LangEnglish},
		"note": {UserID: user, QuestionKey: "note", RevisionID: rev.ID, Body: dirty, Format: domain.FormatRich, Lang: domain.LangEnglish},
	}}

	assertClean := func(where, body string) {
		t.Helper()
		for _, bad := range []string{"<script", "alert(", "javascript:", "onclick"} {
			if strings.Contains(body, bad) {
				t.Errorf("%s leaked %q: %s", where, bad, body)
			}
		}
		assist.Contains(t, body, "<p>hi</p>")
	}

	public, err := e.svc.PublicAnswers(ctx, user)
	assist.NoError(t, err)
	assist.Len(t, 2, len(public))
	for _, p := range public {
		assertClean("PublicAnswers", p.Answer.Body)
	}
	form, err := e.svc.GetForm(ctx, user)
	assist.NoError(t, err)
	for _, a := range form.Answers {
		assertClean("GetForm", a.Body)
	}
	view, err := e.svc.ReviewView(ctx, user, nil)
	assist.NoError(t, err)
	for _, it := range view.Items {
		assertClean("ReviewView", it.Answer.Body)
	}
}

func TestLinkAnswers_mustBeWebURLs(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	user := uuid.New()
	e.publish(t, domain.ChangeSubstantive, q("website", domain.TypeLink), q("bio", domain.TypeShortText))

	for _, bad := range []string{
		"javascript:alert(1)", "JaVaScRiPt:alert(1)", "data:text/html,<script>alert(1)</script>",
		"vbscript:x", "file:///etc/passwd", "//evil.example/x", "example.com", "https://user:pw@example.com",
		"https:///nohost", "  javascript:alert(1)",
	} {
		err := e.svc.SaveAnswers(ctx, user, []inbound.AnswerInput{{Key: "website", Body: bad}})
		if err == nil {
			t.Errorf("%q should be rejected", bad)
			continue
		}
		assist.ErrorIs(t, err, apperrors.ErrValidation)
	}
	for _, good := range []string{"https://example.com", "http://example.com/a?b=c#d", ""} {
		assist.NoError(t, e.svc.SaveAnswers(ctx, user, []inbound.AnswerInput{{Key: "website", Body: good}}))
	}
}

func TestPublicAnswers_dropsUnsafeLinkRows(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	user := uuid.New()
	rev := e.publish(t, domain.ChangeSubstantive, q("website", domain.TypeLink), q("bio", domain.TypeShortText))

	// A row that bypassed SaveAnswers (script, manual edit, older code).
	e.answers.byUser = map[uuid.UUID]map[string]domain.Answer{user: {
		"website": {UserID: user, QuestionKey: "website", RevisionID: rev.ID, Body: "javascript:alert(1)", Format: domain.FormatPlain, Lang: domain.LangEnglish},
		"bio":     {UserID: user, QuestionKey: "bio", RevisionID: rev.ID, Body: "hello", Format: domain.FormatPlain, Lang: domain.LangEnglish},
	}}
	got, err := e.svc.PublicAnswers(ctx, user)
	assist.NoError(t, err)
	assist.Len(t, 1, len(got))
	assist.Equal(t, "bio", got[0].Question.Key)

	e.answers.byUser[user]["website"] = domain.Answer{UserID: user, QuestionKey: "website", RevisionID: rev.ID, Body: "https://example.com", Format: domain.FormatPlain, Lang: domain.LangEnglish}
	got, _ = e.svc.PublicAnswers(ctx, user)
	assist.Len(t, 2, len(got))
}

type artistsStub struct {
	bySlug map[string]*profiledomain.ArtistProfile
}

func (a artistsStub) GetArtistBySlug(ctx context.Context, slug string) (*profiledomain.ArtistProfile, error) {
	if p, ok := a.bySlug[slug]; ok {
		return p, nil
	}
	return nil, apperrors.ErrNotFound
}

type applicationsStub struct {
	byID map[uuid.UUID]*onboardingdomain.OnboardingApplication
}

func (a applicationsStub) GetByID(ctx context.Context, id uuid.UUID) (*onboardingdomain.OnboardingApplication, error) {
	if app, ok := a.byID[id]; ok {
		return app, nil
	}
	return nil, apperrors.ErrNotFound
}

func TestPublicAnswersBySlug_onlyForPublicProfiles(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	approved, draft := uuid.New(), uuid.New()
	e.svc = NewService(e.revs, e.answers, e.nudges, richtext.NewSanitizer(), &userRepoStub{},
		WithLookups(artistsStub{bySlug: map[string]*profiledomain.ArtistProfile{
			"ada":   {UserID: approved, Status: profiledomain.ProfileStatusApproved},
			"draft": {UserID: draft, Status: profiledomain.ProfileStatusDraft},
		}}, nil))
	e.publish(t, domain.ChangeSubstantive, q("bio", domain.TypeShortText))
	for _, user := range []uuid.UUID{approved, draft} {
		assist.NoError(t, e.svc.SaveAnswers(ctx, user, []inbound.AnswerInput{{Key: "bio", Body: "hello"}}))
	}

	got, err := e.svc.PublicAnswersBySlug(ctx, "ada")
	assist.NoError(t, err)
	assist.Len(t, 1, len(got))

	// A draft profile has answers too, but they look exactly like an unknown slug.
	_, err = e.svc.PublicAnswersBySlug(ctx, "draft")
	assist.ErrorIs(t, err, apperrors.ErrNotFound)
	_, err = e.svc.PublicAnswersBySlug(ctx, "nobody")
	assist.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestApplicationReviewView_resolvesTheApplicant(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	user, appID := uuid.New(), uuid.New()
	rev := e.publish(t, domain.ChangeSubstantive, q("bio", domain.TypeShortText))
	assist.NoError(t, e.svc.SaveAnswers(ctx, user, []inbound.AnswerInput{{Key: "bio", Body: "hello"}}))
	e.svc = NewService(e.revs, e.answers, e.nudges, richtext.NewSanitizer(), &userRepoStub{},
		WithLookups(nil, applicationsStub{byID: map[uuid.UUID]*onboardingdomain.OnboardingApplication{
			appID: {ID: appID, ApplicantID: user, QuestionnaireRevisionID: &rev.ID},
		}}))

	view, err := e.svc.ApplicationReviewView(ctx, appID)
	assist.NoError(t, err)
	assist.Equal(t, rev.ID, view.Revision.ID)
	assist.Len(t, 1, len(view.Items))

	_, err = e.svc.ApplicationReviewView(ctx, uuid.New())
	assist.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestApplicationReviewView_showsWhatWasSubmittedNotWhatTheArtistWroteSince(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	user, appID := uuid.New(), uuid.New()
	e.publish(t, domain.ChangeSubstantive, q("bio", domain.TypeShortText))
	assist.NoError(t, e.svc.SaveAnswers(ctx, user, []inbound.AnswerInput{{Key: "bio", Body: "as submitted"}}))

	submitted, err := e.svc.CheckComplete(ctx, user)
	assist.NoError(t, err)
	app := &onboardingdomain.OnboardingApplication{
		ID: appID, ApplicantID: user, QuestionnaireRevisionID: &submitted.RevisionID, QuestionnaireAnswers: submitted.Answers,
	}
	e.svc = NewService(e.revs, e.answers, e.nudges, richtext.NewSanitizer(), &userRepoStub{},
		WithLookups(nil, applicationsStub{byID: map[uuid.UUID]*onboardingdomain.OnboardingApplication{appID: app}}))

	// The artist edits the live answer after the application was decided.
	assist.NoError(t, e.svc.SaveAnswers(ctx, user, []inbound.AnswerInput{{Key: "bio", Body: "edited later"}}))

	view, err := e.svc.ApplicationReviewView(ctx, appID)
	assist.NoError(t, err)
	assist.Equal(t, "as submitted", view.Items[0].Answer.Body)
	assist.Equal(t, user, view.Items[0].Answer.UserID)
}

func TestApplicationReviewView_snapshotIsSanitisedOnTheWayOut(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	user, appID := uuid.New(), uuid.New()
	rev := e.publish(t, domain.ChangeSubstantive, q("bio", domain.TypeLongText))
	raw, err := encodeSnapshot([]domain.Answer{{
		QuestionKey: "bio", RevisionID: rev.ID, Format: domain.FormatRich, Lang: domain.LangEnglish,
		Body: "<p>hi</p><script>alert(1)</script>",
	}})
	assist.NoError(t, err)
	e.svc = NewService(e.revs, e.answers, e.nudges, richtext.NewSanitizer(), &userRepoStub{},
		WithLookups(nil, applicationsStub{byID: map[uuid.UUID]*onboardingdomain.OnboardingApplication{
			appID: {ID: appID, ApplicantID: user, QuestionnaireRevisionID: &rev.ID, QuestionnaireAnswers: raw},
		}}))

	view, err := e.svc.ApplicationReviewView(ctx, appID)
	assist.NoError(t, err)
	if strings.Contains(view.Items[0].Answer.Body, "script") {
		t.Errorf("snapshot html must be re-sanitised, got %q", view.Items[0].Answer.Body)
	}
}

func TestAdmin_SaveDraft_refusesAStaleEditor(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	e.publish(t, domain.ChangeSubstantive, q("bio", domain.TypeShortText))

	loaded, err := e.svc.StartDraft(ctx, e.admin)
	assist.NoError(t, err)
	seen := loaded.UpdatedAt

	// Another admin saves first (the clock has moved on, so updated_at differs).
	time.Sleep(2 * time.Millisecond)
	_, err = e.svc.SaveDraft(ctx, e.admin, inbound.DraftInput{ChangeKind: domain.ChangeMinor, Questions: []inbound.QuestionInput{q("bio", domain.TypeShortText), q("city", domain.TypeShortText)}})
	assist.NoError(t, err)

	// Ours was made against the older copy.
	_, err = e.svc.SaveDraft(ctx, e.admin, inbound.DraftInput{
		ChangeKind: domain.ChangeMinor, ExpectedUpdatedAt: &seen, Questions: []inbound.QuestionInput{q("bio", domain.TypeShortText)},
	})
	assist.ErrorIs(t, err, apperrors.ErrConflict)
	assist.Contains(t, err.Error(), "reload")
	assist.Len(t, 2, len(e.revs.draft.Questions)) // their save survived

	// Using the current timestamp works.
	current := e.revs.draft.UpdatedAt
	_, err = e.svc.SaveDraft(ctx, e.admin, inbound.DraftInput{
		ChangeKind: domain.ChangeMinor, ExpectedUpdatedAt: &current, Questions: []inbound.QuestionInput{q("bio", domain.TypeShortText)},
	})
	assist.NoError(t, err)
}
