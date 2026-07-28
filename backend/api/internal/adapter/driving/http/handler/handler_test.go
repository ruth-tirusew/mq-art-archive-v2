package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	"github.com/mq/api/internal/domain/apperrors"
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

type mockContent struct {
	listPublished        func(ctx context.Context, filter content.ListFilter) ([]content.Article, error)
	getBySlug            func(ctx context.Context, slug string) (*content.Article, error)
	createDraft          func(ctx context.Context, authorID uuid.UUID, title, body string) (*content.Article, error)
	adminList            func(ctx context.Context, status *content.ArticleStatus, limit, offset int) ([]content.Article, error)
	adminGet             func(ctx context.Context, id uuid.UUID) (*content.Article, error)
	adminCreate          func(ctx context.Context, authorID uuid.UUID, write content.ArticleWrite) (*content.Article, error)
	adminUpdate          func(ctx context.Context, id, editorID uuid.UUID, write content.ArticleWrite) (*content.Article, error)
	adminSetStatus       func(ctx context.Context, id uuid.UUID, status *content.ArticleStatus, verified *bool) (*content.Article, error)
	adminListRevisions   func(ctx context.Context, articleID uuid.UUID, limit, offset int) ([]content.ArticleRevision, error)
	adminGetRevision     func(ctx context.Context, articleID uuid.UUID, version int) (*content.ArticleRevision, error)
	adminRestoreRevision func(ctx context.Context, articleID uuid.UUID, version int, editorID uuid.UUID) (*content.Article, error)
}

func (m *mockContent) ListPublished(ctx context.Context, filter content.ListFilter) ([]content.Article, error) {
	return m.listPublished(ctx, filter)
}
func (m *mockContent) GetBySlug(ctx context.Context, slug string) (*content.Article, error) {
	return m.getBySlug(ctx, slug)
}
func (m *mockContent) CreateDraft(ctx context.Context, authorID uuid.UUID, title, body string) (*content.Article, error) {
	return m.createDraft(ctx, authorID, title, body)
}
func (m *mockContent) AdminList(ctx context.Context, status *content.ArticleStatus, limit, offset int) ([]content.Article, error) {
	if m.adminList != nil {
		return m.adminList(ctx, status, limit, offset)
	}
	return nil, nil
}
func (m *mockContent) AdminGet(ctx context.Context, id uuid.UUID) (*content.Article, error) {
	if m.adminGet != nil {
		return m.adminGet(ctx, id)
	}
	return nil, apperrors.ErrNotFound
}
func (m *mockContent) AdminCreate(ctx context.Context, authorID uuid.UUID, write content.ArticleWrite) (*content.Article, error) {
	if m.adminCreate != nil {
		return m.adminCreate(ctx, authorID, write)
	}
	return nil, nil
}
func (m *mockContent) AdminUpdate(ctx context.Context, id, editorID uuid.UUID, write content.ArticleWrite) (*content.Article, error) {
	if m.adminUpdate != nil {
		return m.adminUpdate(ctx, id, editorID, write)
	}
	return nil, apperrors.ErrNotFound
}
func (m *mockContent) AdminSetStatus(ctx context.Context, id uuid.UUID, status *content.ArticleStatus, verified *bool) (*content.Article, error) {
	if m.adminSetStatus != nil {
		return m.adminSetStatus(ctx, id, status, verified)
	}
	return nil, apperrors.ErrNotFound
}
func (m *mockContent) AdminListRevisions(ctx context.Context, articleID uuid.UUID, limit, offset int) ([]content.ArticleRevision, error) {
	if m.adminListRevisions != nil {
		return m.adminListRevisions(ctx, articleID, limit, offset)
	}
	return nil, nil
}
func (m *mockContent) AdminGetRevision(ctx context.Context, articleID uuid.UUID, version int) (*content.ArticleRevision, error) {
	if m.adminGetRevision != nil {
		return m.adminGetRevision(ctx, articleID, version)
	}
	return nil, apperrors.ErrNotFound
}
func (m *mockContent) AdminRestoreRevision(ctx context.Context, articleID uuid.UUID, version int, editorID uuid.UUID) (*content.Article, error) {
	if m.adminRestoreRevision != nil {
		return m.adminRestoreRevision(ctx, articleID, version, editorID)
	}
	return nil, apperrors.ErrNotFound
}

type mockProfile struct {
	getArtistBySlug   func(ctx context.Context, slug string) (*profile.ArtistProfile, error)
	getArtistByHandle func(ctx context.Context, handle string) (*profile.ArtistProfile, error)
	listApproved      func(ctx context.Context, filter profile.ListFilter) ([]profile.ArtistProfile, error)
	getArtistByID     func(ctx context.Context, id uuid.UUID) (*profile.ArtistProfile, error)
	getArtistByUserID func(ctx context.Context, userID uuid.UUID) (*profile.ArtistProfile, error)
	updateArtist      func(ctx context.Context, p profile.ArtistProfile) (*profile.ArtistProfile, error)
	updateOwnProfile  func(ctx context.Context, userID uuid.UUID, update profile.OwnProfileUpdate) (*profile.ArtistProfile, error)
}

