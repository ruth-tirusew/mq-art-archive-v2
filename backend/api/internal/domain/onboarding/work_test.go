package onboarding

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestApplicantWorkValidate(t *testing.T) {
	now := time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC)
	year := func(y int) *int { return &y }
	ok := ApplicantWork{MediaAssetID: uuid.New(), Title: "Untitled No. 3", Medium: "oil", Year: year(2024)}
	if err := ok.Validate(now); err != nil {
		t.Fatalf("valid work rejected: %v", err)
	}
	cases := map[string]func(*ApplicantWork){
		"blank title":  func(w *ApplicantWork) { w.Title = "   " },
		"long title":   func(w *ApplicantWork) { w.Title = strings.Repeat("a", 201) },
		"long medium":  func(w *ApplicantWork) { w.Medium = strings.Repeat("a", 101) },
		"year far off": func(w *ApplicantWork) { w.Year = year(1) },
		"future year":  func(w *ApplicantWork) { w.Year = year(2040) },
		"no image":     func(w *ApplicantWork) { w.MediaAssetID = uuid.Nil },
	}
	for name, mutate := range cases {
		w := ok
		mutate(&w)
		if err := w.Validate(now); err == nil {
			t.Errorf("%s should be rejected", name)
		}
	}
	if err := (ApplicantWork{MediaAssetID: uuid.New(), Title: "x"}).Validate(now); err != nil {
		t.Errorf("medium and year are optional: %v", err)
	}
}

func TestWorksShortfall(t *testing.T) {
	for have, want := range map[int]int{0: 5, 1: 4, 4: 1, 5: 0, 9: 0} {
		if got := WorksShortfall(have); got != want {
			t.Errorf("WorksShortfall(%d) = %d, want %d", have, got, want)
		}
	}
}
