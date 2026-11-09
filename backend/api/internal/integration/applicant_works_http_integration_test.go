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

// The whole point of approving an artist: they show up on the site, with their work, without
// having to do anything else first.
func TestHTTP_approvedArtistIsPublicWithTheirWork_integration(t *testing.T) {
	app := integration.NewApp(t)
	c := caller{t, app.Router()}
	admin := integration.InsertAdminUser(t, app.Pool).String()
	applicant := insertVerifiedUser(t, app, "live@example.com", true)

	addWorks(t, app, c, applicant, domain.MinApplicationWorks)
	rec := c.do("POST", "/api/v1/applications", applicant, map[string]any{
		"applicant_type": "artist", "display_name": "Live Artist", "requested_handle": "live_artist",
	})
	expect(t, rec, http.StatusCreated)
	appID := decode[map[string]any](t, rec)["id"].(string)

	// Before approval: nobody can see them.
	expect(t, c.do("GET", "/api/v1/artists/live_artist", "", nil), http.StatusNotFound)
	expect(t, c.do("GET", "/api/v1/artists/live_artist/posts", "", nil), http.StatusNotFound)

	expect(t, c.do("PUT", "/admin/v1/applications/"+appID, admin, map[string]any{"status": "approved"}), http.StatusOK)

	// After: public, with no sign-in, the profile and all five works.
	profile := c.do("GET", "/api/v1/artists/live_artist", "", nil)
	expect(t, profile, http.StatusOK)
	assist.Equal(t, "Live Artist", decode[map[string]any](t, profile)["display_name"].(string))
	expect(t, c.do("GET", "/api/v1/profiles/@live_artist", "", nil), http.StatusOK)

	posts := c.do("GET", "/api/v1/artists/live_artist/posts", "", nil)
	expect(t, posts, http.StatusOK)
	listed := decode[[]map[string]any](t, posts)
	assist.Len(t, 5, len(listed))
	media := listed[0]["media"].([]any)[0].(map[string]any)
	assist.Contains(t, media["url"].(string), "res.cloudinary.com")
	assist.Equal(t, "TESTHASH", media["blurhash"].(string)) // each image keeps its own blur hash

	// They also appear in the artists directory and the public archive.
	directory := c.do("GET", "/api/v1/artists", "", nil)
	expect(t, directory, http.StatusOK)
	assist.Contains(t, directory.Body.String(), "live_artist")
	archive := c.do("GET", "/api/v1/posts", "", nil)
	expect(t, archive, http.StatusOK)
	assist.Contains(t, archive.Body.String(), "Work 1")
}

func TestHTTP_anArtistNeedsAPortraitToApply_integration(t *testing.T) {
	app := integration.NewApp(t)
	c := caller{t, app.Router()}
	admin := integration.InsertAdminUser(t, app.Pool).String()
	applicant := insertVerifiedUser(t, app, "portrait@example.com", true)
	stranger := insertVerifiedUser(t, app, "portrait-stranger@example.com", true)
	apply := map[string]any{"applicant_type": "artist", "display_name": "Portrait Applicant", "requested_handle": "portrait_1"}

	// Five works but no portrait: refused, and the message says what to do.
	for i := 0; i < domain.MinApplicationWorks; i++ {
		expect(t, c.do("POST", "/api/v1/me/application/works", applicant, map[string]any{
			"media_asset_id": insertAsset(t, app, applicant), "title": "Work",
		}), http.StatusCreated)
	}
	rec := c.do("POST", "/api/v1/applications", applicant, apply)
	expect(t, rec, http.StatusBadRequest)
	assist.Contains(t, rec.Body.String(), "portrait")
	expect(t, c.do("GET", "/api/v1/me/application/portrait", applicant, nil), http.StatusNotFound)

	// Someone else's image can't be your portrait.
	expect(t, c.do("PUT", "/api/v1/me/application/portrait", applicant, map[string]any{"media_asset_id": insertAsset(t, app, stranger)}), http.StatusNotFound)

	// Setting a second one replaces the first; it is never counted as a work.
	first := insertAsset(t, app, applicant)
	second := insertAsset(t, app, applicant)
	expect(t, c.do("PUT", "/api/v1/me/application/portrait", applicant, map[string]any{"media_asset_id": first}), http.StatusOK)
	got := c.do("PUT", "/api/v1/me/application/portrait", applicant, map[string]any{"media_asset_id": second})
	expect(t, got, http.StatusOK)
	assist.Equal(t, second, decode[map[string]any](t, got)["media_asset_id"].(string))
	assist.Equal(t, "TESTHASH", decode[map[string]any](t, got)["blurhash"].(string))
	works := decode[[]map[string]any](t, c.do("GET", "/api/v1/me/application/works", applicant, nil))
	assist.Len(t, domain.MinApplicationWorks, len(works))

	// Submitting freezes the portrait onto the application, and the reviewer sees it.
	rec = c.do("POST", "/api/v1/applications", applicant, apply)
	expect(t, rec, http.StatusCreated)
	submitted := decode[map[string]any](t, rec)
	assist.Equal(t, second, submitted["portrait"].(map[string]any)["media_asset_id"].(string))
	appID := submitted["id"].(string)

	// Under review it is locked like the works.
	expect(t, c.do("PUT", "/api/v1/me/application/portrait", applicant, map[string]any{"media_asset_id": insertAsset(t, app, applicant)}), http.StatusConflict)
	expect(t, c.do("DELETE", "/api/v1/me/application/portrait", applicant, nil), http.StatusConflict)
	review := decode[map[string]any](t, c.do("GET", "/admin/v1/applications/"+appID, admin, nil))
	assist.Equal(t, second, review["portrait"].(map[string]any)["media_asset_id"].(string))

	// Approval makes it the profile photo.
	expect(t, c.do("PUT", "/admin/v1/applications/"+appID, admin, map[string]any{"status": "approved"}), http.StatusOK)
	var url string
	assist.NoError(t, app.Pool.QueryRow(context.Background(),
		`SELECT portrait_url FROM artist_profiles WHERE handle = 'portrait_1'`).Scan(&url))
	assist.Contains(t, url, second)
}