func (m *mockProfile) GetArtistBySlug(ctx context.Context, slug string) (*profile.ArtistProfile, error) {
	return m.getArtistBySlug(ctx, slug)
}
func (m *mockProfile) GetArtistByHandle(ctx context.Context, handle string) (*profile.ArtistProfile, error) {
	if m.getArtistByHandle != nil {
		return m.getArtistByHandle(ctx, handle)
	}
	return nil, apperrors.ErrNotFound
}
func (m *mockProfile) ListApproved(ctx context.Context, filter profile.ListFilter) ([]profile.ArtistProfile, error) {
	if m.listApproved != nil {
		return m.listApproved(ctx, filter)
	}
	return nil, nil
}
func (m *mockProfile) ListAll(ctx context.Context, status *profile.ProfileStatus, limit, offset int) ([]profile.ArtistProfile, error) {
	return nil, nil
}
func (m *mockProfile) GetArtistByID(ctx context.Context, id uuid.UUID) (*profile.ArtistProfile, error) {
	return m.getArtistByID(ctx, id)
}
func (m *mockProfile) GetArtistByUserID(ctx context.Context, userID uuid.UUID) (*profile.ArtistProfile, error) {
	return m.getArtistByUserID(ctx, userID)
}
func (m *mockProfile) UpdateArtist(ctx context.Context, p profile.ArtistProfile) (*profile.ArtistProfile, error) {
	return m.updateArtist(ctx, p)
}
func (m *mockProfile) UpdateOwnProfile(ctx context.Context, userID uuid.UUID, update profile.OwnProfileUpdate) (*profile.ArtistProfile, error) {
	if m.updateOwnProfile != nil {
		return m.updateOwnProfile(ctx, userID, update)
	}
	return nil, apperrors.ErrNotImplemented
}
func (m *mockProfile) AdminCreate(ctx context.Context, write profile.AdminArtistWrite) (*profile.ArtistProfile, error) {
	return &profile.ArtistProfile{ID: uuid.New(), DisplayName: write.DisplayName}, nil
}
func (m *mockProfile) AdminUpdateContent(ctx context.Context, id uuid.UUID, write profile.AdminArtistWrite) (*profile.ArtistProfile, error) {
	return &profile.ArtistProfile{ID: id, DisplayName: write.DisplayName}, nil
}
func (m *mockProfile) AdminDelete(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockProfile) AdminUpdate(ctx context.Context, id uuid.UUID, status *profile.ProfileStatus, featured *bool) (*profile.ArtistProfile, error) {
	return &profile.ArtistProfile{ID: id}, nil
}

type mockInstitution struct {
	getBySlug func(ctx context.Context, slug string) (*institution.Institution, error)
	getByID   func(ctx context.Context, id uuid.UUID) (*institution.Institution, error)
	update    func(ctx context.Context, inst institution.Institution) (*institution.Institution, error)
}

func (m *mockInstitution) GetBySlug(ctx context.Context, slug string) (*institution.Institution, error) {
	return m.getBySlug(ctx, slug)
}
func (m *mockInstitution) GetByID(ctx context.Context, id uuid.UUID) (*institution.Institution, error) {
	return m.getByID(ctx, id)
}
func (m *mockInstitution) Update(ctx context.Context, inst institution.Institution) (*institution.Institution, error) {
	return m.update(ctx, inst)
}

type mockArt struct {
	listByArtist       func(ctx context.Context, artistID uuid.UUID) ([]art.ArtPost, error)
	listOwnedByArtist  func(ctx context.Context, artistID uuid.UUID) ([]art.ArtPost, error)
	listPublished      func(ctx context.Context, filter art.ListFilter) ([]art.ArtPostWithArtist, error)
	listAll            func(ctx context.Context, status *art.ArtStatus, limit, offset int) ([]art.ArtPostWithArtist, error)
	getByID            func(ctx context.Context, id uuid.UUID) (*art.ArtPost, error)
	getOwned           func(ctx context.Context, artistID, postID uuid.UUID) (*art.ArtPost, error)
	createDraft        func(ctx context.Context, artistID uuid.UUID, write art.ArtPostWrite) (*art.ArtPost, error)
	updateOwned        func(ctx context.Context, artistID, postID uuid.UUID, write art.ArtPostWrite) (*art.ArtPost, error)
	publishOwned       func(ctx context.Context, artistID, postID uuid.UUID) (*art.ArtPost, error)
	unpublishOwned     func(ctx context.Context, artistID, postID uuid.UUID) (*art.ArtPost, error)
	archiveOwned       func(ctx context.Context, artistID, postID uuid.UUID) (*art.ArtPost, error)
	deleteOwned        func(ctx context.Context, artistID, postID uuid.UUID) error
	adminCreate        func(ctx context.Context, artistID uuid.UUID, write art.ArtPostWrite, status *art.ArtStatus) (*art.ArtPost, error)
	adminUpdateContent func(ctx context.Context, postID uuid.UUID, write art.ArtPostWrite) (*art.ArtPost, error)
	adminDelete        func(ctx context.Context, postID uuid.UUID) error
	adminUpdate        func(ctx context.Context, postID uuid.UUID, status *art.ArtStatus, featured *bool) (*art.ArtPost, error)
}

