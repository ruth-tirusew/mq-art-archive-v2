package content

import (
	"strings"
	"testing"

	"github.com/mq/api/internal/testutil/assist"
)

func TestSlugify(t *testing.T) {
	assist.Equal(t, "how-to-paint", slugify("How To Paint"))
	assist.Equal(t, "hello---world", slugify("  hello   world  "))
}

func TestRequireTitle(t *testing.T) {
	assist.NoError(t, requireTitle("valid"))
	assist.Error(t, requireTitle(""))
	assist.Error(t, requireTitle("   "))
}

func TestEstimateReadingTime(t *testing.T) {
	assist.Equal(t, 1, estimateReadingTime(""))
	assist.Equal(t, 1, estimateReadingTime("one two three"))
	words := make([]string, 250)
	for i := range words {
		words[i] = "word"
	}
	assist.Equal(t, 2, estimateReadingTime(strings.Join(words, " ")))
}
