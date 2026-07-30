package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/testutil/assist"
)

type mockAuth struct {
	beginOAuth    func(ctx context.Context, provider, returnTo string) (string, string, error)
	completeOAuth func(ctx context.Context, provider, code, state string) (string, *identity.User, string, error)
	register      func(ctx context.Context, email, password string) (string, *identity.User, error)
	login         func(ctx context.Context, email, password string) (string, *identity.User, error)
}

func (m *mockAuth) BeginOAuth(ctx context.Context, provider, returnTo string) (string, string, error) {
	return m.beginOAuth(ctx, provider, returnTo)
}

func (m *mockAuth) CompleteOAuth(ctx context.Context, provider, code, state string) (string, *identity.User, string, error) {
	return m.completeOAuth(ctx, provider, code, state)
}

func (m *mockAuth) Register(ctx context.Context, email, password string) (string, *identity.User, error) {
	return m.register(ctx, email, password)
}

func (m *mockAuth) Login(ctx context.Context, email, password string) (string, *identity.User, error) {
	return m.login(ctx, email, password)
}

func (m *mockAuth) ForgotPassword(ctx context.Context, email string) error {
	return nil
}

func (m *mockAuth) ResetPassword(ctx context.Context, token, newPassword string) error {
	return nil
}
func (m *mockAuth) VerifyEmail(context.Context, string) error                { return nil }
func (m *mockAuth) ResendEmailVerification(context.Context, uuid.UUID) error { return nil }

func (m *mockAuth) GetMe(ctx context.Context, userID uuid.UUID) (*identity.User, error) {
	return &identity.User{ID: userID, Email: "user@example.com", Role: identity.RolePublic, HasPassword: true}, nil
}

func (m *mockAuth) UpdateMyProfile(ctx context.Context, userID uuid.UUID, displayName, avatarURL string) (*identity.User, error) {
	return &identity.User{ID: userID, Email: "user@example.com", Role: identity.RolePublic, DisplayName: displayName, AvatarURL: avatarURL, HasPassword: true}, nil
}

func (m *mockAuth) ChangeEmail(ctx context.Context, userID uuid.UUID, newEmail, currentPassword string) (*identity.User, error) {
	return &identity.User{ID: userID, Email: newEmail, Role: identity.RolePublic, HasPassword: true}, nil
}

func (m *mockAuth) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	return nil
}

func (m *mockAuth) GetNotificationPreferences(ctx context.Context, userID uuid.UUID) (*identity.NotificationPreferences, error) {
	prefs := identity.DefaultNotificationPreferences(userID)
	return &prefs, nil
}

func (m *mockAuth) UpdateNotificationPreferences(ctx context.Context, userID uuid.UUID, prefs identity.NotificationPreferences) (*identity.NotificationPreferences, error) {
	prefs.UserID = userID
	return &prefs, nil
}
func (m *mockAuth) ListUsers(context.Context, *identity.Role, int, int) ([]identity.User, int, error) {
	return nil, 0, nil
}
func (m *mockAuth) UpdateUserRole(context.Context, uuid.UUID, uuid.UUID, identity.Role) (*identity.User, error) {
	return nil, nil
}

func TestAuthHandler_GoogleLogin_success(t *testing.T) {
	h := NewAuthHandler(&mockAuth{
		beginOAuth: func(ctx context.Context, provider, returnTo string) (string, string, error) {
			assist.Equal(t, "google", provider)
			assist.Equal(t, "http://localhost:5174/auth/callback", returnTo)
			return "https://accounts.google.com/o/oauth2/auth", "state", nil
		},
	}, "mq_access_token")

	w := serve(t, http.MethodGet, "/api/v1/auth/google?return_to=http://localhost:5174/auth/callback", nil, nil, nil, h.GoogleLogin)
	assist.Equal(t, http.StatusFound, w.Code)
}

