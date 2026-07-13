package requestauth

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mq/api/internal/testutil/assist"
)

func TestUserIDFromHeader(t *testing.T) {
	id := uuid.New()

	got, err := UserIDFromHeader(id.String())
	assist.NoError(t, err)
	assist.Equal(t, id, got)

	_, err = UserIDFromHeader("")
	assist.Error(t, err)

	_, err = UserIDFromHeader("bad")
	assist.Error(t, err)
}
