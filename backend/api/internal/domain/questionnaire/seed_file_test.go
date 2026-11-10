package questionnaire

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"
)

// seedFile is the shape the admin API's "save draft" endpoint accepts, which is also what
// scripts/seed_questionnaire.py sends.
type seedFile struct {
	ChangeKind string `json:"change_kind"`
	Questions  []struct {
		Key       string   `json:"key"`
		Type      string   `json:"type"`
		PromptEN  string   `json:"prompt_en"`
		PromptAM  string   `json:"prompt_am"`
		HelpEN    string   `json:"help_en"`
		HelpAM    string   `json:"help_am"`
		Required  bool     `json:"required"`
		IsPublic  bool     `json:"is_public"`
		MaxLength int      `json:"max_length"`
		Options   []Option `json:"options"`
	} `json:"questions"`
}

func loadSeed(t *testing.T, name string) seedFile {
	t.Helper()
	_, here, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(here), "..", "..", "..", "..", "..")
	raw, err := os.ReadFile(filepath.Join(root, "docs", "questionnaire", name))
	if err != nil {
		t.Fatalf("read seed: %v", err)
	}
	var s seedFile
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatalf("parse seed: %v", err)
	}
	return s
}

func TestArtivArchiveSeed_isAValidRevision(t *testing.T) {
	seed := loadSeed(t, "artiv-archive-v1.1.json")
	rev := Revision{}
	for i, q := range seed.Questions {
		max := q.MaxLength
		if max == 0 {
			max = DefaultMaxLength(QuestionType(q.Type))
		}
		rev.Questions = append(rev.Questions, Question{
			Key: q.Key, Position: i + 1, Type: QuestionType(q.Type), PromptEN: q.PromptEN, Required: q.Required,
			IsPublic: q.IsPublic, MaxLength: max, Options: q.Options,
		})
		// The API's request limits (dto.QuestionRequest): reject here so a seed that the server would refuse is caught in CI.
		if utf8.RuneCountInString(q.PromptEN) > 500 || utf8.RuneCountInString(q.PromptAM) > 500 {
			t.Errorf("%s: prompt longer than 500 characters", q.Key)
		}
		if utf8.RuneCountInString(q.HelpEN) > 1000 || utf8.RuneCountInString(q.HelpAM) > 1000 {
			t.Errorf("%s: help text longer than 1000 characters", q.Key)
		}
		if max > 20000 {
			t.Errorf("%s: max_length above 20000", q.Key)
		}
	}
	if err := rev.Validate(); err != nil {
		t.Fatalf("seed is not a valid revision: %v", err)
	}
	if seed.ChangeKind != "substantive" && seed.ChangeKind != "minor" {
		t.Errorf("change_kind %q", seed.ChangeKind)
	}
}

// The questionnaire promises that consent answers and private collector identities are never
// published. is_public is the only thing enforcing that, so pin it.
func TestArtivArchiveSeed_keepsSensitiveAnswersPrivate(t *testing.T) {
	seed := loadSeed(t, "artiv-archive-v1.1.json")
	private := map[string]bool{"q7_collections": true, "credit_preferences": true, "q15_refresh": true}
	byKey := map[string]bool{}
	for _, q := range seed.Questions {
		byKey[q.Key] = q.IsPublic
		if (strings.HasPrefix(q.Key, "consent_") || private[q.Key]) && q.IsPublic {
			t.Errorf("%s must be admin-only", q.Key)
		}
	}
	for key := range private {
		if _, ok := byKey[key]; !ok {
			t.Errorf("expected question %q in the seed", key)
		}
	}
}

func TestArtivArchiveSeed_requiresOnlyTheConsentToPublish(t *testing.T) {
	seed := loadSeed(t, "artiv-archive-v1.1.json")
	var required []string
	for _, q := range seed.Questions {
		if q.Required {
			required = append(required, q.Key)
		}
	}
	// The interview guide says artists may skip any question, so only the consent record is required.
	if len(required) != 1 || required[0] != "consent_publication" {
		t.Errorf("required = %v, want only consent_publication", required)
	}
}
