//go:build integration

package integration_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

func insertStaff(t *testing.T, app *integration.App, role identity.Role) string {
	t.Helper()
	id := uuid.New()
	integration.InsertUser(t, app.Pool, id, role)
	return id.String()
}

// A moderator does the review and content work; what they cannot do is run the admin side of the house.
func TestHTTP_moderatorsCanModerateButNotManageUsersOrSettings_integration(t *testing.T) {
	app := integration.NewApp(t)
	c := caller{t, app.Router()}
	admin := integration.InsertAdminUser(t, app.Pool).String()
	moderator := insertStaff(t, app, identity.RoleModerator)
	artist := insertStaff(t, app, identity.RoleArtist)
	someone := uuid.NewString()

	// Moderation and content work: admins and moderators, nobody else. This checks the GATE (not 401/403): a few
	// handlers (wiki, for one) are not wired into the test app, so the answer past the gate can be a 500.
	allowed := []struct{ method, path string }{
		{"GET", "/admin/v1/applications"},
		{"GET", "/admin/v1/wiki/submissions"},
		{"GET", "/admin/v1/posts"},
		{"GET", "/admin/v1/artists"},
		{"GET", "/admin/v1/events"},
		{"GET", "/admin/v1/articles"},
		{"GET", "/admin/v1/invites"},
		{"GET", "/admin/v1/questionnaire/revisions"},
	}
	for _, r := range allowed {
		for who, user := range map[string]string{"admin": admin, "moderator": moderator} {
			if rec := c.do(r.method, r.path, user, nil); rec.Code == http.StatusForbidden || rec.Code == http.StatusUnauthorized {
				t.Errorf("%s should be allowed %s %s, got %d: %s", who, r.method, r.path, rec.Code, rec.Body.String())
			}
		}
		expect(t, c.do(r.method, r.path, artist, nil), http.StatusForbidden)
		expect(t, c.do(r.method, r.path, "", nil), http.StatusUnauthorized)
	}

	// Admin-only: users and roles, system settings, and inviting other staff.
	adminOnly := []struct {
		method, path string
		body         any
	}{
		{"GET", "/admin/v1/users", nil},
		{"PATCH", "/admin/v1/users/" + someone, map[string]any{"role": "admin"}},
		{"GET", "/admin/v1/settings/scrape", nil},
		{"PUT", "/admin/v1/settings/scrape", map[string]any{}},
		{"GET", "/admin/v1/staff-invites", nil},
		{"POST", "/admin/v1/staff-invites", map[string]any{"email": "x@example.org"}},
		{"POST", "/admin/v1/staff-invites/" + someone + "/resend", nil},
		{"DELETE", "/admin/v1/staff-invites/" + someone, nil},
	}
	for _, r := range adminOnly {
		expect(t, c.do(r.method, r.path, moderator, r.body), http.StatusForbidden)
		expect(t, c.do(r.method, r.path, artist, r.body), http.StatusForbidden)
		expect(t, c.do(r.method, r.path, "", r.body), http.StatusUnauthorized)
		if rec := c.do(r.method, r.path, admin, r.body); rec.Code == http.StatusForbidden || rec.Code == http.StatusUnauthorized {
			t.Errorf("admin should be allowed %s %s, got %d", r.method, r.path, rec.Code)
		}
	}
}

// An admin can turn an existing account into a moderator on the Users page.
func TestHTTP_anAdminCanPromoteAnExistingAccountToModerator_integration(t *testing.T) {
	app := integration.NewApp(t)
	c := caller{t, app.Router()}
	admin := integration.InsertAdminUser(t, app.Pool).String()
	person := insertVerifiedUser(t, app, "promote-me@example.org", true)

	// A plain user has no access to the admin app yet.
	expect(t, c.do("GET", "/admin/v1/applications", person, nil), http.StatusForbidden)

	rec := c.do("PATCH", "/admin/v1/users/"+person, admin, map[string]any{"role": "moderator"})
	expect(t, rec, http.StatusOK)
	assist.Equal(t, "moderator", decode[map[string]any](t, rec)["role"].(string))

	expect(t, c.do("GET", "/admin/v1/applications", person, nil), http.StatusOK)
	expect(t, c.do("GET", "/admin/v1/users", person, nil), http.StatusForbidden)

	// ...and a moderator cannot promote themselves, or anyone else, to admin.
	expect(t, c.do("PATCH", "/admin/v1/users/"+person, person, map[string]any{"role": "admin"}), http.StatusForbidden)

	// An unknown role is still refused.
	expect(t, c.do("PATCH", "/admin/v1/users/"+person, admin, map[string]any{"role": "superuser"}), http.StatusBadRequest)
}