func (m *mockArt) ListByArtist(ctx context.Context, artistID uuid.UUID) ([]art.ArtPost, error) {
	if m.listByArtist != nil {
		return m.listByArtist(ctx, artistID)
	}
	return nil, nil
}
func (m *mockArt) ListOwnedByArtist(ctx context.Context, artistID uuid.UUID) ([]art.ArtPost, error) {
	if m.listOwnedByArtist != nil {
		return m.listOwnedByArtist(ctx, artistID)
	}
	return nil, nil
}
func (m *mockArt) ListPublished(ctx context.Context, filter art.ListFilter) ([]art.ArtPostWithArtist, error) {
	if m.listPublished != nil {
		return m.listPublished(ctx, filter)
	}
	return nil, nil
}
func (m *mockArt) ListAll(ctx context.Context, status *art.ArtStatus, limit, offset int) ([]art.ArtPostWithArtist, error) {
	if m.listAll != nil {
		return m.listAll(ctx, status, limit, offset)
	}
	return nil, nil
}
func (m *mockArt) GetByID(ctx context.Context, id uuid.UUID) (*art.ArtPost, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return nil, nil
}
func (m *mockArt) GetOwned(ctx context.Context, artistID, postID uuid.UUID) (*art.ArtPost, error) {
	if m.getOwned != nil {
		return m.getOwned(ctx, artistID, postID)
	}
	return nil, nil
}
func (m *mockArt) CreateDraft(ctx context.Context, artistID uuid.UUID, write art.ArtPostWrite) (*art.ArtPost, error) {
	if m.createDraft != nil {
		return m.createDraft(ctx, artistID, write)
	}
	return nil, nil
}
func (m *mockArt) UpdateOwned(ctx context.Context, artistID, postID uuid.UUID, write art.ArtPostWrite) (*art.ArtPost, error) {
	if m.updateOwned != nil {
		return m.updateOwned(ctx, artistID, postID, write)
	}
	return nil, nil
}
func (m *mockArt) PublishOwned(ctx context.Context, artistID, postID uuid.UUID) (*art.ArtPost, error) {
	if m.publishOwned != nil {
		return m.publishOwned(ctx, artistID, postID)
	}
	return nil, nil
}
func (m *mockArt) UnpublishOwned(ctx context.Context, artistID, postID uuid.UUID) (*art.ArtPost, error) {
	if m.unpublishOwned != nil {
		return m.unpublishOwned(ctx, artistID, postID)
	}
	return nil, nil
}
func (m *mockArt) ArchiveOwned(ctx context.Context, artistID, postID uuid.UUID) (*art.ArtPost, error) {
	if m.archiveOwned != nil {
		return m.archiveOwned(ctx, artistID, postID)
	}
	return nil, nil
}
func (m *mockArt) DeleteOwned(ctx context.Context, artistID, postID uuid.UUID) error {
	if m.deleteOwned != nil {
		return m.deleteOwned(ctx, artistID, postID)
	}
	return nil
}
func (m *mockArt) AdminCreate(ctx context.Context, artistID uuid.UUID, write art.ArtPostWrite, status *art.ArtStatus) (*art.ArtPost, error) {
	if m.adminCreate != nil {
		return m.adminCreate(ctx, artistID, write, status)
	}
	return nil, nil
}
func (m *mockArt) AdminUpdateContent(ctx context.Context, postID uuid.UUID, write art.ArtPostWrite) (*art.ArtPost, error) {
	if m.adminUpdateContent != nil {
		return m.adminUpdateContent(ctx, postID, write)
	}
	return nil, nil
}
func (m *mockArt) AdminDelete(ctx context.Context, postID uuid.UUID) error {
	if m.adminDelete != nil {
		return m.adminDelete(ctx, postID)
	}
	return nil
}
func (m *mockArt) AdminUpdate(ctx context.Context, postID uuid.UUID, status *art.ArtStatus, featured *bool) (*art.ArtPost, error) {
	if m.adminUpdate != nil {
		return m.adminUpdate(ctx, postID, status, featured)
	}
	return nil, nil
}

type mockOnboarding struct {
	listPending      func(ctx context.Context) ([]onboarding.OnboardingApplication, error)
	getByID          func(ctx context.Context, id uuid.UUID) (*onboarding.OnboardingApplication, error)
	getMyApplication func(ctx context.Context, applicantID uuid.UUID) (*onboarding.OnboardingApplication, error)
	submit           func(ctx context.Context, applicantID uuid.UUID, applicantType onboarding.ApplicantType, displayName, notes string) (*onboarding.OnboardingApplication, error)
	review           func(ctx context.Context, id, reviewerID uuid.UUID, status onboarding.ApprovalStatus, notes string) (*onboarding.OnboardingApplication, error)
}

func (m *mockOnboarding) ListPending(ctx context.Context) ([]onboarding.OnboardingApplication, error) {
	return m.listPending(ctx)
}
func (m *mockOnboarding) GetByID(ctx context.Context, id uuid.UUID) (*onboarding.OnboardingApplication, error) {
	return m.getByID(ctx, id)
}
func (m *mockOnboarding) GetMyApplication(ctx context.Context, applicantID uuid.UUID) (*onboarding.OnboardingApplication, error) {
	if m.getMyApplication != nil {
		return m.getMyApplication(ctx, applicantID)
	}
	return nil, apperrors.ErrNotFound
}
func (m *mockOnboarding) Submit(ctx context.Context, applicantID uuid.UUID, applicantType onboarding.ApplicantType, displayName, notes string) (*onboarding.OnboardingApplication, error) {
	if m.submit != nil {
		return m.submit(ctx, applicantID, applicantType, displayName, notes)
	}
	return nil, apperrors.ErrNotImplemented
}
func (m *mockOnboarding) SubmitWithHandle(ctx context.Context, applicantID uuid.UUID, applicantType onboarding.ApplicantType, displayName, handle, notes string) (*onboarding.OnboardingApplication, error) {
	return m.Submit(ctx, applicantID, applicantType, displayName, notes)
}
func (m *mockOnboarding) CheckHandleAvailable(context.Context, string) (bool, error) {
	return true, nil
}
func (m *mockOnboarding) Review(ctx context.Context, id, reviewerID uuid.UUID, status onboarding.ApprovalStatus, notes string) (*onboarding.OnboardingApplication, error) {
	return m.review(ctx, id, reviewerID, status, notes)
}

type mockEvent struct {
	list        func(ctx context.Context, filter events.ListFilter) ([]events.Event, error)
	listPending func(ctx context.Context) ([]events.Event, error)
	getByID     func(ctx context.Context, id uuid.UUID) (*events.Event, error)
	getBySlug   func(ctx context.Context, slug string) (*events.Event, error)
	review      func(ctx context.Context, id, reviewerID uuid.UUID, status events.EventStatus, notes string) (*events.Event, error)
	sync        func(ctx context.Context) (int, error)
}