func TestAuthHandler_GoogleLogin_missingReturnTo(t *testing.T) {
	h := NewAuthHandler(&mockAuth{}, "mq_access_token")
	w := serve(t, http.MethodGet, "/api/v1/auth/google", nil, nil, nil, h.GoogleLogin)
	assist.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_GoogleCallback_success(t *testing.T) {
	userID := uuid.New()
	h := NewAuthHandler(&mockAuth{
		completeOAuth: func(ctx context.Context, provider, code, state string) (string, *identity.User, string, error) {
			return "token", &identity.User{ID: userID, Email: "admin@example.com", Role: identity.RoleAdmin},
				"http://localhost:5174/", nil
		},
	}, "mq_access_token")

	w := serve(t, http.MethodGet, "/api/v1/auth/google/callback?code=abc&state=xyz", nil, nil, nil, h.GoogleCallback)
	assist.Equal(t, http.StatusFound, w.Code)
}

func TestAuthHandler_GoogleCallback_missingParams(t *testing.T) {
	h := NewAuthHandler(&mockAuth{}, "mq_access_token")
	w := serve(t, http.MethodGet, "/api/v1/auth/google/callback", nil, nil, nil, h.GoogleCallback)
	assist.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Logout(t *testing.T) {
	h := NewAuthHandler(&mockAuth{}, "mq_access_token")
	w := serve(t, http.MethodPost, "/api/v1/auth/logout", nil, nil, nil, h.Logout)
	// Gin writes 200 once Set-Cookie is applied after Status(204).
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_Me_success(t *testing.T) {
	userID := uuid.New()
	h := NewAuthHandler(&mockAuth{}, "mq_access_token")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	c.Set(requestauth.ContextUserID, userID)
	c.Set(requestauth.ContextUserRole, identity.RoleAdmin)
	c.Set(requestauth.ContextUserEmail, "admin@example.com")
	h.Me(c)

	assist.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_Me_unauthorized(t *testing.T) {
	h := NewAuthHandler(&mockAuth{}, "mq_access_token")
	w := serve(t, http.MethodGet, "/api/v1/auth/me", nil, nil, nil, h.Me)
	assist.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWriteAuthError_mappings(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"unauthorized", apperrors.ErrUnauthorized, http.StatusUnauthorized},
		{"forbidden", apperrors.ErrForbidden, http.StatusForbidden},
		{"invalid token", apperrors.ErrInvalidToken, http.StatusBadRequest},
		{"internal", errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			writeAuthError(c, tc.err)
			assist.Equal(t, tc.want, w.Code)
		})
	}
}

func TestOnboardingHandler_GetMyApplication_success(t *testing.T) {
	userID := uuid.New()
	appID := uuid.New()
	h := NewOnboardingHandler(&mockOnboarding{
		getMyApplication: func(ctx context.Context, applicantID uuid.UUID) (*onboarding.OnboardingApplication, error) {
			assist.Equal(t, userID, applicantID)
			return &onboarding.OnboardingApplication{ID: appID, ApplicantID: applicantID, DisplayName: "Studio"}, nil
		},
	})

	w := serve(t, http.MethodGet, "/api/v1/applications/me", nil, nil,
		map[string]string{"X-User-ID": userID.String()}, h.GetMyApplication)
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestOnboardingHandler_Submit_invalidJSON(t *testing.T) {
	h := NewOnboardingHandler(&mockOnboarding{})
	w := serve(t, http.MethodPost, "/api/v1/applications", strings.NewReader("{"),
		nil, map[string]string{"X-User-ID": uuid.New().String()}, h.Submit)
	assist.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_GoogleLogin_forbidden(t *testing.T) {
	h := NewAuthHandler(&mockAuth{
		beginOAuth: func(ctx context.Context, provider, returnTo string) (string, string, error) {
			return "", "", apperrors.ErrForbidden
		},
	}, "mq_access_token")
	w := serve(t, http.MethodGet, "/api/v1/auth/google?return_to=http://localhost:5174", nil, nil, nil, h.GoogleLogin)
	assist.Equal(t, http.StatusForbidden, w.Code)
}

func TestAuthHandler_Me_responseShape(t *testing.T) {
	userID := uuid.New()
	h := NewAuthHandler(&mockAuth{}, "mq_access_token")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	c.Set(requestauth.ContextUserID, userID)
	h.Me(c)

	assist.Equal(t, http.StatusOK, w.Code)
	var resp dto.UserResponse
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assist.Equal(t, userID.String(), resp.ID)
	assist.Equal(t, true, resp.HasPassword)
}

func TestAuthHandler_Register_success(t *testing.T) {
	userID := uuid.New()
	h := NewAuthHandler(&mockAuth{
		register: func(ctx context.Context, email, password string) (string, *identity.User, error) {
			assist.Equal(t, "new@example.com", email)
			assist.Equal(t, "password1", password)
			return "token", &identity.User{ID: userID, Email: email, Role: identity.RolePublic}, nil
		},
	}, "mq_access_token")

	body := strings.NewReader(`{"email":"new@example.com","password":"password1"}`)
	w := serve(t, http.MethodPost, "/api/v1/auth/register", body, nil, nil, h.Register)
	assist.Equal(t, http.StatusCreated, w.Code)
	assist.Contains(t, w.Header().Get("Set-Cookie"), "mq_access_token=")

	var resp dto.UserResponse
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assist.Equal(t, userID.String(), resp.ID)
	assist.Equal(t, "new@example.com", resp.Email)
}

func TestAuthHandler_Register_conflict(t *testing.T) {
	h := NewAuthHandler(&mockAuth{
		register: func(ctx context.Context, email, password string) (string, *identity.User, error) {
			return "", nil, apperrors.ErrConflict
		},
	}, "mq_access_token")

	body := strings.NewReader(`{"email":"new@example.com","password":"password1"}`)
	w := serve(t, http.MethodPost, "/api/v1/auth/register", body, nil, nil, h.Register)
	assist.Equal(t, http.StatusConflict, w.Code)
}

func TestAuthHandler_Login_success(t *testing.T) {
	userID := uuid.New()
	h := NewAuthHandler(&mockAuth{
		login: func(ctx context.Context, email, password string) (string, *identity.User, error) {
			return "token", &identity.User{ID: userID, Email: email, Role: identity.RoleArtist}, nil
		},
	}, "mq_access_token")

	body := strings.NewReader(`{"email":"user@example.com","password":"password1"}`)
	w := serve(t, http.MethodPost, "/api/v1/auth/login", body, nil, nil, h.Login)
	assist.Equal(t, http.StatusOK, w.Code)
	assist.Contains(t, w.Header().Get("Set-Cookie"), "mq_access_token=")
}

func TestAuthHandler_Login_unauthorized(t *testing.T) {
	h := NewAuthHandler(&mockAuth{
		login: func(ctx context.Context, email, password string) (string, *identity.User, error) {
			return "", nil, apperrors.ErrUnauthorized
		},
	}, "mq_access_token")

	body := strings.NewReader(`{"email":"user@example.com","password":"wrong"}`)
	w := serve(t, http.MethodPost, "/api/v1/auth/login", body, nil, nil, h.Login)
	assist.Equal(t, http.StatusUnauthorized, w.Code)
}
