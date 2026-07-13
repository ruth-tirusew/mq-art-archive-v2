package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mq/api/config"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/adapter/driving/http/handler"
	"github.com/mq/api/internal/adapter/driving/http/middleware"
	"github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/domain/institution"
	"github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/domain/profile"
	"github.com/mq/api/internal/testutil/assist"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type stubContent struct{}

func (stubContent) ListPublished(ctx context.Context, filter content.ListFilter) ([]content.Article, error) {
	return []content.Article{{ID: uuid.New(), Slug: "seed", Title: "Seed", Status: content.ArticleStatusPublished}}, nil
}
func (stubContent) GetBySlug(ctx context.Context, slug string) (*content.Article, error) {
	return &content.Article{ID: uuid.New(), Slug: slug, Title: "Article", Status: content.ArticleStatusPublished}, nil
}
func (stubContent) CreateDraft(ctx context.Context, authorID uuid.UUID, title, body string) (*content.Article, error) {
	return &content.Article{ID: uuid.New(), Title: title, Slug: "draft", Status: content.ArticleStatusDraft}, nil
}

type stubProfile struct{}

func (stubProfile) GetArtistByHandle(ctx context.Context, handle string) (*profile.ArtistProfile, error) {
	return &profile.ArtistProfile{ID: uuid.New(), Handle: handle, Slug: handle, DisplayName: "Artist", Status: profile.ProfileStatusApproved}, nil
}
func (stubProfile) GetArtistBySlug(ctx context.Context, slug string) (*profile.ArtistProfile, error) {
	return &profile.ArtistProfile{ID: uuid.New(), Slug: slug, DisplayName: "Artist", Status: profile.ProfileStatusApproved}, nil
}
func (stubProfile) ListApproved(ctx context.Context, filter profile.ListFilter) ([]profile.ArtistProfile, error) {
	return []profile.ArtistProfile{{ID: uuid.New(), Slug: "artist", DisplayName: "Artist", Status: profile.ProfileStatusApproved}}, nil
}
func (stubProfile) GetArtistByID(ctx context.Context, id uuid.UUID) (*profile.ArtistProfile, error) {
	return &profile.ArtistProfile{ID: id, Slug: "artist"}, nil
}
func (stubProfile) GetArtistByUserID(ctx context.Context, userID uuid.UUID) (*profile.ArtistProfile, error) {
	return &profile.ArtistProfile{ID: uuid.New(), UserID: userID, Slug: "artist"}, nil
}
func (stubProfile) UpdateArtist(ctx context.Context, p profile.ArtistProfile) (*profile.ArtistProfile, error) {
	return &p, nil
}

type stubInstitution struct{}

func (stubInstitution) GetBySlug(ctx context.Context, slug string) (*institution.Institution, error) {
	return &institution.Institution{ID: uuid.New(), Slug: slug, Name: "Gallery", Status: institution.StatusApproved}, nil
}
func (stubInstitution) GetByID(ctx context.Context, id uuid.UUID) (*institution.Institution, error) {
	return &institution.Institution{ID: id, Slug: "gallery"}, nil
}
func (stubInstitution) Update(ctx context.Context, inst institution.Institution) (*institution.Institution, error) {
	return &inst, nil
}

type stubEvent struct{}

func (stubEvent) List(ctx context.Context, filter events.ListFilter) ([]events.Event, error) {
	return []events.Event{{ID: uuid.New(), Slug: "opening", Title: "Opening", Status: events.EventStatusApproved, StartsAt: time.Now().UTC()}}, nil
}
func (stubEvent) Search(ctx context.Context, query string, limit int) ([]events.Event, error) {
	return nil, nil
}
func (stubEvent) GetByID(ctx context.Context, id uuid.UUID) (*events.Event, error) {
	return &events.Event{ID: id, Slug: "event"}, nil
}
func (stubEvent) GetBySlug(ctx context.Context, slug string) (*events.Event, error) {
	return &events.Event{ID: uuid.New(), Slug: slug, Title: "Event", Status: events.EventStatusApproved, StartsAt: time.Now().UTC()}, nil
}
func (stubEvent) Review(ctx context.Context, id uuid.UUID, reviewerID uuid.UUID, status events.EventStatus, notes string) (*events.Event, error) {
	return &events.Event{ID: id, Status: status}, nil
}
func (stubEvent) Sync(ctx context.Context) (int, error) {
	return 0, nil
}

type stubArt struct{}

func (stubArt) ListByArtist(ctx context.Context, artistID uuid.UUID) ([]art.ArtPost, error) {
	return []art.ArtPost{{ID: uuid.New(), ArtistID: artistID, Title: "Work", Status: art.ArtStatusPublished, CreatedAt: time.Now().UTC()}}, nil
}

func (stubArt) ListPublished(ctx context.Context, filter art.ListFilter) ([]art.ArtPostWithArtist, error) {
	return []art.ArtPostWithArtist{{ArtPost: art.ArtPost{ID: uuid.New(), Title: "Work", Status: art.ArtStatusPublished, CreatedAt: time.Now().UTC()}, ArtistSlug: "artist", ArtistName: "Artist"}}, nil
}
func (stubArt) GetByID(ctx context.Context, id uuid.UUID) (*art.ArtPost, error) {
	return &art.ArtPost{ID: id, Title: "Work"}, nil
}
func (stubArt) CreateDraft(ctx context.Context, artistID uuid.UUID, title, description, medium string) (*art.ArtPost, error) {
	return &art.ArtPost{ID: uuid.New(), ArtistID: artistID, Title: title}, nil
}

