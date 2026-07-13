package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driven/auth"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/testutil/assist"
)

type identityStub struct {
	users map[uuid.UUID]*identity.User
}

func (s *identityStub) GetUser(_ context.Context, id uuid.UUID) (*identity.User, error) {
	u, ok := s.users[id]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func TestAuthenticate_bearerToken(t *testing.T) {
	userID := uuid.New()
	tokenSvc := auth.NewTokenService("secret", time.Hour)
	token, err := tokenSvc.Issue(t.Context(), &identity.User{
		ID: userID, Email: "user@example.com", Role: identity.RoleArtist,
	})
	assist.NoError(t, err)

	identitySvc := &identityStub{
		users: map[uuid.UUID]*identity.User{
			userID: {ID: userID, Email: "user@example.com", Role: identity.RoleArtist},
		},
	}

	r := gin.New()
	r.Use(Authenticate(AuthConfig{
		Verifier:   tokenSvc,
		Identity:   identitySvc,
		CookieName: "mq_access_token",
	}))
	r.GET("/", func(c *gin.Context) {
		id, err := requestauth.UserIDFromContext(c)
		assist.NoError(t, err)
		assist.Equal(t, userID, id)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_forbidden(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(requestauth.ContextUserRole, identity.RolePublic)
		c.Next()
	}, RequireRole("admin"))
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	assist.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireRole_allowsAdmin(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(requestauth.ContextUserRole, identity.RoleAdmin)
		c.Next()
	}, RequireRole("admin"))
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	assist.Equal(t, http.StatusOK, w.Code)
}
