// Package questionnaire models the admin-editable artist questionnaire: immutable published
// revisions of a question set, and the answers artists give to them.
package questionnaire

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type RevisionStatus string

const (
	RevisionDraft     RevisionStatus = "draft"
	RevisionPublished RevisionStatus = "published"
)

// ChangeKind says whether a revision warrants nudging artists. Only substantive changes
// (a new or changed question) do; minor ones are wording polish.
type ChangeKind string

const (
	ChangeMinor       ChangeKind = "minor"
	ChangeSubstantive ChangeKind = "substantive"
)

type QuestionType string

const (
	TypeShortText QuestionType = "short_text"
	TypeLongText  QuestionType = "long_text"
	TypeChoice    QuestionType = "choice"
	TypeLink      QuestionType = "link"
)

type Format string

const (
	// FormatPlain is verbatim text.
	FormatPlain Format = "plain"
	// FormatRich is HTML already sanitised on the server against a fixed allowlist.
	// Only long_text questions accept it.
	FormatRich Format = "rich"
)

type Lang string

const (
	LangEnglish Lang = "en"
	LangAmharic Lang = "am"
	LangMixed   Lang = "mixed"
)

// Revision is a full snapshot of the questions. It is editable while a draft and immutable once
// published; Number is nil until then.
type Revision struct {
	ID          uuid.UUID
	Number      *int
	Status      RevisionStatus
	ChangeKind  ChangeKind
	Note        string
	CreatedBy   *uuid.UUID
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Questions   []Question
}

type Option struct {
	Value   string `json:"value"`
	LabelEN string `json:"label_en"`
	LabelAM string `json:"label_am,omitempty"`
}

// Question's Key is its identity across revisions: answers attach to it, so changing what is
// being asked must mint a new key rather than reuse the old one.
type Question struct {
	ID         uuid.UUID
	RevisionID uuid.UUID
	Key        string
	Position   int
	Type       QuestionType
	PromptEN   string
	PromptAM   string
	HelpEN     string
	HelpAM     string
	Required   bool
	// IsPublic false means admin-only; the answer never appears on the public profile.
	IsPublic  bool
	MaxLength int
	Options   []Option
}