type stubOnboarding struct{}

func (stubOnboarding) ListPending(ctx context.Context) ([]onboarding.OnboardingApplication, error) {
	return []onboarding.OnboardingApplication{{ID: uuid.New(), DisplayName: "Applicant", Status: onboarding.ApprovalStatusPending}}, nil
}
func (stubOnboarding) GetByID(ctx context.Context, id uuid.UUID) (*onboarding.OnboardingApplication, error) {
	return &onboarding.OnboardingApplication{ID: id, Status: onboarding.ApprovalStatusPending}, nil
}
func (stubOnboarding) Review(ctx context.Context, id, reviewerID uuid.UUID, status onboarding.ApprovalStatus, notes string) (*onboarding.OnboardingApplication, error) {
	return &onboarding.OnboardingApplication{ID: id, Status: status}, nil
}

type stubIdentity struct{}

func (stubIdentity) GetUser(_ context.Context, id uuid.UUID) (*identity.User, error) {
	return &identity.User{ID: id, Email: "stub@example.com", Role: identity.RoleAdmin}, nil
}

type stubAuth struct{}

func (stubAuth) BeginOAuth(_ context.Context, _, _ string) (string, string, error) {
	return "https://example.com/auth", "state", nil
}
func (stubAuth) CompleteOAuth(_ context.Context, _, _, _ string) (string, *identity.User, string, error) {
	return "token", &identity.User{ID: uuid.New(), Email: "user@example.com", Role: identity.RolePublic}, "http://localhost:5173/auth/callback", nil
}

func testRouter() *gin.Engine {
	cfg := config.Config{
		CORSOrigins:    []string{"http://localhost:5173"},
		AuthCookieName: "mq_access_token",
		AuthDevMode:    true,
	}
	return NewRouter(cfg, Handlers{
		Health:      handler.NewHealthHandler(),
		Article:     handler.NewArticleHandler(stubContent{}),
		Profile:     handler.NewProfileHandler(stubProfile{}),
		Institution: handler.NewInstitutionHandler(stubInstitution{}),
		Art:         handler.NewArtHandler(stubArt{}, stubProfile{}),
		Event:       handler.NewEventHandler(stubEvent{}),
		Onboarding:  handler.NewOnboardingHandler(stubOnboarding{}),
		Auth:        handler.NewAuthHandler(stubAuth{}, cfg.AuthCookieName),
	}, RouterDeps{
		Identity: stubIdentity{},
		Auth: middleware.AuthConfig{
			Verifier:   nil,
			Identity:   stubIdentity{},
			CookieName: cfg.AuthCookieName,
			DevMode:    true,
		},
	})
}

func serveRouter(t *testing.T, method, path string, body io.Reader, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	testRouter().ServeHTTP(w, req)
	return w
}

func TestRouter_health(t *testing.T) {
	w := serveRouter(t, http.MethodGet, "/health", nil, nil)

	assist.Equal(t, http.StatusOK, w.Code)
	var resp dto.HealthResponse
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assist.Equal(t, "ok", resp.Status)
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID response header")
	}
}

func TestRouter_listArticles(t *testing.T) {
	w := serveRouter(t, http.MethodGet, "/api/v1/articles", nil, nil)
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_getArticleBySlug(t *testing.T) {
	w := serveRouter(t, http.MethodGet, "/api/v1/articles/seed", nil, nil)
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_createArticle(t *testing.T) {
	authorID := uuid.New()
	w := serveRouter(t, http.MethodPost, "/api/v1/articles", strings.NewReader(`{"title":"New"}`),
		map[string]string{"X-User-ID": authorID.String()})
	assist.Equal(t, http.StatusCreated, w.Code)
}

func TestRouter_getArtistBySlug(t *testing.T) {
	w := serveRouter(t, http.MethodGet, "/api/v1/artists/abebe", nil, nil)
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_listArtistPosts(t *testing.T) {
	w := serveRouter(t, http.MethodGet, "/api/v1/artists/abebe/posts", nil, nil)
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_getInstitutionBySlug(t *testing.T) {
	w := serveRouter(t, http.MethodGet, "/api/v1/institutions/gallery", nil, nil)
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_getPostByID(t *testing.T) {
	postID := uuid.New()
	w := serveRouter(t, http.MethodGet, "/api/v1/posts/"+postID.String(), nil, nil)
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_createPost(t *testing.T) {
	userID := uuid.New()
	w := serveRouter(t, http.MethodPost, "/api/v1/posts", strings.NewReader(`{"title":"New Work"}`),
		map[string]string{"X-User-ID": userID.String()})
	assist.Equal(t, http.StatusCreated, w.Code)
}

func TestRouter_listPendingApplications(t *testing.T) {
	w := serveRouter(t, http.MethodGet, "/admin/v1/applications", nil, map[string]string{"X-User-ID": uuid.New().String()})
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_getApplicationByID(t *testing.T) {
	w := serveRouter(t, http.MethodGet, "/admin/v1/applications/"+uuid.New().String(), nil, map[string]string{"X-User-ID": uuid.New().String()})
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_reviewApplication(t *testing.T) {
	w := serveRouter(t, http.MethodPut, "/admin/v1/applications/"+uuid.New().String(),
		strings.NewReader(`{"status":"approved"}`),
		map[string]string{"X-User-ID": uuid.New().String()})
	assist.Equal(t, http.StatusOK, w.Code)
}