// The whole invite: admin creates it, the invitee opens the link, picks a password and is a moderator.
func TestHTTP_staffInviteLifecycle_integration(t *testing.T) {
	app := integration.NewApp(t)
	c := caller{t, app.Router()}
	admin := integration.InsertAdminUser(t, app.Pool).String()

	// Only moderators can be invited by email; admins are made on the Users page.
	expect(t, c.do("POST", "/admin/v1/staff-invites", admin, map[string]any{"email": "boss@example.org", "role": "admin"}), http.StatusBadRequest)
	expect(t, c.do("POST", "/admin/v1/staff-invites", admin, map[string]any{"email": "not-an-email"}), http.StatusBadRequest)

	// Someone who already has an account is refused with the reason, and nothing is emailed.
	insertVerifiedUser(t, app, "has-account@example.org", true)
	rec := c.do("POST", "/admin/v1/staff-invites", admin, map[string]any{"email": "Has-Account@example.org"})
	expect(t, rec, http.StatusConflict)
	assist.Contains(t, rec.Body.String(), "already has an Artiv account")
	assist.Len(t, 0, len(app.Mailer.Messages()))

	// Create: the response carries the one-time link; the list never does.
	created := c.do("POST", "/admin/v1/staff-invites", admin, map[string]any{"email": "New.Mod@example.org"})
	expect(t, created, http.StatusCreated)
	body := decode[struct {
		Invite map[string]any `json:"invite"`
		Link   string         `json:"link"`
	}](t, created)
	assist.Equal(t, "new.mod@example.org", body.Invite["email"].(string))
	assist.Equal(t, "moderator", body.Invite["role"].(string))
	assist.Equal(t, "live", body.Invite["status"].(string))
	assist.Contains(t, body.Link, "http://localhost:5174/accept-invite?token=")
	token := body.Link[strings.Index(body.Link, "token=")+len("token="):]

	list := c.do("GET", "/admin/v1/staff-invites", admin, nil)
	expect(t, list, http.StatusOK)
	if strings.Contains(list.Body.String(), token) || strings.Contains(list.Body.String(), "token_hash") {
		t.Errorf("the list must never expose a token or hash: %s", list.Body.String())
	}
	mails := app.Mailer.Messages()
	assist.Len(t, 1, len(mails))
	assist.Equal(t, "new.mod@example.org", mails[0].To)
	assist.Contains(t, mails[0].Body, token)

	// Preview is public and shows who the link is for.
	prev := c.do("POST", "/api/v1/staff-invites/preview", "", map[string]any{"token": token})
	expect(t, prev, http.StatusOK)
	assist.Contains(t, prev.Body.String(), "new.mod@example.org")
	expect(t, c.do("POST", "/api/v1/staff-invites/preview", "", map[string]any{"token": "nope"}), http.StatusNotFound)

	// A weak password is refused without using up the link.
	expect(t, c.do("POST", "/api/v1/staff-invites/accept", "", map[string]any{"token": token, "password": "short"}), http.StatusBadRequest)

	// Accept: an account with the invited role, signed in (the session cookie is set).
	accepted := c.do("POST", "/api/v1/staff-invites/accept", "", map[string]any{"token": token, "password": "a-good-password"})
	expect(t, accepted, http.StatusCreated)
	assist.Contains(t, accepted.Header().Get("Set-Cookie"), "HttpOnly")
	user := decode[map[string]any](t, accepted)
	assist.Equal(t, "moderator", user["role"].(string))
	assist.Equal(t, "new.mod@example.org", user["email"].(string))
	assist.Equal(t, true, user["email_verified"].(bool))
	newID := user["id"].(string)

	// The new moderator can work, but not manage users.
	expect(t, c.do("GET", "/admin/v1/applications", newID, nil), http.StatusOK)
	expect(t, c.do("GET", "/admin/v1/users", newID, nil), http.StatusForbidden)

	// The link is single-use, and the list now says it was used.
	expect(t, c.do("POST", "/api/v1/staff-invites/accept", "", map[string]any{"token": token, "password": "a-good-password"}), http.StatusConflict)
	assist.Contains(t, c.do("GET", "/admin/v1/staff-invites", admin, nil).Body.String(), `"status":"claimed"`)
	// Nothing can be resent or revoked once it has been used.
	inviteID := body.Invite["id"].(string)
	expect(t, c.do("POST", "/admin/v1/staff-invites/"+inviteID+"/resend", admin, nil), http.StatusConflict)
	expect(t, c.do("DELETE", "/admin/v1/staff-invites/"+inviteID, admin, nil), http.StatusConflict)
}

func TestHTTP_staffInviteResendAndRevoke_integration(t *testing.T) {
	app := integration.NewApp(t)
	c := caller{t, app.Router()}
	admin := integration.InsertAdminUser(t, app.Pool).String()

	first := decode[struct {
		Invite map[string]any `json:"invite"`
		Link   string         `json:"link"`
	}](t, c.do("POST", "/admin/v1/staff-invites", admin, map[string]any{"email": "mod2@example.org"}))
	firstToken := first.Link[strings.Index(first.Link, "token=")+len("token="):]
	inviteID := first.Invite["id"].(string)

	// Resend: a fresh link, and the old one stops working.
	resent := c.do("POST", "/admin/v1/staff-invites/"+inviteID+"/resend", admin, nil)
	expect(t, resent, http.StatusCreated)
	second := decode[struct {
		Invite map[string]any `json:"invite"`
		Link   string         `json:"link"`
	}](t, resent)
	secondToken := second.Link[strings.Index(second.Link, "token=")+len("token="):]
	if secondToken == firstToken {
		t.Fatal("a resend must issue a new link")
	}
	expect(t, c.do("POST", "/api/v1/staff-invites/accept", "", map[string]any{"token": firstToken, "password": "a-good-password"}), http.StatusConflict)

	// Revoke: the new link stops working too, and nothing was created.
	expect(t, c.do("DELETE", "/admin/v1/staff-invites/"+second.Invite["id"].(string), admin, nil), http.StatusNoContent)
	expect(t, c.do("POST", "/api/v1/staff-invites/accept", "", map[string]any{"token": secondToken, "password": "a-good-password"}), http.StatusConflict)
	var accounts int
	assist.NoError(t, app.Pool.QueryRow(t.Context(), `SELECT COUNT(*) FROM users WHERE email = 'mod2@example.org'`).Scan(&accounts))
	assist.Equal(t, 0, accounts)
}
