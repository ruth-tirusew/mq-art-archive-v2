//go:build integration

package integration_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

func insertAsset(t *testing.T, app *integration.App, owner string) string {
	t.Helper()
	id := uuid.New()
	now := time.Now().UTC()
	_, err := app.Pool.Exec(context.Background(), `
		INSERT INTO media_assets (id, owner_user_id, public_id, secure_url, resource_type, width, height, bytes, folder, blurhash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'image', 100, 100, 1000, 'mq', 'TESTHASH', $5, $5)
	`, id, owner, "mq/"+id.String(), "https://res.cloudinary.com/cloud/image/upload/v1/mq/"+id.String()+".webp", now)
	assist.NoError(t, err)
	return id.String()
}

func TestHTTP_anArtistNeedsFiveWorksToApply_integration(t *testing.T) {
	app := integration.NewApp(t)
	c := caller{t, app.Router()}
	admin := integration.InsertAdminUser(t, app.Pool).String()
	applicant := insertVerifiedUser(t, app, "works@example.com", true)
	stranger := insertVerifiedUser(t, app, "stranger@example.com", true)
	apply := map[string]any{"applicant_type": "artist", "display_name": "Works Applicant", "requested_handle": "works_1"}

	// Nothing uploaded yet.
	rec := c.do("POST", "/api/v1/applications", applicant, apply)
	expect(t, rec, http.StatusBadRequest)
	assist.Contains(t, rec.Body.String(), "at least 5")

	// Four is not enough, and the message says how many are missing.
	addWorks(t, app, c, applicant, domain.MinApplicationWorks-1)
	rec = c.do("POST", "/api/v1/applications", applicant, apply)
	expect(t, rec, http.StatusBadRequest)
	assist.Contains(t, rec.Body.String(), "you have 4")

	// Somebody else's upload can't be used as your work; it looks exactly like a missing one.
	theirs := insertAsset(t, app, stranger)
	expect(t, c.do("POST", "/api/v1/me/application/works", applicant, map[string]any{"media_asset_id": theirs, "title": "Not mine"}), http.StatusNotFound)
	// A blank title is refused.
	expect(t, c.do("POST", "/api/v1/me/application/works", applicant, map[string]any{"media_asset_id": insertAsset(t, app, applicant), "title": "  "}), http.StatusBadRequest)

	addWorks(t, app, c, applicant, 1)
	listed := c.do("GET", "/api/v1/me/application/works", applicant, nil)
	expect(t, listed, http.StatusOK)
	works := decode[[]map[string]any](t, listed)
	assist.Len(t, 5, len(works))
	assist.Equal(t, "TESTHASH", works[0]["blurhash"].(string)) // each image's own blur hash comes along

	rec = c.do("POST", "/api/v1/applications", applicant, apply)
	expect(t, rec, http.StatusCreated)
	submitted := decode[map[string]any](t, rec)
	appID := submitted["id"].(string)
	assist.Len(t, 5, len(submitted["works"].([]any)))

	// While the application is under review the works are locked, but can still be read.
	expect(t, c.do("POST", "/api/v1/me/application/works", applicant, map[string]any{"media_asset_id": insertAsset(t, app, applicant), "title": "Late"}), http.StatusConflict)
	expect(t, c.do("DELETE", "/api/v1/me/application/works/"+works[0]["id"].(string), applicant, nil), http.StatusConflict)
	expect(t, c.do("GET", "/api/v1/me/application/works", applicant, nil), http.StatusOK)

	// The reviewer sees them.
	review := c.do("GET", "/admin/v1/applications/"+appID, admin, nil)
	expect(t, review, http.StatusOK)
	seen := decode[map[string]any](t, review)["works"].([]any)
	assist.Len(t, 5, len(seen))
	assist.Equal(t, "Work 1", seen[0].(map[string]any)["title"].(string))

	// After a rejection the applicant can change them; the decided application keeps what it had.
	expect(t, c.do("PUT", "/admin/v1/applications/"+appID, admin, map[string]any{"status": "rejected", "notes": "Please add stronger pieces"}), http.StatusOK)
	expect(t, c.do("DELETE", "/api/v1/me/application/works/"+works[0]["id"].(string), applicant, nil), http.StatusNoContent)
	expect(t, c.do("DELETE", "/api/v1/me/application/works/"+works[0]["id"].(string), applicant, nil), http.StatusNotFound)
	rec = c.do("POST", "/api/v1/applications", applicant, apply)
	expect(t, rec, http.StatusBadRequest) // back to four
	old := decode[map[string]any](t, c.do("GET", "/admin/v1/applications/"+appID, admin, nil))
	assist.Len(t, 5, len(old["works"].([]any)))
}

func TestHTTP_worksAreOnlyForSignedInUsers_integration(t *testing.T) {
	app := integration.NewApp(t)
	c := caller{t, app.Router()}
	for _, path := range []string{"/api/v1/me/application/works", "/api/v1/me/application/media/sign"} {
		method := http.MethodGet
		if path == "/api/v1/me/application/media/sign" {
			method = http.MethodPost
		}
		expect(t, c.do(method, path, "", nil, "Origin", allowedOrigin), http.StatusUnauthorized)
	}
}
