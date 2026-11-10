package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/domain/questionnaire"
	"github.com/mq/api/internal/port/inbound"
)

// ---- questions and revisions ----

type OptionDTO struct {
	Value   string `json:"value" binding:"required,max=100"`
	LabelEN string `json:"label_en" binding:"required,max=200"`
	LabelAM string `json:"label_am,omitempty" binding:"max=200"`
}

type QuestionRequest struct {
	Key       string      `json:"key" binding:"required,max=64"`
	Type      string      `json:"type" binding:"required"`
	PromptEN  string      `json:"prompt_en" binding:"required,max=500"`
	PromptAM  string      `json:"prompt_am" binding:"max=500"`
	HelpEN    string      `json:"help_en" binding:"max=1000"`
	HelpAM    string      `json:"help_am" binding:"max=1000"`
	Required  bool        `json:"required"`
	IsPublic  bool        `json:"is_public"`
	MaxLength int         `json:"max_length" binding:"min=0,max=20000"`
	Options   []OptionDTO `json:"options" binding:"max=50,dive"`
}

type SaveDraftRequest struct {
	ChangeKind string            `json:"change_kind" binding:"required"`
	Note       string            `json:"note" binding:"max=1000"`
	Questions  []QuestionRequest `json:"questions" binding:"required,min=1,max=100,dive"`
	// ExpectedUpdatedAt is the draft's updated_at when the editor loaded it; a stale save gets 409.
	ExpectedUpdatedAt *time.Time `json:"expected_updated_at,omitempty"`
}

type QuestionResponse struct {
	Key       string      `json:"key"`
	Position  int         `json:"position"`
	Type      string      `json:"type"`
	PromptEN  string      `json:"prompt_en"`
	PromptAM  string      `json:"prompt_am,omitempty"`
	HelpEN    string      `json:"help_en,omitempty"`
	HelpAM    string      `json:"help_am,omitempty"`
	Required  bool        `json:"required"`
	IsPublic  bool        `json:"is_public"`
	MaxLength int         `json:"max_length"`
	Options   []OptionDTO `json:"options,omitempty"`
}

