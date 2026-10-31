//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

const allowedOrigin = "http://localhost:5173"

type caller struct {
	t      *testing.T
	router *gin.Engine
}

// do sends a JSON request as the given user ("" for anonymous). extra sets additional headers.
func (c caller) do(method, path, user string, body any, extra ...string) *httptest.ResponseRecorder {
	c.t.Helper()
	var rd *strings.Reader
	if body != nil {
		b, err := json.Marshal(body)
		assist.NoError(c.t, err)
		rd = strings.NewReader(string(b))
	} else {
		rd = strings.NewReader("")
	}
	req := httptest.NewRequest(method, path, rd)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if user != "" {
		req.Header.Set("X-User-ID", user)
	}
	for i := 0; i+1 < len(extra); i += 2 {
		req.Header.Set(extra[i], extra[i+1])
	}
	rec := httptest.NewRecorder()
	c.router.ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	return v
}

func expect(t *testing.T, rec *httptest.ResponseRecorder, status int) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, status, rec.Body.String())
	}
}

func insertVerifiedUser(t *testing.T, app *integration.App, email string, verified bool) string {
	t.Helper()
	id := uuid.New()
	now := time.Now().UTC()
	var verifiedAt any
	if verified {
		verifiedAt = now
	}
	_, err := app.Pool.Exec(context.Background(), `
		INSERT INTO users (id, email, role, display_name, email_verified_at, created_at, updated_at)
		VALUES ($1, $2, 'public', $3, $4, $5, $5)
	`, id, email, "Test User", verifiedAt, now)
	assist.NoError(t, err)
	return id.String()
}

func questionPayload(withNew bool) map[string]any {
	qs := []map[string]any{
		{"key": "bio", "type": "long_text", "prompt_en": "Tell us about your practice", "prompt_am": "ስለ ስራዎ ይንገሩን", "required": true, "is_public": true, "max_length": 500},
		{"key": "income", "type": "short_text", "prompt_en": "Income range", "is_public": false},
		{"key": "city", "type": "short_text", "prompt_en": "City", "is_public": true},
	}
	if withNew {
		qs = append(qs, map[string]any{"key": "influences", "type": "long_text", "prompt_en": "Influences", "is_public": true})
	}
	return map[string]any{"change_kind": "substantive", "note": "rev", "questions": qs}
}

