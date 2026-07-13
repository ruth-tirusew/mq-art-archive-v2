package art

import (
	"testing"

	"github.com/mq/api/internal/testutil/assist"
)

func TestRequireTitle(t *testing.T) {
	assist.NoError(t, requireTitle("Sunset"))
	assist.Error(t, requireTitle(""))
	assist.Error(t, requireTitle("  "))
}
