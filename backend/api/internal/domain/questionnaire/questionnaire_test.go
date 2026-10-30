package questionnaire

import (
	"errors"
	"strings"
	"testing"

	"github.com/mq/api/internal/domain/apperrors"
)

func q(key string, pos int, t QuestionType) Question {
	return Question{Key: key, Position: pos, Type: t, PromptEN: "Prompt", MaxLength: 100}
}

func TestRevisionValidate(t *testing.T) {
	choice := q("medium", 2, TypeChoice)
	choice.Options = []Option{{Value: "a", LabelEN: "A"}, {Value: "b", LabelEN: "B"}}

	tests := []struct {
		name string
		r    Revision
		ok   bool
	}{
		{"valid", Revision{Questions: []Question{q("bio", 1, TypeLongText), choice}}, true},
		{"empty", Revision{}, false},
		{"duplicate key", Revision{Questions: []Question{q("bio", 1, TypeLongText), q("bio", 2, TypeLongText)}}, false},
		{"duplicate position", Revision{Questions: []Question{q("a", 1, TypeShortText), q("b", 1, TypeShortText)}}, false},
		{"bad key", Revision{Questions: []Question{q("Has Space", 1, TypeShortText)}}, false},
		{"choice without options", Revision{Questions: []Question{q("c", 1, TypeChoice)}}, false},
		{"options on text", Revision{Questions: []Question{func() Question {
			x := q("t", 1, TypeShortText)
			x.Options = []Option{{Value: "a", LabelEN: "A"}, {Value: "b", LabelEN: "B"}}
			return x
		}()}}, false},
	}
	for _, tc := range tests {
		err := tc.r.Validate()
		if (err == nil) != tc.ok {
			t.Errorf("%s: err=%v, want ok=%v", tc.name, err, tc.ok)
		}
		if err != nil && !errors.Is(err, apperrors.ErrValidation) {
			t.Errorf("%s: error should wrap ErrValidation: %v", tc.name, err)
		}
	}
}

func TestCheckAnswer(t *testing.T) {
	long := q("bio", 1, TypeLongText)
	long.MaxLength = 5
	short := q("name", 2, TypeShortText)
	choice := q("m", 3, TypeChoice)
	choice.Options = []Option{{Value: "a", LabelEN: "A"}, {Value: "b", LabelEN: "B"}}

	// 7 runes against a max of 5 must fail: length is counted in characters, not bytes.
	if err := long.CheckAnswer("እንደምን አለህ", FormatPlain, LangAmharic); err == nil {
		t.Error("expected 7-rune answer to exceed max 5")
	}
	if err := long.CheckAnswer("ሰላም", FormatPlain, LangAmharic); err != nil {
		t.Errorf("4 Amharic runes (12 bytes) should fit max 5: %v", err)
	}
	if err := short.CheckAnswer("x", FormatRich, LangEnglish); err == nil {
		t.Error("rich text on a short_text question should be rejected")
	}
	if err := long.CheckAnswer("hi", FormatRich, LangEnglish); err != nil {
		t.Errorf("rich text on long_text should be accepted: %v", err)
	}
	if err := choice.CheckAnswer("zzz", FormatPlain, LangEnglish); err == nil {
		t.Error("answer outside options should be rejected")
	}
	if err := choice.CheckAnswer("", FormatPlain, LangEnglish); err != nil {
		t.Errorf("clearing an optional choice should be allowed: %v", err)
	}
	if err := short.CheckAnswer("x", FormatPlain, Lang("fr")); err == nil {
		t.Error("unknown language should be rejected")
	}
	if err := short.CheckAnswer(strings.Repeat("a", 101), FormatPlain, LangEnglish); err == nil {
		t.Error("over-long answer should be rejected")
	}
}

func TestMissingRequiredAndUnanswered(t *testing.T) {
	a, b, c := q("a", 1, TypeShortText), q("b", 2, TypeShortText), q("c", 3, TypeShortText)
	a.Required, b.Required = true, true
	r := Revision{Questions: []Question{a, b, c}}
	answers := []Answer{{QuestionKey: "a", Body: "x"}, {QuestionKey: "b", Body: "   "}}

	if got := r.MissingRequired(answers); len(got) != 1 || got[0] != "b" {
		t.Errorf("MissingRequired = %v, want [b] (blank counts as missing)", got)
	}
	if got := r.Unanswered(answers); len(got) != 2 || got[0] != "b" || got[1] != "c" {
		t.Errorf("Unanswered = %v, want [b c]", got)
	}
}