func TestHTTP_questionnaireInviteLifecycle_integration(t *testing.T) {
	app := integration.NewApp(t)
	c := caller{t, app.Router()}
	admin := integration.InsertAdminUser(t, app.Pool).String()
	artist := insertVerifiedUser(t, app, "ada@example.com", true)

	// Nothing published yet: the form 404s so the client falls back to the plain application.
	expect(t, c.do("GET", "/api/v1/me/questionnaire", artist, nil), http.StatusNotFound)

	// Admin builds and publishes revision 1.
	expect(t, c.do("PUT", "/admin/v1/questionnaire/draft", admin, questionPayload(false)), http.StatusOK)
	published := c.do("POST", "/admin/v1/questionnaire/draft/publish", admin, nil)
	expect(t, published, http.StatusOK)
	rev1 := decode[map[string]any](t, published)
	assist.Equal(t, "published", rev1["status"].(string))
	assist.Equal(t, 1.0, rev1["number"].(float64))

	// Admin invites the artist. The response carries the one-time token and link; the list never does.
	created := c.do("POST", "/admin/v1/invites", admin, map[string]any{"email": "ada@example.com", "display_name": "Ada Lovelace", "handle": "ada_1"})
	expect(t, created, http.StatusCreated)
	inv := decode[struct {
		Invite map[string]any `json:"invite"`
		Token  string         `json:"token"`
		Link   string         `json:"link"`
	}](t, created)
	assist.Contains(t, inv.Link, "/apply?invite="+inv.Token)
	assist.Equal(t, "live", inv.Invite["status"].(string))
	list := c.do("GET", "/admin/v1/invites", admin, nil)
	expect(t, list, http.StatusOK)
	if strings.Contains(list.Body.String(), inv.Token) || strings.Contains(list.Body.String(), "token") {
		t.Errorf("invite list must not expose tokens or hashes: %s", list.Body.String())
	}
	mails := app.Mailer.Messages()
	assist.Len(t, 1, len(mails))
	assist.Equal(t, "ada@example.com", mails[0].To)
	assist.Contains(t, mails[0].Body, inv.Token)

	// The handle is reserved while the invite is live.
	taken := c.do("GET", "/api/v1/handles/ada_1/available", "", nil)
	assist.Contains(t, taken.Body.String(), `"available":false`)

	// Public preview, token in the body.
	prev := c.do("POST", "/api/v1/invites/preview", "", map[string]any{"token": inv.Token})
	expect(t, prev, http.StatusOK)
	assist.Contains(t, prev.Body.String(), "ada@example.com")
	expect(t, c.do("POST", "/api/v1/invites/preview", "", map[string]any{"token": "nope"}), http.StatusNotFound)
	// ...and never reachable via the URL, where it would end up in logs.
	assert404or405 := c.do("GET", "/api/v1/invites/preview?token="+inv.Token, "", nil)
	if assert404or405.Code != http.StatusNotFound && assert404or405.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET preview should not exist, got %d", assert404or405.Code)
	}

	// Required answers are enforced when submitting, and a failed attempt does not burn the invite.
	claim := map[string]any{"token": inv.Token}
	rec := c.do("POST", "/api/v1/applications/invite", artist, claim)
	expect(t, rec, http.StatusBadRequest)
	assist.Contains(t, rec.Body.String(), "bio")

	// The artist answers: rich text in Amharic, with hostile markup that must not survive.
	dirty := `<p>ሰላም <strong>ዓለም</strong></p><script>alert(1)</script><a href="javascript:alert(2)" onclick="x()">x</a>`
	expect(t, c.do("PUT", "/api/v1/me/questionnaire/answers", artist, map[string]any{"answers": []map[string]any{
		{"key": "bio", "body": dirty, "format": "rich", "lang": "am"},
		{"key": "income", "body": "private figure", "format": "plain"},
		{"key": "city", "body": "Addis Ababa"},
	}}), http.StatusNoContent)

	form := c.do("GET", "/api/v1/me/questionnaire", artist, nil)
	expect(t, form, http.StatusOK)
	for _, bad := range []string{"<script", "javascript:", "onclick"} {
		if strings.Contains(form.Body.String(), bad) {
			t.Errorf("stored/served answer contains %q: %s", bad, form.Body.String())
		}
	}
	assist.Contains(t, form.Body.String(), "ሰላም")

	// Now the claim works, links the invite and the revision, and still waits for an admin.
	rec = c.do("POST", "/api/v1/applications/invite", artist, claim)
	expect(t, rec, http.StatusCreated)
	application := decode[map[string]any](t, rec)
	assist.Equal(t, "pending", application["status"].(string))
	assist.Equal(t, "ada_1", application["requested_handle"].(string))
	assist.Equal(t, inv.Invite["id"].(string), application["invite_id"].(string))
	assist.Equal(t, rev1["id"].(string), application["questionnaire_revision_id"].(string))
	appID := application["id"].(string)

	// Single use, and answers freeze while the application is under review.
	expect(t, c.do("POST", "/api/v1/applications/invite", artist, claim), http.StatusConflict)
	expect(t, c.do("PUT", "/api/v1/me/questionnaire/answers", artist, map[string]any{"answers": []map[string]any{{"key": "city", "body": "Elsewhere"}}}), http.StatusConflict)
	assist.Contains(t, c.do("GET", "/api/v1/me/questionnaire", artist, nil).Body.String(), `"locked":true`)

	// The reviewer sees every answer, including the private one.
	review := c.do("GET", "/admin/v1/applications/"+appID+"/answers", admin, nil)
	expect(t, review, http.StatusOK)
	assist.Contains(t, review.Body.String(), "private figure")
	assist.Contains(t, review.Body.String(), "Addis Ababa")

	// Approve. The profile starts as a private draft, so its answers are not public yet.
	expect(t, c.do("PUT", "/admin/v1/applications/"+appID, admin, map[string]any{"status": "approved"}), http.StatusOK)
	expect(t, c.do("GET", "/api/v1/artists/ada_1/answers", "", nil), http.StatusNotFound)

	// Once the profile is public, visitors see only the public answers.
	_, err := app.Pool.Exec(context.Background(), `UPDATE artist_profiles SET status = 'approved' WHERE slug = 'ada_1'`)
	assist.NoError(t, err)
	pub := c.do("GET", "/api/v1/artists/ada_1/answers", "", nil)
	expect(t, pub, http.StatusOK)
	assist.Contains(t, pub.Body.String(), "ሰላም")
	assist.Contains(t, pub.Body.String(), "Addis Ababa")
	if strings.Contains(pub.Body.String(), "private figure") || strings.Contains(pub.Body.String(), "Income") {
		t.Errorf("private answer leaked publicly: %s", pub.Body.String())
	}

	// Taken down again (e.g. moderation): the answers disappear with it.
	_, err = app.Pool.Exec(context.Background(), `UPDATE artist_profiles SET status = 'draft' WHERE slug = 'ada_1'`)
	assist.NoError(t, err)
	expect(t, c.do("GET", "/api/v1/artists/ada_1/answers", "", nil), http.StatusNotFound)
	_, err = app.Pool.Exec(context.Background(), `UPDATE artist_profiles SET status = 'approved' WHERE slug = 'ada_1'`)
	assist.NoError(t, err)

	// Revision 2 adds a question. Nudging is optional and one-shot; nothing about the profile changes.
	expect(t, c.do("PUT", "/admin/v1/questionnaire/draft", admin, questionPayload(true)), http.StatusOK)
	expect(t, c.do("POST", "/admin/v1/questionnaire/draft/publish", admin, nil), http.StatusOK)
	preview := c.do("GET", "/admin/v1/questionnaire/nudge", admin, nil)
	expect(t, preview, http.StatusOK)
	assist.Contains(t, preview.Body.String(), `"can_send":true`)

	sent := c.do("POST", "/admin/v1/questionnaire/nudge", admin, nil)
	expect(t, sent, http.StatusOK)
	assist.Contains(t, sent.Body.String(), `"sent":`)
	again := c.do("POST", "/admin/v1/questionnaire/nudge", admin, nil)
	assist.Contains(t, again.Body.String(), `"sent":0`)

	nudgedAda := false
	for _, m := range app.Mailer.Messages() {
		if m.To == "ada@example.com" && strings.Contains(m.Subject, "New questions") {
			nudgedAda = true
		}
	}
	assist.Equal(t, true, nudgedAda)

	prompt := c.do("GET", "/api/v1/me/questionnaire/nudge", artist, nil)
	expect(t, prompt, http.StatusOK)
	assist.Contains(t, prompt.Body.String(), "influences")
	revID := decode[map[string]any](t, prompt)["revision_id"].(string)
	expect(t, c.do("POST", "/api/v1/me/questionnaire/nudges/"+revID+"/dismiss", artist, nil), http.StatusNoContent)
	expect(t, c.do("GET", "/api/v1/me/questionnaire/nudge", artist, nil), http.StatusNotFound)

	// An artist who is already approved can answer the new question; the profile page picks it up.
	expect(t, c.do("PUT", "/api/v1/me/questionnaire/answers", artist, map[string]any{"answers": []map[string]any{{"key": "influences", "body": "Gebre Kristos"}}}), http.StatusNoContent)
	assist.Contains(t, c.do("GET", "/api/v1/artists/ada_1/answers", "", nil).Body.String(), "Gebre Kristos")
}

