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
	"github.com/mq/api/internal/domain/settings"
	"github.com/mq/api/internal/port/inbound"
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
func (stubContent) AdminList(ctx context.Context, status *content.ArticleStatus, limit, offset int) ([]content.Article, error) {
	return []content.Article{{ID: uuid.New(), Slug: "seed", Title: "Seed", Status: content.ArticleStatusPublished}}, nil
}
func (stubContent) AdminGet(ctx context.Context, id uuid.UUID) (*content.Article, error) {
	return &content.Article{ID: id, Slug: "seed", Title: "Article", Status: content.ArticleStatusDraft}, nil
}
func (stubContent) AdminCreate(ctx context.Context, authorID uuid.UUID, write content.ArticleWrite) (*content.Article, error) {
	return &content.Article{ID: uuid.New(), Title: write.Title, Slug: "draft", Body: write.Body, Status: content.ArticleStatusDraft}, nil
}
func (stubContent) AdminUpdate(ctx context.Context, id, editorID uuid.UUID, write content.ArticleWrite) (*content.Article, error) {
	return &content.Article{ID: id, Title: write.Title, Slug: "updated", Body: write.Body, Status: content.ArticleStatusDraft, Version: 2}, nil
}
func (stubContent) AdminSetStatus(ctx context.Context, id uuid.UUID, status *content.ArticleStatus, verified *bool) (*content.Article, error) {
	article := &content.Article{ID: id, Title: "Article", Slug: "seed", Status: content.ArticleStatusDraft, Version: 1}
	if status != nil {
		article.Status = *status
	}
	if verified != nil {
		article.Verified = *verified
	}
	return article, nil
}
func (stubContent) AdminListRevisions(ctx context.Context, articleID uuid.UUID, limit, offset int) ([]content.ArticleRevision, error) {
	return []content.ArticleRevision{}, nil
}
func (stubContent) AdminGetRevision(ctx context.Context, articleID uuid.UUID, version int) (*content.ArticleRevision, error) {
	return &content.ArticleRevision{ArticleID: articleID, Version: version, Title: "Old"}, nil
}
func (stubContent) AdminRestoreRevision(ctx context.Context, articleID uuid.UUID, version int, editorID uuid.UUID) (*content.Article, error) {
	return &content.Article{ID: articleID, Title: "Restored", Version: version + 1}, nil
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
func (stubProfile) ListAll(ctx context.Context, status *profile.ProfileStatus, limit, offset int) ([]profile.ArtistProfile, error) {
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
func (stubProfile) UpdateOwnProfile(ctx context.Context, userID uuid.UUID, update profile.OwnProfileUpdate) (*profile.ArtistProfile, error) {
	return &profile.ArtistProfile{ID: uuid.New(), UserID: userID, DisplayName: update.DisplayName, Slug: "artist"}, nil
}
func (stubProfile) AdminCreate(ctx context.Context, write profile.AdminArtistWrite) (*profile.ArtistProfile, error) {
	return &profile.ArtistProfile{ID: uuid.New(), DisplayName: write.DisplayName, Slug: "artist", Status: profile.ProfileStatusDraft}, nil
}
func (stubProfile) AdminUpdateContent(ctx context.Context, id uuid.UUID, write profile.AdminArtistWrite) (*profile.ArtistProfile, error) {
	return &profile.ArtistProfile{ID: id, DisplayName: write.DisplayName, Slug: "artist"}, nil
}
func (stubProfile) AdminDelete(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (stubProfile) AdminUpdate(ctx context.Context, id uuid.UUID, status *profile.ProfileStatus, featured *bool) (*profile.ArtistProfile, error) {
	p := &profile.ArtistProfile{ID: id, Slug: "artist"}
	if status != nil {
		p.Status = *status
	}
	if featured != nil {
		p.Featured = *featured
	}
	return p, nil
}

type stubInstitution struct{}

func (stubInstitution) GetBySlug(ctx context.Context, slug string) (*institution.Institution, error) {
	return &institution.Institution{ID: uuid.New(), Slug: slug, Name: "Gallery", Status: institution.InstitutionStatusApproved}, nil
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
func (stubEvent) ListPending(ctx context.Context) ([]events.Event, error) {
	return []events.Event{{ID: uuid.New(), Slug: "pending", Title: "Pending", Status: events.EventStatusPending, StartsAt: time.Now().UTC()}}, nil
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
func (stubEvent) AdminCreate(ctx context.Context, write events.EventWrite) (*events.Event, error) {
	return &events.Event{ID: uuid.New(), Title: write.Title, Status: events.EventStatusPending, StartsAt: write.StartsAt}, nil
}
func (stubEvent) Submit(ctx context.Context, write events.EventWrite) (*events.Event, error) {
	return &events.Event{ID: uuid.New(), Title: write.Title, Status: events.EventStatusPending, StartsAt: write.StartsAt}, nil
}
func (stubEvent) AdminUpdateContent(ctx context.Context, id uuid.UUID, write events.EventWrite) (*events.Event, error) {
	return &events.Event{ID: id, Title: write.Title, StartsAt: write.StartsAt}, nil
}
func (stubEvent) AdminDelete(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (stubEvent) Review(ctx context.Context, id uuid.UUID, reviewerID uuid.UUID, status events.EventStatus, notes string) (*events.Event, error) {
	return &events.Event{ID: id, Status: status}, nil
}
func (stubEvent) Sync(ctx context.Context) (int, error) {
	return 0, nil
}

type stubSearch struct{}

func (stubSearch) Search(ctx context.Context, query string, limit int) (*inbound.SearchResults, error) {
	return &inbound.SearchResults{}, nil
}

type stubArt struct{}

func (stubArt) ListByArtist(ctx context.Context, artistID uuid.UUID) ([]art.ArtPost, error) {
	return []art.ArtPost{{ID: uuid.New(), ArtistID: artistID, Title: "Work", Status: art.ArtStatusPublished, CreatedAt: time.Now().UTC()}}, nil
}

func (stubArt) ListOwnedByArtist(ctx context.Context, artistID uuid.UUID) ([]art.ArtPost, error) {
	return []art.ArtPost{{ID: uuid.New(), ArtistID: artistID, Title: "Draft Work", Status: art.ArtStatusDraft, CreatedAt: time.Now().UTC()}}, nil
}

func (stubArt) ListPublished(ctx context.Context, filter art.ListFilter) ([]art.ArtPostWithArtist, error) {
	return []art.ArtPostWithArtist{{ArtPost: art.ArtPost{ID: uuid.New(), Title: "Work", Status: art.ArtStatusPublished, CreatedAt: time.Now().UTC()}, ArtistSlug: "artist", ArtistName: "Artist"}}, nil
}

func (stubArt) ListAll(ctx context.Context, status *art.ArtStatus, limit, offset int) ([]art.ArtPostWithArtist, error) {
	return []art.ArtPostWithArtist{{ArtPost: art.ArtPost{ID: uuid.New(), Title: "Work", Status: art.ArtStatusPublished, CreatedAt: time.Now().UTC()}, ArtistSlug: "artist", ArtistName: "Artist"}}, nil
}

func (stubArt) GetByID(ctx context.Context, id uuid.UUID) (*art.ArtPost, error) {
	return &art.ArtPost{ID: id, Title: "Work"}, nil
}

func (stubArt) GetOwned(ctx context.Context, artistID, postID uuid.UUID) (*art.ArtPost, error) {
	return &art.ArtPost{ID: postID, ArtistID: artistID, Title: "Work"}, nil
}

func (stubArt) CreateDraft(ctx context.Context, artistID uuid.UUID, write art.ArtPostWrite) (*art.ArtPost, error) {
	return &art.ArtPost{ID: uuid.New(), ArtistID: artistID, Title: write.Title}, nil
}

func (stubArt) UpdateOwned(ctx context.Context, artistID, postID uuid.UUID, write art.ArtPostWrite) (*art.ArtPost, error) {
	return &art.ArtPost{ID: postID, ArtistID: artistID, Title: write.Title}, nil
}

func (stubArt) PublishOwned(ctx context.Context, artistID, postID uuid.UUID) (*art.ArtPost, error) {
	return &art.ArtPost{ID: postID, ArtistID: artistID, Title: "Work", Status: art.ArtStatusPublished}, nil
}

func (stubArt) UnpublishOwned(ctx context.Context, artistID, postID uuid.UUID) (*art.ArtPost, error) {
	return &art.ArtPost{ID: postID, ArtistID: artistID, Title: "Work", Status: art.ArtStatusDraft}, nil
}

func (stubArt) ArchiveOwned(ctx context.Context, artistID, postID uuid.UUID) (*art.ArtPost, error) {
	return &art.ArtPost{ID: postID, ArtistID: artistID, Title: "Work", Status: art.ArtStatusArchived}, nil
}

func (stubArt) DeleteOwned(ctx context.Context, artistID, postID uuid.UUID) error {
	return nil
}

func (stubArt) AdminCreate(ctx context.Context, artistID uuid.UUID, write art.ArtPostWrite, status *art.ArtStatus) (*art.ArtPost, error) {
	st := art.ArtStatusDraft
	if status != nil {
		st = *status
	}
	return &art.ArtPost{ID: uuid.New(), ArtistID: artistID, Title: write.Title, Status: st}, nil
}

func (stubArt) AdminUpdateContent(ctx context.Context, postID uuid.UUID, write art.ArtPostWrite) (*art.ArtPost, error) {
	return &art.ArtPost{ID: postID, Title: write.Title}, nil
}

func (stubArt) AdminDelete(ctx context.Context, postID uuid.UUID) error {
	return nil
}

func (stubArt) AdminUpdate(ctx context.Context, postID uuid.UUID, status *art.ArtStatus, featured *bool) (*art.ArtPost, error) {
	post := &art.ArtPost{ID: postID, Title: "Work", Status: art.ArtStatusPublished}
	if status != nil {
		post.Status = *status
	}
	if featured != nil {
		post.FeaturedAcquisition = *featured
	}
	return post, nil
}

type stubOnboarding struct{}

func (stubOnboarding) ListPending(ctx context.Context) ([]onboarding.OnboardingApplication, error) {
	return []onboarding.OnboardingApplication{{ID: uuid.New(), DisplayName: "Applicant", Status: onboarding.ApprovalStatusPending}}, nil
}
func (stubOnboarding) GetByID(ctx context.Context, id uuid.UUID) (*onboarding.OnboardingApplication, error) {
	return &onboarding.OnboardingApplication{ID: id, Status: onboarding.ApprovalStatusPending}, nil
}
func (stubOnboarding) GetMyApplication(ctx context.Context, applicantID uuid.UUID) (*onboarding.OnboardingApplication, error) {
	return &onboarding.OnboardingApplication{ID: uuid.New(), ApplicantID: applicantID, Status: onboarding.ApprovalStatusPending}, nil
}
func (stubOnboarding) Submit(ctx context.Context, applicantID uuid.UUID, applicantType onboarding.ApplicantType, displayName, notes string) (*onboarding.OnboardingApplication, error) {
	return &onboarding.OnboardingApplication{ID: uuid.New(), ApplicantID: applicantID, ApplicantType: applicantType, DisplayName: displayName, Status: onboarding.ApprovalStatusPending}, nil
}
func (stubOnboarding) SubmitWithHandle(ctx context.Context, applicantID uuid.UUID, applicantType onboarding.ApplicantType, displayName, handle, notes string) (*onboarding.OnboardingApplication, error) {
	return &onboarding.OnboardingApplication{ID: uuid.New(), ApplicantID: applicantID, ApplicantType: applicantType, DisplayName: displayName, RequestedHandle: handle, Status: onboarding.ApprovalStatusPending}, nil
}
func (stubOnboarding) CheckHandleAvailable(context.Context, string) (bool, error) { return true, nil }
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
func (stubAuth) Register(_ context.Context, email, _ string) (string, *identity.User, error) {
	return "token", &identity.User{ID: uuid.New(), Email: email, Role: identity.RolePublic}, nil
}
func (stubAuth) Login(_ context.Context, email, _ string) (string, *identity.User, error) {
	return "token", &identity.User{ID: uuid.New(), Email: email, Role: identity.RolePublic}, nil
}
func (stubAuth) ForgotPassword(_ context.Context, _ string) error         { return nil }
func (stubAuth) ResetPassword(_ context.Context, _, _ string) error       { return nil }
func (stubAuth) VerifyEmail(context.Context, string) error                { return nil }
func (stubAuth) ResendEmailVerification(context.Context, uuid.UUID) error { return nil }
func (stubAuth) GetMe(_ context.Context, id uuid.UUID) (*identity.User, error) {
	return &identity.User{ID: id, Email: "user@example.com", Role: identity.RolePublic, HasPassword: true}, nil
}
func (stubAuth) UpdateMyProfile(_ context.Context, id uuid.UUID, displayName, avatarURL string) (*identity.User, error) {
	return &identity.User{ID: id, Email: "user@example.com", Role: identity.RolePublic, DisplayName: displayName, AvatarURL: avatarURL, HasPassword: true}, nil
}
func (stubAuth) ChangeEmail(_ context.Context, id uuid.UUID, email, _ string) (*identity.User, error) {
	return &identity.User{ID: id, Email: email, Role: identity.RolePublic, HasPassword: true}, nil
}
func (stubAuth) ChangePassword(_ context.Context, _ uuid.UUID, _, _ string) error { return nil }
func (stubAuth) GetNotificationPreferences(_ context.Context, id uuid.UUID) (*identity.NotificationPreferences, error) {
	prefs := identity.DefaultNotificationPreferences(id)
	return &prefs, nil
}
func (stubAuth) UpdateNotificationPreferences(_ context.Context, id uuid.UUID, prefs identity.NotificationPreferences) (*identity.NotificationPreferences, error) {
	prefs.UserID = id
	return &prefs, nil
}
func (stubAuth) ListUsers(context.Context, *identity.Role, int, int) ([]identity.User, int, error) {
	return nil, 0, nil
}
func (stubAuth) UpdateUserRole(context.Context, uuid.UUID, uuid.UUID, identity.Role) (*identity.User, error) {
	return nil, nil
}

type stubSettings struct{}

func (stubSettings) GetScrapeSettings(_ context.Context) (*settings.ScrapeSettingsView, error) {
	return &settings.ScrapeSettingsView{ScrapeSources: []string{}, TelegramChannels: []string{}, TelegramKeywords: []string{}}, nil
}
func (stubSettings) UpdateScrapeSettings(_ context.Context, _ uuid.UUID, _ settings.ScrapeSettingsUpdate) (*settings.ScrapeSettingsView, error) {
	return &settings.ScrapeSettingsView{ScrapeSources: []string{}, TelegramChannels: []string{}, TelegramKeywords: []string{}}, nil
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
		Search:      handler.NewSearchHandler(stubSearch{}),
		Onboarding:  handler.NewOnboardingHandler(stubOnboarding{}),
		Auth:        handler.NewAuthHandler(stubAuth{}, cfg.AuthCookieName),
		Settings:    handler.NewSettingsHandler(stubSettings{}),
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

func TestRouter_listArticlesAdmin(t *testing.T) {
	w := serveRouter(t, http.MethodGet, "/admin/v1/articles", nil, map[string]string{"X-User-ID": uuid.New().String()})
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_createArticleAdmin(t *testing.T) {
	w := serveRouter(t, http.MethodPost, "/admin/v1/articles", strings.NewReader(`{"title":"Guide","body":"Body"}`),
		map[string]string{"X-User-ID": uuid.New().String()})
	assist.Equal(t, http.StatusCreated, w.Code)
}

func TestRouter_updateArticleAdmin(t *testing.T) {
	w := serveRouter(t, http.MethodPut, "/admin/v1/articles/"+uuid.New().String(),
		strings.NewReader(`{"title":"Updated","body":"Body"}`),
		map[string]string{"X-User-ID": uuid.New().String()})
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_patchArticleAdmin(t *testing.T) {
	w := serveRouter(t, http.MethodPatch, "/admin/v1/articles/"+uuid.New().String(),
		strings.NewReader(`{"status":"published","verified":true}`),
		map[string]string{"X-User-ID": uuid.New().String()})
	assist.Equal(t, http.StatusOK, w.Code)
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

func TestRouter_listPendingEvents(t *testing.T) {
	w := serveRouter(t, http.MethodGet, "/admin/v1/events", nil, map[string]string{"X-User-ID": uuid.New().String()})
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_getEventByID(t *testing.T) {
	w := serveRouter(t, http.MethodGet, "/admin/v1/events/"+uuid.New().String(), nil, map[string]string{"X-User-ID": uuid.New().String()})
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_reviewEvent(t *testing.T) {
	w := serveRouter(t, http.MethodPatch, "/admin/v1/events/"+uuid.New().String(),
		strings.NewReader(`{"status":"approved"}`),
		map[string]string{"X-User-ID": uuid.New().String()})
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_syncEvents(t *testing.T) {
	w := serveRouter(t, http.MethodPost, "/admin/v1/events/sync", nil, map[string]string{"X-User-ID": uuid.New().String()})
	assist.Equal(t, http.StatusOK, w.Code)
	var resp dto.SyncEventsResponse
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
}

func TestRouter_register(t *testing.T) {
	w := serveRouter(t, http.MethodPost, "/api/v1/auth/register",
		strings.NewReader(`{"email":"new@example.com","password":"password1"}`), nil)
	assist.Equal(t, http.StatusCreated, w.Code)
	assist.Contains(t, w.Header().Get("Set-Cookie"), "mq_access_token=")
}

func TestRouter_login(t *testing.T) {
	w := serveRouter(t, http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"email":"user@example.com","password":"password1"}`), nil)
	assist.Equal(t, http.StatusOK, w.Code)
	assist.Contains(t, w.Header().Get("Set-Cookie"), "mq_access_token=")
}