func (m *mockEvent) List(ctx context.Context, filter events.ListFilter) ([]events.Event, error) {
	if m.list != nil {
		return m.list(ctx, filter)
	}
	return nil, nil
}
func (m *mockEvent) ListPending(ctx context.Context) ([]events.Event, error) {
	return m.listPending(ctx)
}
func (m *mockEvent) Search(ctx context.Context, query string, limit int) ([]events.Event, error) {
	return nil, nil
}
func (m *mockEvent) GetByID(ctx context.Context, id uuid.UUID) (*events.Event, error) {
	return m.getByID(ctx, id)
}
func (m *mockEvent) GetBySlug(ctx context.Context, slug string) (*events.Event, error) {
	if m.getBySlug != nil {
		return m.getBySlug(ctx, slug)
	}
	return nil, apperrors.ErrNotFound
}
func (m *mockEvent) AdminCreate(ctx context.Context, write events.EventWrite) (*events.Event, error) {
	return &events.Event{ID: uuid.New(), Title: write.Title, StartsAt: write.StartsAt, Status: events.EventStatusPending}, nil
}
func (m *mockEvent) Submit(ctx context.Context, write events.EventWrite) (*events.Event, error) {
	return &events.Event{ID: uuid.New(), Title: write.Title, StartsAt: write.StartsAt, Status: events.EventStatusPending}, nil
}
func (m *mockEvent) AdminUpdateContent(ctx context.Context, id uuid.UUID, write events.EventWrite) (*events.Event, error) {
	return &events.Event{ID: id, Title: write.Title, StartsAt: write.StartsAt}, nil
}
func (m *mockEvent) AdminDelete(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockEvent) Review(ctx context.Context, id, reviewerID uuid.UUID, status events.EventStatus, notes string) (*events.Event, error) {
	return m.review(ctx, id, reviewerID, status, notes)
}
func (m *mockEvent) Sync(ctx context.Context) (int, error) {
	return m.sync(ctx)
}

func serve(t *testing.T, method, path string, body io.Reader, params gin.Params, headers map[string]string, fn func(*gin.Context)) *httptest.ResponseRecorder {
	t.Helper()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, body)
	if body != nil {
		c.Request.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		c.Request.Header.Set(k, v)
	}
	if v := headers["X-User-ID"]; v != "" {
		if id, err := uuid.Parse(v); err == nil {
			c.Set(requestauth.ContextUserID, id)
			c.Set(requestauth.ContextUserRole, identity.RoleArtist)
		}
	}
	c.Params = params
	fn(c)
	return w
}

func decodeError(t *testing.T, w *httptest.ResponseRecorder) dto.ErrorResponse {
	t.Helper()
	var resp dto.ErrorResponse
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	return resp
}

func TestHealthHandler(t *testing.T) {
	w := serve(t, http.MethodGet, "/health", nil, nil, nil, NewHealthHandler().Health)

	assist.Equal(t, http.StatusOK, w.Code)
	var resp dto.HealthResponse
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assist.Equal(t, "ok", resp.Status)
}

func TestArticleHandler_List(t *testing.T) {
	now := time.Now().UTC()
	id := uuid.New()
	h := NewArticleHandler(&mockContent{
		listPublished: func(ctx context.Context, filter content.ListFilter) ([]content.Article, error) {
			return []content.Article{{
				ID: id, Slug: "a", Title: "A", Status: content.ArticleStatusPublished,
				CreatedAt: now, UpdatedAt: now,
			}}, nil
		},
	})

	w := serve(t, http.MethodGet, "/api/v1/articles", nil, nil, nil, h.List)

	assist.Equal(t, http.StatusOK, w.Code)
	var resp []dto.ArticleResponse
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assist.Len(t, 1, len(resp))
	assist.Equal(t, "a", resp[0].Slug)
}