func TestHTTP_questionnaireSecurity_integration(t *testing.T) {
	app := integration.NewApp(t)
	c := caller{t, app.Router()}
	admin := integration.InsertAdminUser(t, app.Pool).String()
	reader := insertVerifiedUser(t, app, "reader@example.com", true)
	stranger := insertVerifiedUser(t, app, "stranger@example.com", true)
	unverified := insertVerifiedUser(t, app, "ada@example.com", false)

	expect(t, c.do("PUT", "/admin/v1/questionnaire/draft", admin, questionPayload(false)), http.StatusOK)
	expect(t, c.do("POST", "/admin/v1/questionnaire/draft/publish", admin, nil), http.StatusOK)
	created := c.do("POST", "/admin/v1/invites", admin, map[string]any{"email": "ada@example.com", "handle": "ada_1"})
	expect(t, created, http.StatusCreated)
	token := decode[struct{ Token string }](t, created).Token
	invitedAnswers := map[string]any{"answers": []map[string]any{{"key": "bio", "body": "hello"}}}

	t.Run("admin routes reject non-admins", func(t *testing.T) {
		for _, r := range []struct{ method, path string }{
			{"GET", "/admin/v1/questionnaire/draft"}, {"PUT", "/admin/v1/questionnaire/draft"},
			{"POST", "/admin/v1/questionnaire/draft/publish"}, {"GET", "/admin/v1/questionnaire/revisions"},
			{"GET", "/admin/v1/questionnaire/nudge"}, {"POST", "/admin/v1/questionnaire/nudge"},
			{"GET", "/admin/v1/invites"}, {"POST", "/admin/v1/invites"},
			{"POST", "/admin/v1/invites/" + uuid.NewString() + "/resend"}, {"DELETE", "/admin/v1/invites/" + uuid.NewString()},
			{"GET", "/admin/v1/applications/" + uuid.NewString() + "/answers"},
		} {
			var body any
			if r.method == "PUT" || r.path == "/admin/v1/invites" {
				body = map[string]any{"email": "x@example.com"}
			}
			expect(t, c.do(r.method, r.path, reader, body), http.StatusForbidden)
			expect(t, c.do(r.method, r.path, "", body), http.StatusUnauthorized)
		}
	})

	t.Run("signed-in routes reject anonymous users", func(t *testing.T) {
		expect(t, c.do("GET", "/api/v1/me/questionnaire", "", nil), http.StatusUnauthorized)
		expect(t, c.do("PUT", "/api/v1/me/questionnaire/answers", "", invitedAnswers), http.StatusUnauthorized)
		expect(t, c.do("GET", "/api/v1/me/questionnaire/nudge", "", nil), http.StatusUnauthorized)
		expect(t, c.do("POST", "/api/v1/applications/invite", "", map[string]any{"token": token}), http.StatusUnauthorized)
	})

	t.Run("csrf: foreign origins are blocked on every state-changing route", func(t *testing.T) {
		evil := []string{"Origin", "https://evil.example"}
		for _, r := range []struct {
			method, path, user string
			body               any
		}{
			{"PUT", "/api/v1/me/questionnaire/answers", reader, invitedAnswers},
			{"POST", "/api/v1/me/questionnaire/nudges/" + uuid.NewString() + "/dismiss", reader, nil},
			{"POST", "/api/v1/applications/invite", reader, map[string]any{"token": token}},
			{"POST", "/api/v1/invites/preview", "", map[string]any{"token": token}},
			{"POST", "/admin/v1/invites", admin, map[string]any{"email": "z@example.com"}},
			{"DELETE", "/admin/v1/invites/" + uuid.NewString(), admin, nil},
			{"PUT", "/admin/v1/questionnaire/draft", admin, questionPayload(false)},
			{"POST", "/admin/v1/questionnaire/draft/publish", admin, nil},
			{"POST", "/admin/v1/questionnaire/nudge", admin, nil},
		} {
			expect(t, c.do(r.method, r.path, r.user, r.body, evil...), http.StatusForbidden)
		}
		// A cross-site fetch that omits Origin is caught by Sec-Fetch-Site.
		expect(t, c.do("POST", "/admin/v1/questionnaire/nudge", admin, nil, "Sec-Fetch-Site", "cross-site"), http.StatusForbidden)
		// The real app's origin is accepted.
		expect(t, c.do("PUT", "/api/v1/me/questionnaire/answers", reader, invitedAnswers, "Origin", allowedOrigin), http.StatusNoContent)
	})

	t.Run("csrf: non-json bodies are rejected", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/api/v1/me/questionnaire/answers", strings.NewReader(`{"answers":[]}`))
		req.Header.Set("X-User-ID", reader)
		req.Header.Set("Content-Type", "text/plain")
		rec := httptest.NewRecorder()
		c.router.ServeHTTP(rec, req)
		expect(t, rec, http.StatusUnsupportedMediaType)
	})

	t.Run("invites are bound to a verified matching email", func(t *testing.T) {
		claim := map[string]any{"token": token}
		expect(t, c.do("POST", "/api/v1/applications/invite", stranger, claim), http.StatusForbidden)   // someone else's invite
		expect(t, c.do("POST", "/api/v1/applications/invite", unverified, claim), http.StatusForbidden) // right email, not verified
		// Neither attempt used the invite up.
		assist.Contains(t, c.do("POST", "/api/v1/invites/preview", "", map[string]any{"token": token}).Body.String(), `"status":"live"`)
	})

	t.Run("malformed and oversized input is rejected", func(t *testing.T) {
		expect(t, c.do("PUT", "/api/v1/me/questionnaire/answers", reader, map[string]any{"answers": []map[string]any{{"key": "nope", "body": "x"}}}), http.StatusBadRequest)
		expect(t, c.do("PUT", "/api/v1/me/questionnaire/answers", reader, map[string]any{"answers": []map[string]any{{"key": "bio", "body": "x", "format": "html"}}}), http.StatusBadRequest)
		expect(t, c.do("PUT", "/api/v1/me/questionnaire/answers", reader, map[string]any{"answers": []map[string]any{}}), http.StatusBadRequest)
		expect(t, c.do("PUT", "/api/v1/me/questionnaire/answers", reader, map[string]any{"answers": []map[string]any{{"key": "bio", "body": strings.Repeat("a", 600)}}}), http.StatusBadRequest) // over max_length
		expect(t, c.do("POST", "/admin/v1/invites", admin, map[string]any{"email": "not-an-email"}), http.StatusBadRequest)
		expect(t, c.do("POST", "/admin/v1/invites", admin, map[string]any{"email": "a@example.com", "handle": "Bad Handle!"}), http.StatusBadRequest)
		expect(t, c.do("DELETE", "/admin/v1/invites/not-a-uuid", admin, nil), http.StatusBadRequest)
		// SQL-looking input is just data: it is stored verbatim and nothing breaks.
		inj := "'; DROP TABLE questionnaire_answers; --"
		expect(t, c.do("PUT", "/api/v1/me/questionnaire/answers", reader, map[string]any{"answers": []map[string]any{{"key": "city", "body": inj}}}), http.StatusNoContent)
		assist.Contains(t, c.do("GET", "/api/v1/me/questionnaire", reader, nil).Body.String(), "DROP TABLE")
		var n int
		assist.NoError(t, app.Pool.QueryRow(context.Background(), `SELECT count(*) FROM questionnaire_answers`).Scan(&n))
		assist.GreaterOrEqual(t, n, 1)
	})

	t.Run("plain applications also need the questionnaire once one is published", func(t *testing.T) {
		applicant := insertVerifiedUser(t, app, "plain@example.com", true)
		body := map[string]any{"applicant_type": "artist", "display_name": "Plain Applicant", "requested_handle": "plain_1"}
		rec := c.do("POST", "/api/v1/applications", applicant, body)
		expect(t, rec, http.StatusBadRequest)
		assist.Contains(t, rec.Body.String(), "bio")
		expect(t, c.do("PUT", "/api/v1/me/questionnaire/answers", applicant, invitedAnswers), http.StatusNoContent)
		expect(t, c.do("POST", "/api/v1/applications", applicant, body), http.StatusCreated)
	})

	t.Run("a revoked invite cannot be used", func(t *testing.T) {
		expect(t, c.do("DELETE", "/admin/v1/invites/"+decode[struct {
			Invite struct{ ID string } `json:"invite"`
		}](t, created).Invite.ID, admin, nil), http.StatusNoContent)
		// The invited person now verifies their email, so only the revocation stands in the way.
		_, err := app.Pool.Exec(context.Background(), `UPDATE users SET email_verified_at = now() WHERE id = $1`, unverified)
		assist.NoError(t, err)
		matching := unverified
		expect(t, c.do("PUT", "/api/v1/me/questionnaire/answers", matching, invitedAnswers), http.StatusNoContent)
		expect(t, c.do("POST", "/api/v1/applications/invite", matching, map[string]any{"token": token}), http.StatusConflict)
	})
}