type RevisionResponse struct {
	ID          uuid.UUID          `json:"id"`
	Number      *int               `json:"number,omitempty"`
	Status      string             `json:"status"`
	ChangeKind  string             `json:"change_kind"`
	Note        string             `json:"note,omitempty"`
	PublishedAt *time.Time         `json:"published_at,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	Questions   []QuestionResponse `json:"questions"`
}

func ToQuestionResponse(q questionnaire.Question) QuestionResponse {
	r := QuestionResponse{
		Key: q.Key, Position: q.Position, Type: string(q.Type), PromptEN: q.PromptEN, PromptAM: q.PromptAM,
		HelpEN: q.HelpEN, HelpAM: q.HelpAM, Required: q.Required, IsPublic: q.IsPublic, MaxLength: q.MaxLength,
	}
	for _, o := range q.Options {
		r.Options = append(r.Options, OptionDTO{Value: o.Value, LabelEN: o.LabelEN, LabelAM: o.LabelAM})
	}
	return r
}

func ToRevisionResponse(r questionnaire.Revision) RevisionResponse {
	out := RevisionResponse{
		ID: r.ID, Number: r.Number, Status: string(r.Status), ChangeKind: string(r.ChangeKind), Note: r.Note,
		PublishedAt: r.PublishedAt, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		Questions: make([]QuestionResponse, 0, len(r.Questions)),
	}
	for _, q := range r.Questions {
		out.Questions = append(out.Questions, ToQuestionResponse(q))
	}
	return out
}

func ToRevisionResponses(rs []questionnaire.Revision) []RevisionResponse {
	out := make([]RevisionResponse, len(rs))
	for i, r := range rs {
		out[i] = ToRevisionResponse(r)
	}
	return out
}

func ToDraftInput(req SaveDraftRequest) inbound.DraftInput {
	in := inbound.DraftInput{ChangeKind: questionnaire.ChangeKind(req.ChangeKind), Note: req.Note, ExpectedUpdatedAt: req.ExpectedUpdatedAt}
	for _, q := range req.Questions {
		qi := inbound.QuestionInput{
			Key: q.Key, Type: questionnaire.QuestionType(q.Type), PromptEN: q.PromptEN, PromptAM: q.PromptAM,
			HelpEN: q.HelpEN, HelpAM: q.HelpAM, Required: q.Required, IsPublic: q.IsPublic, MaxLength: q.MaxLength,
		}
		for _, o := range q.Options {
			qi.Options = append(qi.Options, questionnaire.Option{Value: o.Value, LabelEN: o.LabelEN, LabelAM: o.LabelAM})
		}
		in.Questions = append(in.Questions, qi)
	}
	return in
}

// ---- answers ----

type AnswerRequest struct {
	Key    string `json:"key" binding:"required,max=64"`
	Body   string `json:"body" binding:"max=100000"`
	Format string `json:"format" binding:"omitempty,oneof=plain rich"`
	Lang   string `json:"lang" binding:"omitempty,oneof=en am mixed"`
}

type SaveAnswersRequest struct {
	Answers []AnswerRequest `json:"answers" binding:"required,min=1,max=100,dive"`
}

type AnswerResponse struct {
	Key       string    `json:"key"`
	Body      string    `json:"body"`
	Format    string    `json:"format"`
	Lang      string    `json:"lang"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToAnswerResponse(a questionnaire.Answer) AnswerResponse {
	return AnswerResponse{Key: a.QuestionKey, Body: a.Body, Format: string(a.Format), Lang: string(a.Lang), UpdatedAt: a.UpdatedAt}
}

func ToAnswerResponses(as []questionnaire.Answer) []AnswerResponse {
	out := make([]AnswerResponse, len(as))
	for i, a := range as {
		out[i] = ToAnswerResponse(a)
	}
	return out
}

func ToSaveAnswersInput(req SaveAnswersRequest) []inbound.AnswerInput {
	out := make([]inbound.AnswerInput, len(req.Answers))
	for i, a := range req.Answers {
		out[i] = inbound.AnswerInput{
			Key: a.Key, Body: a.Body, Format: questionnaire.Format(a.Format), Lang: questionnaire.Lang(a.Lang),
		}
	}
	return out
}

type FormResponse struct {
	Revision RevisionResponse `json:"revision"`
	Answers  []AnswerResponse `json:"answers"`
	Locked   bool             `json:"locked"`
}

func ToFormResponse(f inbound.Form) FormResponse {
	return FormResponse{Revision: ToRevisionResponse(f.Revision), Answers: ToAnswerResponses(f.Answers), Locked: f.Locked}
}

// PublicAnswerResponse is one visible answer on an artist's public page. For format "rich" the
// body is HTML that has been sanitised server-side and is safe to render; for "plain" it must be
// rendered as text. Choice answers carry the option value in body and the options for lookup.
type PublicAnswerResponse struct {
	Key      string      `json:"key"`
	Type     string      `json:"type"`
	PromptEN string      `json:"prompt_en"`
	PromptAM string      `json:"prompt_am,omitempty"`
	Options  []OptionDTO `json:"options,omitempty"`
	Body     string      `json:"body"`
	Format   string      `json:"format"`
	Lang     string      `json:"lang"`
}

func ToPublicAnswerResponses(as []inbound.PublicAnswer) []PublicAnswerResponse {
	out := make([]PublicAnswerResponse, len(as))
	for i, a := range as {
		q := ToQuestionResponse(a.Question)
		out[i] = PublicAnswerResponse{
			Key: q.Key, Type: q.Type, PromptEN: q.PromptEN, PromptAM: q.PromptAM, Options: q.Options,
			Body: a.Answer.Body, Format: string(a.Answer.Format), Lang: string(a.Answer.Lang),
		}
	}
	return out
}

type ReviewItemResponse struct {
	Question QuestionResponse `json:"question"`
	Answer   *AnswerResponse  `json:"answer"`
}

type ReviewViewResponse struct {
	Revision RevisionResponse     `json:"revision"`
	Items    []ReviewItemResponse `json:"items"`
	Other    []AnswerResponse     `json:"other"`
}

func ToReviewViewResponse(v inbound.ReviewView) ReviewViewResponse {
	out := ReviewViewResponse{
		Revision: ToRevisionResponse(v.Revision), Items: make([]ReviewItemResponse, 0, len(v.Items)),
		Other: ToAnswerResponses(v.Other),
	}
	for _, it := range v.Items {
		item := ReviewItemResponse{Question: ToQuestionResponse(it.Question)}
		if it.Answer != nil {
			a := ToAnswerResponse(*it.Answer)
			item.Answer = &a
		}
		out.Items = append(out.Items, item)
	}
	return out
}

// ---- nudges ----

type NudgePromptResponse struct {
	RevisionID uuid.UUID          `json:"revision_id"`
	SentAt     time.Time          `json:"sent_at"`
	Questions  []QuestionResponse `json:"questions"`
}

func ToNudgePromptResponse(p inbound.NudgePrompt) NudgePromptResponse {
	out := NudgePromptResponse{RevisionID: p.Nudge.RevisionID, SentAt: p.Nudge.SentAt, Questions: make([]QuestionResponse, 0, len(p.Questions))}
	for _, q := range p.Questions {
		out.Questions = append(out.Questions, ToQuestionResponse(q))
	}
	return out
}

type NudgePreviewResponse struct {
	RevisionID     uuid.UUID `json:"revision_id"`
	RevisionNumber *int      `json:"revision_number,omitempty"`
	ChangeKind     string    `json:"change_kind"`
	Eligible       int       `json:"eligible"`
	CanSend        bool      `json:"can_send"`
}

func ToNudgePreviewResponse(p inbound.NudgePreview) NudgePreviewResponse {
	return NudgePreviewResponse{
		RevisionID: p.Revision.ID, RevisionNumber: p.Revision.Number, ChangeKind: string(p.Revision.ChangeKind),
		Eligible: p.Eligible, CanSend: p.CanSend,
	}
}

type SendNudgeResponse struct {
	Sent int `json:"sent"`
}

// ---- invites ----

type CreateInviteRequest struct {
	Email       string `json:"email" binding:"required,max=254"`
	DisplayName string `json:"display_name" binding:"max=100"`
	Handle      string `json:"handle" binding:"max=30"`
	TTLDays     int    `json:"ttl_days" binding:"min=0,max=60"`
}

// InviteResponse never includes the token or its hash.
type InviteResponse struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	DisplayName     string     `json:"display_name,omitempty"`
	RequestedHandle string     `json:"requested_handle,omitempty"`
	Status          string     `json:"status"`
	ExpiresAt       time.Time  `json:"expires_at"`
	ClaimedAt       *time.Time `json:"claimed_at,omitempty"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

func ToInviteResponse(i onboarding.Invite, now time.Time) InviteResponse {
	return InviteResponse{
		ID: i.ID, Email: i.Email, DisplayName: i.DisplayName, RequestedHandle: i.RequestedHandle,
		Status: string(i.Status(now)), ExpiresAt: i.ExpiresAt, ClaimedAt: i.ClaimedAt, RevokedAt: i.RevokedAt, CreatedAt: i.CreatedAt,
	}
}

func ToInviteResponses(is []onboarding.Invite, now time.Time) []InviteResponse {
	out := make([]InviteResponse, len(is))
	for i, inv := range is {
		out[i] = ToInviteResponse(inv, now)
	}
	return out
}

// InviteCreatedResponse is the only place the raw token and link are ever returned. They cannot be
// retrieved again, so the admin UI should present them once (the email is best-effort).
type InviteCreatedResponse struct {
	Invite InviteResponse `json:"invite"`
	Token  string         `json:"token"`
	Link   string         `json:"link"`
}

type InvitePreviewRequest struct {
	Token string `json:"token" binding:"required,max=200"`
}

type InvitePreviewResponse struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name,omitempty"`
	Handle      string `json:"handle,omitempty"`
	Status      string `json:"status"`
}

type ClaimInviteRequest struct {
	Token           string `json:"token" binding:"required,max=200"`
	DisplayName     string `json:"display_name" binding:"max=100"`
	RequestedHandle string `json:"requested_handle" binding:"max=30"`
	Notes           string `json:"notes" binding:"max=5000"`
}

// InviteRegisterRequest carries no email on purpose: the address comes from the invite.
type InviteRegisterRequest struct {
	Token    string `json:"token" binding:"required,max=200"`
	Password string `json:"password" binding:"required,max=200"`
}
