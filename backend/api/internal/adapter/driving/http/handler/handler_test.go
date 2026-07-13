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
	listPublished func(ctx context.Context, filter content.ListFilter) ([]content.Article, error)
	getBySlug     func(ctx context.Context, slug string) (*content.Article, error)
	createDraft   func(ctx context.Context, authorID uuid.UUID, title, body string) (*content.Article, error)
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

type mockProfile struct {
	getArtistBySlug   func(ctx context.Context, slug string) (*profile.ArtistProfile, error)
	getArtistByHandle func(ctx context.Context, handle string) (*profile.ArtistProfile, error)
	listApproved      func(ctx context.Context, filter profile.ListFilter) ([]profile.ArtistProfile, error)
	getArtistByID     func(ctx context.Context, id uuid.UUID) (*profile.ArtistProfile, error)
	getArtistByUserID func(ctx context.Context, userID uuid.UUID) (*profile.ArtistProfile, error)
	updateArtist      func(ctx context.Context, p profile.ArtistProfile) (*profile.ArtistProfile, error)
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
func (m *mockProfile) GetArtistByID(ctx context.Context, id uuid.UUID) (*profile.ArtistProfile, error) {
	return m.getArtistByID(ctx, id)
}
func (m *mockProfile) GetArtistByUserID(ctx context.Context, userID uuid.UUID) (*profile.ArtistProfile, error) {
	return m.getArtistByUserID(ctx, userID)
}
func (m *mockProfile) UpdateArtist(ctx context.Context, p profile.ArtistProfile) (*profile.ArtistProfile, error) {
	return m.updateArtist(ctx, p)
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
	listByArtist   func(ctx context.Context, artistID uuid.UUID) ([]art.ArtPost, error)
	listPublished  func(ctx context.Context, filter art.ListFilter) ([]art.ArtPostWithArtist, error)
	getByID        func(ctx context.Context, id uuid.UUID) (*art.ArtPost, error)
	createDraft    func(ctx context.Context, artistID uuid.UUID, title, description, medium string) (*art.ArtPost, error)
}

func (m *mockArt) ListByArtist(ctx context.Context, artistID uuid.UUID) ([]art.ArtPost, error) {
	return m.listByArtist(ctx, artistID)
}
func (m *mockArt) ListPublished(ctx context.Context, filter art.ListFilter) ([]art.ArtPostWithArtist, error) {
	if m.listPublished != nil {
		return m.listPublished(ctx, filter)
	}
	return nil, nil
}
func (m *mockArt) GetByID(ctx context.Context, id uuid.UUID) (*art.ArtPost, error) {
	return m.getByID(ctx, id)
}
func (m *mockArt) CreateDraft(ctx context.Context, artistID uuid.UUID, title, description, medium string) (*art.ArtPost, error) {
	return m.createDraft(ctx, artistID, title, description, medium)
}

type mockOnboarding struct {
	listPending func(ctx context.Context) ([]onboarding.OnboardingApplication, error)
	getByID     func(ctx context.Context, id uuid.UUID) (*onboarding.OnboardingApplication, error)
	review      func(ctx context.Context, id, reviewerID uuid.UUID, status onboarding.ApprovalStatus, notes string) (*onboarding.OnboardingApplication, error)
}

func (m *mockOnboarding) ListPending(ctx context.Context) ([]onboarding.OnboardingApplication, error) {
	return m.listPending(ctx)
}
func (m *mockOnboarding) GetByID(ctx context.Context, id uuid.UUID) (*onboarding.OnboardingApplication, error) {
	return m.getByID(ctx, id)
}
func (m *mockOnboarding) Review(ctx context.Context, id, reviewerID uuid.UUID, status onboarding.ApprovalStatus, notes string) (*onboarding.OnboardingApplication, error) {
	return m.review(ctx, id, reviewerID, status, notes)
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
				ID: instID, Slug: slug, Name: "National Gallery", Status: institution.StatusApproved,
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
			createDraft: func(ctx context.Context, aid uuid.UUID, title, description, medium string) (*art.ArtPost, error) {
				assist.Equal(t, artistID, aid)
				assist.Equal(t, "New Work", title)
				return &art.ArtPost{ID: postID, ArtistID: artistID, Title: title, Status: art.ArtStatusDraft}, nil
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
	var resp []onboarding.OnboardingApplication
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assist.Len(t, 1, len(resp))
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
	var resp onboarding.OnboardingApplication
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
