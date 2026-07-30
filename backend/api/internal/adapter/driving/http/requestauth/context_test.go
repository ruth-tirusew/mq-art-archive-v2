package requestauth

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/testutil/assist"
)

func TestUserIDFromContext_success(t *testing.T) {
	id := uuid.New()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(ContextUserID, id)

	got, err := UserIDFromContext(c)
	assist.NoError(t, err)
	assist.Equal(t, id, got)
}

func TestUserIDFromContext_missing(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	_, err := UserIDFromContext(c)
	assist.Error(t, err)
}

func TestUserRoleFromContext_success(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(ContextUserRole, identity.RoleAdmin)

	got, err := UserRoleFromContext(c)
	assist.NoError(t, err)
	assist.Equal(t, identity.RoleAdmin, got)
}

func TestUserRoleFromContext_missing(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	_, err := UserRoleFromContext(c)
	assist.Error(t, err)
}

func TestUnauthorizedError(t *testing.T) {
	assist.Error(t, UnauthorizedError())
}