func TestArticleHandler_List_error(t *testing.T) {
	h := NewArticleHandler(&mockContent{
		listPublished: func(ctx context.Context, filter content.ListFilter) ([]content.Article, error) {
			return nil, errors.New("db down")
		},
	})

	w := serve(t, http.MethodGet, "/api/v1/articles", nil, nil, nil, h.List)

	assist.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestArticleHandler_GetBySlug_success(t *testing.T) {
	now := time.Now().UTC()
	articleID := uuid.New()
	h := NewArticleHandler(&mockContent{
		getBySlug: func(ctx context.Context, slug string) (*content.Article, error) {
			assist.Equal(t, "welcome", slug)
			return &content.Article{
				ID: articleID, Slug: slug, Title: "Welcome", Status: content.ArticleStatusPublished,
				CreatedAt: now, UpdatedAt: now,
			}, nil
		},
	})

	w := serve(t, http.MethodGet, "/api/v1/articles/welcome", nil, gin.Params{{Key: "slug", Value: "welcome"}}, nil, h.GetBySlug)

	assist.Equal(t, http.StatusOK, w.Code)
	var resp dto.ArticleResponse
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assist.Equal(t, "welcome", resp.Slug)
}

func TestArticleHandler_GetBySlug_notFound(t *testing.T) {
	h := NewArticleHandler(&mockContent{
		getBySlug: func(ctx context.Context, slug string) (*content.Article, error) {
			return nil, apperrors.ErrNotFound
		},
	})

	w := serve(t, http.MethodGet, "/api/v1/articles/missing", nil, gin.Params{{Key: "slug", Value: "missing"}}, nil, h.GetBySlug)

	assist.Equal(t, http.StatusNotFound, w.Code)
}

func TestArticleHandler_Create_badRequest(t *testing.T) {
	h := NewArticleHandler(&mockContent{})
	w := serve(t, http.MethodPost, "/api/v1/articles", strings.NewReader(`{}`), nil, nil, h.Create)

	assist.Equal(t, http.StatusBadRequest, w.Code)
}

func TestArticleHandler_Create_unauthorizedWithoutHeader(t *testing.T) {
	h := NewArticleHandler(&mockContent{})
	w := serve(t, http.MethodPost, "/api/v1/articles", strings.NewReader(`{"title":"T"}`), nil, nil, h.Create)

	assist.Equal(t, http.StatusUnauthorized, w.Code)
	assist.Equal(t, "unauthorized", decodeError(t, w).Error)
}

func TestArticleHandler_Create_success(t *testing.T) {
	authorID := uuid.New()
	articleID := uuid.New()
	h := NewArticleHandler(&mockContent{
		createDraft: func(ctx context.Context, id uuid.UUID, title, body string) (*content.Article, error) {
			assist.Equal(t, authorID, id)
			return &content.Article{ID: articleID, Title: title, Slug: "t", Status: content.ArticleStatusDraft}, nil
		},
	})

	w := serve(t, http.MethodPost, "/api/v1/articles", strings.NewReader(`{"title":"T","body":"B"}`), nil,
		map[string]string{"X-User-ID": authorID.String()}, h.Create)

	assist.Equal(t, http.StatusCreated, w.Code)
}

func TestArticleHandler_ListArticlesAdmin(t *testing.T) {
	h := NewArticleHandler(&mockContent{
		adminList: func(ctx context.Context, status *content.ArticleStatus, limit, offset int) ([]content.Article, error) {
			assist.Equal(t, content.ArticleStatusDraft, *status)
			return []content.Article{{ID: uuid.New(), Title: "Draft", Status: content.ArticleStatusDraft}}, nil
		},
	})

	w := serve(t, http.MethodGet, "/admin/v1/articles?status=draft", nil, nil, nil, h.ListArticlesAdmin)
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestArticleHandler_CreateArticleAdmin(t *testing.T) {
	authorID := uuid.New()
	h := NewArticleHandler(&mockContent{
		adminCreate: func(ctx context.Context, id uuid.UUID, write content.ArticleWrite) (*content.Article, error) {
			assist.Equal(t, authorID, id)
			assist.Equal(t, "Legal", write.Category)
			return &content.Article{ID: uuid.New(), Title: write.Title, Status: content.ArticleStatusDraft}, nil
		},
	})

	w := serve(t, http.MethodPost, "/admin/v1/articles",
		strings.NewReader(`{"title":"Guide","body":"Body","category":"Legal"}`),
		nil, map[string]string{"X-User-ID": authorID.String()}, h.CreateArticleAdmin)
	assist.Equal(t, http.StatusCreated, w.Code)
}

func TestArticleHandler_UpdateArticleAdmin(t *testing.T) {
	id := uuid.New()
	editorID := uuid.New()
	h := NewArticleHandler(&mockContent{
		adminUpdate: func(ctx context.Context, gotID, gotEditor uuid.UUID, write content.ArticleWrite) (*content.Article, error) {
			assist.Equal(t, id, gotID)
			assist.Equal(t, editorID, gotEditor)
			assist.Equal(t, "Updated", write.Title)
			return &content.Article{ID: id, Title: write.Title, Status: content.ArticleStatusDraft, Version: 2}, nil
		},
	})

	w := serve(t, http.MethodPut, "/admin/v1/articles/"+id.String(),
		strings.NewReader(`{"title":"Updated","body":"New body"}`),
		gin.Params{{Key: "id", Value: id.String()}},
		map[string]string{"X-User-ID": editorID.String()},
		h.UpdateArticleAdmin)
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestArticleHandler_PatchArticleAdmin(t *testing.T) {
	id := uuid.New()
	h := NewArticleHandler(&mockContent{
		adminSetStatus: func(ctx context.Context, gotID uuid.UUID, status *content.ArticleStatus, verified *bool) (*content.Article, error) {
			assist.Equal(t, id, gotID)
			assist.Equal(t, content.ArticleStatusPublished, *status)
			assist.Equal(t, true, *verified)
			return &content.Article{ID: id, Status: *status, Verified: *verified}, nil
		},
	})

	w := serve(t, http.MethodPatch, "/admin/v1/articles/"+id.String(),
		strings.NewReader(`{"status":"published","verified":true}`),
		gin.Params{{Key: "id", Value: id.String()}}, nil, h.PatchArticleAdmin)
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestArticleHandler_GetArticleAdmin_notFound(t *testing.T) {
	h := NewArticleHandler(&mockContent{})
	id := uuid.New()
	w := serve(t, http.MethodGet, "/admin/v1/articles/"+id.String(), nil,
		gin.Params{{Key: "id", Value: id.String()}}, nil, h.GetArticleAdmin)
	assist.Equal(t, http.StatusNotFound, w.Code)
}

func TestProfileHandler_GetBySlug_success(t *testing.T) {
	artistID := uuid.New()
	h := NewProfileHandler(&mockProfile{
		getArtistBySlug: func(ctx context.Context, slug string) (*profile.ArtistProfile, error) {
			assist.Equal(t, "abebe", slug)
			return &profile.ArtistProfile{
				ID: artistID, Slug: slug, DisplayName: "Abebe", Status: profile.ProfileStatusApproved,
			}, nil
		},
	})

	w := serve(t, http.MethodGet, "/api/v1/artists/abebe", nil, gin.Params{{Key: "slug", Value: "abebe"}}, nil, h.GetBySlug)

	assist.Equal(t, http.StatusOK, w.Code)
	var resp profile.ArtistProfile
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assist.Equal(t, "abebe", resp.Slug)
}

func TestProfileHandler_GetBySlug_notFound(t *testing.T) {
	h := NewProfileHandler(&mockProfile{
		getArtistBySlug: func(ctx context.Context, slug string) (*profile.ArtistProfile, error) {
			return nil, apperrors.ErrNotFound
		},
	})

	w := serve(t, http.MethodGet, "/api/v1/artists/missing", nil, gin.Params{{Key: "slug", Value: "missing"}}, nil, h.GetBySlug)

	assist.Equal(t, http.StatusNotFound, w.Code)
}

func TestInstitutionHandler_GetBySlug_success(t *testing.T) {
	instID := uuid.New()
	h := NewInstitutionHandler(&mockInstitution{
		getBySlug: func(ctx context.Context, slug string) (*institution.Institution, error) {
			assist.Equal(t, "national-gallery", slug)
			return &institution.Institution{
				ID: instID, Slug: slug, Name: "National Gallery", Status: institution.InstitutionStatusApproved,
			}, nil
		},
	})

	w := serve(t, http.MethodGet, "/api/v1/institutions/national-gallery", nil,
		gin.Params{{Key: "slug", Value: "national-gallery"}}, nil, h.GetBySlug)

	assist.Equal(t, http.StatusOK, w.Code)
	var resp institution.Institution
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assist.Equal(t, "national-gallery", resp.Slug)
}

func TestInstitutionHandler_GetBySlug_notFound(t *testing.T) {
	h := NewInstitutionHandler(&mockInstitution{
		getBySlug: func(ctx context.Context, slug string) (*institution.Institution, error) {
			return nil, apperrors.ErrNotFound
		},
	})

	w := serve(t, http.MethodGet, "/api/v1/institutions/missing", nil, gin.Params{{Key: "slug", Value: "missing"}}, nil, h.GetBySlug)

	assist.Equal(t, http.StatusNotFound, w.Code)
}

func TestArtHandler_ListByArtistSlug_success(t *testing.T) {
	artistID := uuid.New()
	postID := uuid.New()
	h := NewArtHandler(
		&mockArt{
			listByArtist: func(ctx context.Context, id uuid.UUID) ([]art.ArtPost, error) {
				assist.Equal(t, artistID, id)
				return []art.ArtPost{{ID: postID, ArtistID: artistID, Title: "Sunset", Status: art.ArtStatusPublished}}, nil
			},
		},
		&mockProfile{
			getArtistBySlug: func(ctx context.Context, slug string) (*profile.ArtistProfile, error) {
				return &profile.ArtistProfile{ID: artistID, Slug: slug}, nil
			},
		},
	)

	w := serve(t, http.MethodGet, "/api/v1/artists/abebe/posts", nil, gin.Params{{Key: "slug", Value: "abebe"}}, nil, h.ListByArtistSlug)

	assist.Equal(t, http.StatusOK, w.Code)
	var resp []art.ArtPost
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assist.Len(t, 1, len(resp))
	assist.Equal(t, "Sunset", resp[0].Title)
}

func TestArtHandler_ListByArtistSlug_artistNotFound(t *testing.T) {
	h := NewArtHandler(&mockArt{}, &mockProfile{
		getArtistBySlug: func(ctx context.Context, slug string) (*profile.ArtistProfile, error) {
			return nil, apperrors.ErrNotFound
		},
	})

	w := serve(t, http.MethodGet, "/api/v1/artists/missing/posts", nil, gin.Params{{Key: "slug", Value: "missing"}}, nil, h.ListByArtistSlug)

	assist.Equal(t, http.StatusNotFound, w.Code)
}

func TestArtHandler_ListByArtistSlug_listError(t *testing.T) {
	artistID := uuid.New()
	h := NewArtHandler(
		&mockArt{
			listByArtist: func(ctx context.Context, id uuid.UUID) ([]art.ArtPost, error) {
				return nil, errors.New("db down")
			},
		},
		&mockProfile{
			getArtistBySlug: func(ctx context.Context, slug string) (*profile.ArtistProfile, error) {
				return &profile.ArtistProfile{ID: artistID, Slug: slug}, nil
			},
		},
	)

	w := serve(t, http.MethodGet, "/api/v1/artists/abebe/posts", nil, gin.Params{{Key: "slug", Value: "abebe"}}, nil, h.ListByArtistSlug)

	assist.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestArtHandler_GetByID_success(t *testing.T) {
	postID := uuid.New()
	artistID := uuid.New()
	h := NewArtHandler(&mockArt{
		getByID: func(ctx context.Context, id uuid.UUID) (*art.ArtPost, error) {
			assist.Equal(t, postID, id)
			return &art.ArtPost{ID: postID, ArtistID: artistID, Title: "Sunset"}, nil
		},
	}, &mockProfile{})

	w := serve(t, http.MethodGet, "/api/v1/posts/"+postID.String(), nil,
		gin.Params{{Key: "id", Value: postID.String()}}, nil, h.GetByID)

	assist.Equal(t, http.StatusOK, w.Code)
	var resp art.ArtPost
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assist.Equal(t, "Sunset", resp.Title)
}

func TestArtHandler_GetByID_invalidID(t *testing.T) {
	h := NewArtHandler(&mockArt{}, &mockProfile{})
	w := serve(t, http.MethodGet, "/api/v1/posts/not-a-uuid", nil,
		gin.Params{{Key: "id", Value: "not-a-uuid"}}, nil, h.GetByID)

	assist.Equal(t, http.StatusBadRequest, w.Code)
}

func TestArtHandler_GetByID_notFound(t *testing.T) {
	postID := uuid.New()
	h := NewArtHandler(&mockArt{
		getByID: func(ctx context.Context, id uuid.UUID) (*art.ArtPost, error) {
			return nil, apperrors.ErrNotFound
		},
	}, &mockProfile{})

	w := serve(t, http.MethodGet, "/api/v1/posts/"+postID.String(), nil,
		gin.Params{{Key: "id", Value: postID.String()}}, nil, h.GetByID)

	assist.Equal(t, http.StatusNotFound, w.Code)
}

func TestArtHandler_Create_success(t *testing.T) {
	userID := uuid.New()
	artistID := uuid.New()
	postID := uuid.New()
	h := NewArtHandler(
		&mockArt{
			createDraft: func(ctx context.Context, aid uuid.UUID, write art.ArtPostWrite) (*art.ArtPost, error) {
				assist.Equal(t, artistID, aid)
				assist.Equal(t, "New Work", write.Title)
				return &art.ArtPost{ID: postID, ArtistID: artistID, Title: write.Title, Status: art.ArtStatusDraft}, nil
			},
		},
		&mockProfile{
			getArtistByUserID: func(ctx context.Context, uid uuid.UUID) (*profile.ArtistProfile, error) {
				assist.Equal(t, userID, uid)
				return &profile.ArtistProfile{ID: artistID, UserID: userID}, nil
			},
		},
	)

	w := serve(t, http.MethodPost, "/api/v1/posts", strings.NewReader(`{"title":"New Work","medium":"oil"}`),
		nil, map[string]string{"X-User-ID": userID.String()}, h.Create)

	assist.Equal(t, http.StatusCreated, w.Code)
	var resp art.ArtPost
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assist.Equal(t, postID, resp.ID)
}

func TestArtHandler_Create_missingUserHeader(t *testing.T) {
	h := NewArtHandler(&mockArt{}, &mockProfile{})
	w := serve(t, http.MethodPost, "/api/v1/posts", strings.NewReader(`{"title":"T"}`), nil, nil, h.Create)

	assist.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestArtHandler_Create_noArtistProfile(t *testing.T) {
	userID := uuid.New()
	h := NewArtHandler(&mockArt{}, &mockProfile{
		getArtistByUserID: func(ctx context.Context, uid uuid.UUID) (*profile.ArtistProfile, error) {
			return nil, apperrors.ErrNotFound
		},
	})

	w := serve(t, http.MethodPost, "/api/v1/posts", strings.NewReader(`{"title":"T"}`),
		nil, map[string]string{"X-User-ID": userID.String()}, h.Create)

	assist.Equal(t, http.StatusNotFound, w.Code)
}

func TestOnboardingHandler_ListPending_success(t *testing.T) {
	appID := uuid.New()
	h := NewOnboardingHandler(&mockOnboarding{
		listPending: func(ctx context.Context) ([]onboarding.OnboardingApplication, error) {
			return []onboarding.OnboardingApplication{{
				ID: appID, DisplayName: "New Artist", Status: onboarding.ApprovalStatusPending,
			}}, nil
		},
	})

	w := serve(t, http.MethodGet, "/admin/v1/applications", nil, nil, nil, h.ListPending)

	assist.Equal(t, http.StatusOK, w.Code)
	var resp []dto.OnboardingApplicationResponse
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assist.Len(t, 1, len(resp))
	assist.Equal(t, "New Artist", resp[0].DisplayName)
}

func TestOnboardingHandler_ListPending_error(t *testing.T) {
	h := NewOnboardingHandler(&mockOnboarding{
		listPending: func(ctx context.Context) ([]onboarding.OnboardingApplication, error) {
			return nil, errors.New("db down")
		},
	})

	w := serve(t, http.MethodGet, "/admin/v1/applications", nil, nil, nil, h.ListPending)

	assist.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestOnboardingHandler_GetByID_success(t *testing.T) {
	appID := uuid.New()
	h := NewOnboardingHandler(&mockOnboarding{
		getByID: func(ctx context.Context, id uuid.UUID) (*onboarding.OnboardingApplication, error) {
			assist.Equal(t, appID, id)
			return &onboarding.OnboardingApplication{ID: appID, DisplayName: "Studio", Status: onboarding.ApprovalStatusPending}, nil
		},
	})

	w := serve(t, http.MethodGet, "/admin/v1/applications/"+appID.String(), nil,
		gin.Params{{Key: "id", Value: appID.String()}}, nil, h.GetByID)

	assist.Equal(t, http.StatusOK, w.Code)
	var resp dto.OnboardingApplicationResponse
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assist.Equal(t, "Studio", resp.DisplayName)
}

func TestOnboardingHandler_GetByID_invalidID(t *testing.T) {
	h := NewOnboardingHandler(&mockOnboarding{})
	w := serve(t, http.MethodGet, "/admin/v1/applications/bad", nil,
		gin.Params{{Key: "id", Value: "bad"}}, nil, h.GetByID)

	assist.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOnboardingHandler_Review_success(t *testing.T) {
	appID := uuid.New()
	reviewerID := uuid.New()
	h := NewOnboardingHandler(&mockOnboarding{
		review: func(ctx context.Context, id, rid uuid.UUID, status onboarding.ApprovalStatus, notes string) (*onboarding.OnboardingApplication, error) {
			assist.Equal(t, appID, id)
			assist.Equal(t, reviewerID, rid)
			assist.Equal(t, onboarding.ApprovalStatusApproved, status)
			assist.Equal(t, "looks good", notes)
			return &onboarding.OnboardingApplication{ID: id, Status: status}, nil
		},
	})

	w := serve(t, http.MethodPut, "/admin/v1/applications/"+appID.String(),
		strings.NewReader(`{"status":"approved","notes":"looks good"}`),
		gin.Params{{Key: "id", Value: appID.String()}},
		map[string]string{"X-User-ID": reviewerID.String()}, h.Review)

	assist.Equal(t, http.StatusOK, w.Code)
}

func TestOnboardingHandler_Review_invalidStatus(t *testing.T) {
	appID := uuid.New()
	h := NewOnboardingHandler(&mockOnboarding{})
	w := serve(t, http.MethodPut, "/admin/v1/applications/"+appID.String(),
		strings.NewReader(`{"status":"pending"}`),
		gin.Params{{Key: "id", Value: appID.String()}},
		map[string]string{"X-User-ID": uuid.New().String()}, h.Review)

	assist.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOnboardingHandler_Review_missingUserHeader(t *testing.T) {
	appID := uuid.New()
	h := NewOnboardingHandler(&mockOnboarding{})
	w := serve(t, http.MethodPut, "/admin/v1/applications/"+appID.String(),
		strings.NewReader(`{"status":"approved"}`),
		gin.Params{{Key: "id", Value: appID.String()}}, nil, h.Review)

	assist.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProfileHandler_GetMyProfile_success(t *testing.T) {
	userID := uuid.New()
	profileID := uuid.New()
	h := NewProfileHandler(&mockProfile{
		getArtistByUserID: func(ctx context.Context, uid uuid.UUID) (*profile.ArtistProfile, error) {
			assist.Equal(t, userID, uid)
			return &profile.ArtistProfile{ID: profileID, UserID: userID, Slug: "studio", DisplayName: "Studio"}, nil
		},
	})

	w := serve(t, http.MethodGet, "/api/v1/me/profile", nil, nil,
		map[string]string{"X-User-ID": userID.String()}, h.GetMyProfile)

	assist.Equal(t, http.StatusOK, w.Code)
	var resp dto.ArtistProfileResponse
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assist.Equal(t, profileID, resp.ID)
	assist.Equal(t, "Studio", resp.DisplayName)
}

func TestProfileHandler_UpdateMyProfile_success(t *testing.T) {
	userID := uuid.New()
	h := NewProfileHandler(&mockProfile{
		updateOwnProfile: func(ctx context.Context, uid uuid.UUID, update profile.OwnProfileUpdate) (*profile.ArtistProfile, error) {
			assist.Equal(t, userID, uid)
			assist.Equal(t, "Updated Name", update.DisplayName)
			return &profile.ArtistProfile{ID: uuid.New(), UserID: userID, DisplayName: update.DisplayName, Slug: "updated-name"}, nil
		},
	})

	w := serve(t, http.MethodPut, "/api/v1/me/profile",
		strings.NewReader(`{"display_name":"Updated Name","bio":"New bio"}`),
		nil, map[string]string{"X-User-ID": userID.String()}, h.UpdateMyProfile)

	assist.Equal(t, http.StatusOK, w.Code)
}

func TestOnboardingHandler_Submit_success(t *testing.T) {
	userID := uuid.New()
	appID := uuid.New()
	h := NewOnboardingHandler(&mockOnboarding{
		submit: func(ctx context.Context, applicantID uuid.UUID, applicantType onboarding.ApplicantType, displayName, notes string) (*onboarding.OnboardingApplication, error) {
			assist.Equal(t, userID, applicantID)
			return &onboarding.OnboardingApplication{ID: appID, ApplicantID: applicantID, DisplayName: displayName, Status: onboarding.ApprovalStatusPending}, nil
		},
	})

	w := serve(t, http.MethodPost, "/api/v1/applications",
		strings.NewReader(`{"applicant_type":"artist","display_name":"Studio X","notes":"link"}`),
		nil, map[string]string{"X-User-ID": userID.String()}, h.Submit)

	assist.Equal(t, http.StatusCreated, w.Code)
}

func TestArtHandler_ListMyPosts_success(t *testing.T) {
	userID := uuid.New()
	artistID := uuid.New()
	postID := uuid.New()
	h := NewArtHandler(
		&mockArt{
			listOwnedByArtist: func(ctx context.Context, aid uuid.UUID) ([]art.ArtPost, error) {
				assist.Equal(t, artistID, aid)
				return []art.ArtPost{{ID: postID, ArtistID: artistID, Title: "Draft", Status: art.ArtStatusDraft}}, nil
			},
		},
		&mockProfile{
			getArtistByUserID: func(ctx context.Context, uid uuid.UUID) (*profile.ArtistProfile, error) {
				return &profile.ArtistProfile{ID: artistID, UserID: userID}, nil
			},
		},
	)

	w := serve(t, http.MethodGet, "/api/v1/me/posts", nil, nil,
		map[string]string{"X-User-ID": userID.String()}, h.ListMyPosts)

	assist.Equal(t, http.StatusOK, w.Code)
	var resp []dto.ArtPostResponse
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assist.Len(t, 1, len(resp))
}

func TestEventHandler_ListPending_success(t *testing.T) {
	eventID := uuid.New()
	h := NewEventHandler(&mockEvent{
		listPending: func(ctx context.Context) ([]events.Event, error) {
			return []events.Event{{
				ID: eventID, Title: "Opening", Status: events.EventStatusPending, StartsAt: time.Now().UTC(),
			}}, nil
		},
	})

	w := serve(t, http.MethodGet, "/admin/v1/events", nil, nil, nil, h.ListPending)

	assist.Equal(t, http.StatusOK, w.Code)
	var resp []dto.EventResponse
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assist.Len(t, 1, len(resp))
	assist.Equal(t, "Opening", resp[0].Title)
}

func TestEventHandler_GetByID_success(t *testing.T) {
	eventID := uuid.New()
	h := NewEventHandler(&mockEvent{
		getByID: func(ctx context.Context, id uuid.UUID) (*events.Event, error) {
			assist.Equal(t, eventID, id)
			return &events.Event{ID: eventID, Title: "Gallery Night", Status: events.EventStatusPending, StartsAt: time.Now().UTC()}, nil
		},
	})

	w := serve(t, http.MethodGet, "/admin/v1/events/"+eventID.String(), nil,
		gin.Params{{Key: "id", Value: eventID.String()}}, nil, h.GetByID)

	assist.Equal(t, http.StatusOK, w.Code)
	var resp dto.EventResponse
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assist.Equal(t, "Gallery Night", resp.Title)
}

func TestEventHandler_Review_success(t *testing.T) {
	eventID := uuid.New()
	reviewerID := uuid.New()
	h := NewEventHandler(&mockEvent{
		review: func(ctx context.Context, id, rid uuid.UUID, status events.EventStatus, notes string) (*events.Event, error) {
			assist.Equal(t, eventID, id)
			assist.Equal(t, reviewerID, rid)
			assist.Equal(t, events.EventStatusApproved, status)
			return &events.Event{ID: id, Status: status}, nil
		},
	})

	w := serve(t, http.MethodPut, "/admin/v1/events/"+eventID.String(),
		strings.NewReader(`{"status":"approved","notes":"looks good"}`),
		gin.Params{{Key: "id", Value: eventID.String()}},
		map[string]string{"X-User-ID": reviewerID.String()}, h.Review)

	assist.Equal(t, http.StatusOK, w.Code)
}

func TestEventHandler_Review_invalidStatus(t *testing.T) {
	eventID := uuid.New()
	h := NewEventHandler(&mockEvent{})
	w := serve(t, http.MethodPut, "/admin/v1/events/"+eventID.String(),
		strings.NewReader(`{"status":"pending"}`),
		gin.Params{{Key: "id", Value: eventID.String()}},
		map[string]string{"X-User-ID": uuid.New().String()}, h.Review)

	assist.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEventHandler_Sync_success(t *testing.T) {
	h := NewEventHandler(&mockEvent{
		sync: func(ctx context.Context) (int, error) {
			return 3, nil
		},
	})

	w := serve(t, http.MethodPost, "/admin/v1/events/sync", nil, nil, nil, h.Sync)

	assist.Equal(t, http.StatusOK, w.Code)
	var resp dto.SyncEventsResponse
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assist.Equal(t, 3, resp.Upserted)
}

func TestWriteError_mapping(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"not found", apperrors.ErrNotFound, http.StatusNotFound},
		{"not implemented", apperrors.ErrNotImplemented, http.StatusNotImplemented},
		{"internal", errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := serve(t, http.MethodGet, "/", nil, nil, nil, func(c *gin.Context) {
				writeError(c, tc.err)
			})
			assist.Equal(t, tc.want, w.Code)
		})
	}
}