// Answer is an artist's current answer to a question key. RevisionID is the revision it was
// last written against.
type Answer struct {
	UserID      uuid.UUID
	QuestionKey string
	RevisionID  uuid.UUID
	Body        string
	Format      Format
	Lang        Lang
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Nudge is an optional, admin-triggered invitation to answer what a revision added.
type Nudge struct {
	RevisionID  uuid.UUID
	UserID      uuid.UUID
	SentBy      *uuid.UUID
	SentAt      time.Time
	DismissedAt *time.Time
}

var keyPattern = regexp.MustCompile(`^[a-z0-9_]{1,64}$`)

// ErrInvalid is matched (errors.Is) by every validation failure in this package. The domain does
// not depend on apperrors; the use-case and adapter layers translate it to apperrors.ErrValidation.
var ErrInvalid = errors.New("invalid questionnaire input")

// ValidationError carries a message safe to show to the admin or artist who caused it.
type ValidationError struct{ Msg string }

func (e *ValidationError) Error() string        { return e.Msg }
func (e *ValidationError) Is(target error) bool { return target == ErrInvalid }

func invalid(format string, args ...any) error {
	return &ValidationError{Msg: fmt.Sprintf(format, args...)}
}

// DefaultMaxLength is applied when an admin leaves the limit blank.
func DefaultMaxLength(t QuestionType) int {
	switch t {
	case TypeLongText:
		return 5000
	case TypeLink:
		return 300
	case TypeChoice:
		return 100
	default:
		return 200
	}
}

func (t QuestionType) valid() bool {
	switch t {
	case TypeShortText, TypeLongText, TypeChoice, TypeLink:
		return true
	}
	return false
}

func (l Lang) valid() bool { return l == LangEnglish || l == LangAmharic || l == LangMixed }

// Validate checks a single question's shape.
func (q Question) Validate() error {
	if !keyPattern.MatchString(q.Key) {
		return invalid("question key %q must be 1-64 lowercase letters, numbers or underscores", q.Key)
	}
	if !q.Type.valid() {
		return invalid("question %q has invalid type %q", q.Key, q.Type)
	}
	if strings.TrimSpace(q.PromptEN) == "" {
		return invalid("question %q needs an English prompt", q.Key)
	}
	if q.MaxLength <= 0 {
		return invalid("question %q needs a positive max length", q.Key)
	}
	if q.Type == TypeChoice {
		if len(q.Options) < 2 {
			return invalid("question %q needs at least two options", q.Key)
		}
		seen := make(map[string]bool, len(q.Options))
		for _, o := range q.Options {
			if o.Value == "" || strings.TrimSpace(o.LabelEN) == "" {
				return invalid("question %q has an option without a value or English label", q.Key)
			}
			if seen[o.Value] {
				return invalid("question %q repeats option %q", q.Key, o.Value)
			}
			seen[o.Value] = true
		}
	} else if len(q.Options) > 0 {
		return invalid("question %q: options are only allowed on choice questions", q.Key)
	}
	return nil
}

// Validate checks every question and that keys and positions are unique within the revision.
// A revision must be valid before it can be published.
func (r Revision) Validate() error {
	if len(r.Questions) == 0 {
		return invalid("a revision needs at least one question")
	}
	keys := make(map[string]bool, len(r.Questions))
	positions := make(map[int]bool, len(r.Questions))
	for _, q := range r.Questions {
		if err := q.Validate(); err != nil {
			return err
		}
		if keys[q.Key] {
			return invalid("question key %q is used twice", q.Key)
		}
		if positions[q.Position] {
			return invalid("position %d is used twice", q.Position)
		}
		keys[q.Key], positions[q.Position] = true, true
	}
	return nil
}

// Question returns the question with the given key, if the revision has it.
func (r Revision) Question(key string) (Question, bool) {
	for _, q := range r.Questions {
		if q.Key == key {
			return q, true
		}
	}
	return Question{}, false
}

// CheckAnswer validates an answer body against its question. An empty body is allowed for
// optional questions (it means "cleared"); required-ness is enforced at submit time by
// MissingRequired, not here, so an autosave of a half-finished form is never rejected.
func (q Question) CheckAnswer(body string, format Format, lang Lang) error {
	if !lang.valid() {
		return invalid("invalid language %q", lang)
	}
	if format != FormatPlain && format != FormatRich {
		return invalid("invalid format %q", format)
	}
	if format == FormatRich && q.Type != TypeLongText {
		return invalid("question %q does not accept rich text", q.Key)
	}
	// Length is counted in characters, not bytes, so Amharic (3 bytes per letter) is not penalised.
	if utf8.RuneCountInString(body) > q.MaxLength {
		return invalid("answer to %q is longer than %d characters", q.Key, q.MaxLength)
	}
	if q.Type == TypeChoice && body != "" {
		for _, o := range q.Options {
			if o.Value == body {
				return nil
			}
		}
		return invalid("answer to %q is not one of the options", q.Key)
	}
	return nil
}

// MissingRequired returns the keys of required questions in r that have no non-blank answer.
func (r Revision) MissingRequired(answers []Answer) []string {
	answered := make(map[string]bool, len(answers))
	for _, a := range answers {
		if strings.TrimSpace(a.Body) != "" {
			answered[a.QuestionKey] = true
		}
	}
	var missing []string
	for _, q := range r.Questions {
		if q.Required && !answered[q.Key] {
			missing = append(missing, q.Key)
		}
	}
	return missing
}

// Unanswered returns the keys in r that the artist has not answered at all. This is what a nudge
// asks about: questions that carried forward unchanged are already answered and never appear here.
func (r Revision) Unanswered(answers []Answer) []string {
	answered := make(map[string]bool, len(answers))
	for _, a := range answers {
		if strings.TrimSpace(a.Body) != "" {
			answered[a.QuestionKey] = true
		}
	}
	var out []string
	for _, q := range r.Questions {
		if !answered[q.Key] {
			out = append(out, q.Key)
		}
	}
	return out
}
