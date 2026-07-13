package content

import (
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
