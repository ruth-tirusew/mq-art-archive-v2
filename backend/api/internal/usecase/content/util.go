package content

import (
	"fmt"
	"strings"
	"unicode"

	domain "github.com/mq/api/internal/domain/content"
)

func slugify(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

func requireTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title is required")
	}
	return nil
}

func estimateReadingTime(body string) int {
	words := 0
	inWord := false
	for _, r := range body {
		if unicode.IsSpace(r) {
			inWord = false
			continue
		}
		if !inWord {
			words++
			inWord = true
		}
	}
	minutes := (words + 199) / 200
	if minutes < 1 {
		return 1
	}
	return minutes
}

func resolveCategory(category string) string {
	if strings.TrimSpace(category) == "" {
		return "General"
	}
	return strings.TrimSpace(category)
}

func resolveDifficulty(difficulty string) string {
	if strings.TrimSpace(difficulty) == "" {
		return "Beginner"
	}
	return strings.TrimSpace(difficulty)
}

func validStatus(status domain.ArticleStatus) bool {
	switch status {
	case domain.ArticleStatusDraft, domain.ArticleStatusPublished, domain.ArticleStatusArchived:
		return true
	default:
		return false
	}
}
