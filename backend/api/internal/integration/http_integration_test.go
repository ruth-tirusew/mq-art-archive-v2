//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

func TestHTTP_health_integration(t *testing.T) {
	app := integration.NewApp(t)
	router := app.Router()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assist.Equal(t, http.StatusOK, rec.Code)

	var resp dto.HealthResponse
	assist.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assist.Equal(t, "ok", resp.Status)
}

func TestHTTP_listPublishedArticles_integration(t *testing.T) {
	app := integration.NewApp(t)
	router := app.Router()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/articles", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assist.Equal(t, http.StatusOK, rec.Code)

	var articles []dto.ArticleResponse
	assist.NoError(t, json.NewDecoder(rec.Body).Decode(&articles))
	assist.GreaterOrEqual(t, len(articles), 1)

	found := false
	for _, a := range articles {
		if a.Slug == "welcome-to-mq" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected seeded article welcome-to-mq in list")
	}
}

func TestHTTP_createArticle_integration(t *testing.T) {
	app := integration.NewApp(t)
	router := app.Router()
	authorID := uuid.New()
	integration.InsertUser(t, app.Pool, authorID, identity.RoleContributor)

	body := `{"title":"HTTP Draft","body":"from integration test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/articles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", authorID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assist.Equal(t, http.StatusCreated, rec.Code)

	var article dto.ArticleResponse
	assist.NoError(t, json.NewDecoder(rec.Body).Decode(&article))
	assist.Equal(t, "http-draft", article.Slug)
	assist.Equal(t, "HTTP Draft", article.Title)
}

func TestHTTP_getArtistBySlug_integration(t *testing.T) {
	app := integration.NewApp(t)
	router := app.Router()

	integration.InsertArtistProfile(t, app.Pool, "http-artist", "HTTP Artist")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/artists/http-artist", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assist.Equal(t, http.StatusOK, rec.Code)
	assist.Contains(t, rec.Body.String(), "HTTP Artist")
}

func TestHTTP_listArtistPosts_integration(t *testing.T) {
	app := integration.NewApp(t)
	router := app.Router()
	ctx := context.Background()

	artistID, _ := integration.InsertArtistProfile(t, app.Pool, "posts-artist", "Posts Artist")

	created, err := app.Art.CreateDraft(ctx, artistID, "Wall Study", "Charcoal", "charcoal")
	assist.NoError(t, err)
	publishedAt := time.Now().UTC()
	_, err = app.Pool.Exec(ctx, `
		UPDATE art_posts SET status = $1, published_at = $2, updated_at = $3 WHERE id = $4
	`, string(art.ArtStatusPublished), publishedAt, publishedAt, created.ID)
	assist.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/artists/posts-artist/posts", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assist.Equal(t, http.StatusOK, rec.Code)
	assist.Contains(t, rec.Body.String(), "Wall Study")
}

func TestHTTP_listPendingApplications_integration(t *testing.T) {
	app := integration.NewApp(t)
	router := app.Router()
	ctx := context.Background()

	now := time.Now().UTC()
	_, err := app.Pool.Exec(ctx, `
		INSERT INTO onboarding_applications (
			id, applicant_id, applicant_type, display_name, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, uuid.New(), uuid.New(), string(onboarding.ApplicantTypeArtist), "HTTP Pending App",
		string(onboarding.ApprovalStatusPending), now, now)
	assist.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/admin/v1/applications", nil)
	req.Header.Set("X-User-ID", integration.InsertAdminUser(t, app.Pool).String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assist.Equal(t, http.StatusOK, rec.Code)
	assist.Contains(t, rec.Body.String(), "HTTP Pending App")
}

func TestHTTP_notFound_integration(t *testing.T) {
	app := integration.NewApp(t)
	router := app.Router()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/artists/missing-artist", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assist.Equal(t, http.StatusNotFound, rec.Code)

	body, err := io.ReadAll(rec.Body)
	assist.NoError(t, err)

	var errResp dto.ErrorResponse
	assist.NoError(t, json.Unmarshal(body, &errResp))
	assist.Equal(t, "not found", errResp.Error)
}

func TestHTTP_createPost_integration(t *testing.T) {
	app := integration.NewApp(t)
	router := app.Router()
	ctx := context.Background()

	_, userID := integration.InsertArtistProfile(t, app.Pool, "create-post-artist", "Create Post Artist")

	body := `{"title":"HTTP Artwork","description":"desc","medium":"watercolor"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assist.Equal(t, http.StatusCreated, rec.Code)
	assist.Contains(t, rec.Body.String(), "HTTP Artwork")

	var created dto.ArtPostResponse
	assist.NoError(t, json.NewDecoder(rec.Body).Decode(&created))

	got, err := app.Art.GetByID(ctx, created.ID)
	assist.NoError(t, err)
	assist.Equal(t, "HTTP Artwork", got.Title)
}

func TestHTTP_getPostByID_integration(t *testing.T) {
	app := integration.NewApp(t)
	router := app.Router()
	ctx := context.Background()

	artistID, userID := integration.InsertArtistProfile(t, app.Pool, "get-post-artist", "Get Post Artist")
	created, err := app.Art.CreateDraft(ctx, artistID, "Visible Work", "desc", "ink")
	assist.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/"+created.ID.String(), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assist.Equal(t, http.StatusOK, rec.Code)
	assist.Contains(t, rec.Body.String(), "Visible Work")
	_ = userID
}

func TestHTTP_getApplicationByID_integration(t *testing.T) {
	app := integration.NewApp(t)
	router := app.Router()
	ctx := context.Background()

	appID := uuid.New()
	now := time.Now().UTC()
	_, err := app.Pool.Exec(ctx, `
		INSERT INTO onboarding_applications (
			id, applicant_id, applicant_type, display_name, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, appID, uuid.New(), string(onboarding.ApplicantTypeArtist), "HTTP App Detail",
		string(onboarding.ApprovalStatusPending), now, now)
	assist.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/admin/v1/applications/"+appID.String(), nil)
	req.Header.Set("X-User-ID", integration.InsertAdminUser(t, app.Pool).String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assist.Equal(t, http.StatusOK, rec.Code)
	assist.Contains(t, rec.Body.String(), "HTTP App Detail")
}

func TestHTTP_reviewApplication_integration(t *testing.T) {
	app := integration.NewApp(t)
	router := app.Router()
	ctx := context.Background()

	appID := uuid.New()
	now := time.Now().UTC()
	_, err := app.Pool.Exec(ctx, `
		INSERT INTO onboarding_applications (
			id, applicant_id, applicant_type, display_name, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, appID, uuid.New(), string(onboarding.ApplicantTypeArtist), "HTTP Review App",
		string(onboarding.ApprovalStatusPending), now, now)
	assist.NoError(t, err)

	reviewerID := integration.InsertAdminUser(t, app.Pool)
	body := `{"status":"approved","notes":"welcome aboard"}`
	req := httptest.NewRequest(http.MethodPut, "/admin/v1/applications/"+appID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", reviewerID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assist.Equal(t, http.StatusOK, rec.Code)

	got, err := app.Onboarding.GetByID(ctx, appID)
	assist.NoError(t, err)
	assist.Equal(t, onboarding.ApprovalStatusApproved, got.Status)
	assist.Equal(t, "welcome aboard", got.Notes)
}
