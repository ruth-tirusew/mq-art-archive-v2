package onboarding

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MinApplicationWorks is how many pieces of work an artist must show to apply. MaxApplicationWorks
// keeps a draft from growing without limit.
const (
	MinApplicationWorks = 5
	MaxApplicationWorks = 20
)

// ApplicantWork is one piece of work an applicant has uploaded to show with their application.
// The JSON shape is also what is frozen onto the application at submission, so keep the tags stable:
// stored applications are read back through them.
type ApplicantWork struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"-"`
	MediaAssetID uuid.UUID `json:"media_asset_id"`
	// URL and Blurhash describe the uploaded image; they are filled from its asset record.
	URL       string    `json:"url"`
	Blurhash  string    `json:"blurhash,omitempty"`
	Title     string    `json:"title"`
	Medium    string    `json:"medium,omitempty"`
	Year      *int      `json:"year,omitempty"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
}

// Validate checks what the applicant typed. The image itself is checked by whoever stores it.
func (w ApplicantWork) Validate(now time.Time) error {
	if strings.TrimSpace(w.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if len([]rune(w.Title)) > 200 {
		return fmt.Errorf("title must be 200 characters or fewer")
	}
	if len([]rune(w.Medium)) > 100 {
		return fmt.Errorf("medium must be 100 characters or fewer")
	}
	if w.Year != nil && (*w.Year < 1000 || *w.Year > now.Year()+1) {
		return fmt.Errorf("year must be between 1000 and %d", now.Year()+1)
	}
	if w.MediaAssetID == uuid.Nil {
		return fmt.Errorf("an uploaded image is required")
	}
	return nil
}

// WorksShortfall says how many more works are needed, or 0 when there are enough.
func WorksShortfall(have int) int {
	if have >= MinApplicationWorks {
		return 0
	}
	return MinApplicationWorks - have
}
