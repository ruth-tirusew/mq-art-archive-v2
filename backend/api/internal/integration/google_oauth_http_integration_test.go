//go:build integration

package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driven/auth"
	"github.com/mq/api/internal/port/outbound"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

const (
	webOrigin         = "http://localhost:5173"
	sessionCookieName = "mq_access_token"
)

// browser keeps cookies between requests the way a real one does, so the tests follow the same redirects and
// carry the same cookies a person would: the state cookie from "Continue with Google" back to the callback, and
// the session cookie after signing in.
type browser struct {
	t       *testing.T
	router  *gin.Engine
	cookies map[string]string
}

func newBrowser(t *testing.T, app *integration.App) *browser {
	return &browser{t: t, router: app.Router(), cookies: map[string]string{}}
}

func (b *browser) send(method, path, body string) *httptest.ResponseRecorder {
	b.t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if method != http.MethodGet {
		req.Header.Set("Origin", webOrigin)
	}
	for name, value := range b.cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}
	rec := httptest.NewRecorder()
	b.router.ServeHTTP(rec, req)
	for _, c := range (&http.Response{Header: rec.Header()}).Cookies() {
		if c.MaxAge < 0 || c.Value == "" {
			delete(b.cookies, c.Name)
		} else {
			b.cookies[c.Name] = c.Value
		}
	}
	return rec
}

func (b *browser) get(path string) *httptest.ResponseRecorder {
	return b.send(http.MethodGet, path, "")
}

func (b *browser) signedIn() bool { _, ok := b.cookies[sessionCookieName]; return ok }

// me is /auth/me for whoever this browser is signed in as.
func (b *browser) me() map[string]any {
	b.t.Helper()
	rec := b.get("/api/v1/auth/me")
	expect(b.t, rec, http.StatusOK)
	return decode[map[string]any](b.t, rec)
}

// startGoogle is a click on "Continue with Google": the API redirects to Google carrying a signed state.
func (b *browser) startGoogle(path, returnTo string) string {
	b.t.Helper()
	rec := b.get(path + "?return_to=" + url.QueryEscape(returnTo))
	expect(b.t, rec, http.StatusFound)
	loc, err := url.Parse(rec.Header().Get("Location"))
	assist.NoError(b.t, err)
	assist.Equal(b.t, "accounts.google.test", loc.Host)
	return loc.Query().Get("state")
}

// backFromGoogle is Google sending the person back to the API with a code.
func (b *browser) backFromGoogle(code, state string) *httptest.ResponseRecorder {
	return b.get("/api/v1/auth/google/callback?code=" + url.QueryEscape(code) + "&state=" + url.QueryEscape(state))
}

func locationQuery(t *testing.T, rec *httptest.ResponseRecorder) url.Values {
	t.Helper()
	loc, err := url.Parse(rec.Header().Get("Location"))
	assist.NoError(t, err)
	return loc.Query()
}

// insertPasswordUser makes an account the way registration does: an email, a bcrypt password and (optionally)
// a verified address. Returns its id.
func insertPasswordUser(t *testing.T, app *integration.App, email, password string, verified bool) uuid.UUID {
	t.Helper()
	hash, err := auth.NewBcryptPasswordHasher().Hash(password)
	assist.NoError(t, err)
	id := uuid.New()
	var verifiedAt *time.Time
	if verified {
		now := time.Now().UTC()
		verifiedAt = &now
	}
	_, err = app.Pool.Exec(context.Background(), `
		INSERT INTO users (id, email, role, created_at, updated_at, password_hash, email_verified_at)
		VALUES ($1, $2, 'public', NOW(), NOW(), $3, $4)`, id, email, hash, verifiedAt)
	assist.NoError(t, err)
	return id
}

func (b *browser) loginWithPassword(email, password string) {
	b.t.Helper()
	rec := b.send(http.MethodPost, "/api/v1/auth/login", `{"email":"`+email+`","password":"`+password+`"}`)
	expect(b.t, rec, http.StatusOK)
	assist.Equal(b.t, true, b.signedIn())
}

func countUsersWithEmail(t *testing.T, app *integration.App, email string) int {
	t.Helper()
	var n int
	assist.NoError(t, app.Pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM users WHERE LOWER(email) = LOWER($1)`, email).Scan(&n))
	return n
}

var returnToCallback = webOrigin + "/auth/callback?return_to=%2Fstudio"

// --- signing in with Google -------------------------------------------------------------------------

func TestHTTP_googleSignIn_aNewPersonGetsAVerifiedAccountAndASession_integration(t *testing.T) {
	app := integration.NewApp(t)
	b := newBrowser(t, app)
	app.Google.Grant("code-new", outbound.OAuthUserInfo{ProviderUserID: "g-new", Email: "new.person@gmail.com", EmailVerified: true})

	state := b.startGoogle("/api/v1/auth/google", returnToCallback)
	rec := b.backFromGoogle("code-new", state)

	expect(t, rec, http.StatusFound)
	assist.Equal(t, returnToCallback, rec.Header().Get("Location"))
	assist.Equal(t, true, b.signedIn()) // the session cookie was set on the way back
	me := b.me()
	assist.Equal(t, "new.person@gmail.com", me["email"].(string))
	assist.Equal(t, true, me["email_verified"].(bool))
	assist.Equal(t, true, me["google_linked"].(bool))
	assist.Equal(t, false, me["has_password"].(bool))
}

func TestHTTP_googleSignIn_anExistingPasswordAccountIsNotTakenOver_integration(t *testing.T) {
	app := integration.NewApp(t)
	insertPasswordUser(t, app, "ada@gmail.com", "ada-password-1", true)
	b := newBrowser(t, app)
	app.Google.Grant("code-ada", outbound.OAuthUserInfo{ProviderUserID: "g-ada", Email: "ada@gmail.com", EmailVerified: true})

	state := b.startGoogle("/api/v1/auth/google", returnToCallback)
	rec := b.backFromGoogle("code-ada", state)

	// Not signed in, and sent back to the app with a flag the sign-in page turns into an explanation.
	expect(t, rec, http.StatusFound)
	assist.Equal(t, "account_exists", locationQuery(t, rec).Get("error"))
	assist.Equal(t, false, b.signedIn())
	assist.Equal(t, 1, countUsersWithEmail(t, app, "ada@gmail.com"))
}

// Google can report an address in a different case than we stored it. That must still count as the same
// person's account, not create a second one.
func TestHTTP_googleSignIn_emailCaseDoesNotCreateADuplicateAccount_integration(t *testing.T) {
	app := integration.NewApp(t)
	insertPasswordUser(t, app, "grace@gmail.com", "grace-password-1", true)
	b := newBrowser(t, app)
	app.Google.Grant("code-grace", outbound.OAuthUserInfo{ProviderUserID: "g-grace", Email: "Grace@Gmail.com", EmailVerified: true})

	rec := b.backFromGoogle("code-grace", b.startGoogle("/api/v1/auth/google", returnToCallback))

	assist.Equal(t, "account_exists", locationQuery(t, rec).Get("error"))
	assist.Equal(t, false, b.signedIn())
	assist.Equal(t, 1, countUsersWithEmail(t, app, "grace@gmail.com"))
}

func TestHTTP_googleSignIn_anUnverifiedGoogleEmailIsRefused_integration(t *testing.T) {
	app := integration.NewApp(t)
	b := newBrowser(t, app)
	app.Google.Grant("code-unv", outbound.OAuthUserInfo{ProviderUserID: "g-unv", Email: "someone@gmail.com", EmailVerified: false})

	rec := b.backFromGoogle("code-unv", b.startGoogle("/api/v1/auth/google", returnToCallback))

	expect(t, rec, http.StatusUnauthorized)
	assist.Equal(t, false, b.signedIn())
	assist.Equal(t, 0, countUsersWithEmail(t, app, "someone@gmail.com"))
}

func TestHTTP_googleSignIn_onlyTheBrowserThatStartedItCanFinishIt_integration(t *testing.T) {
	app := integration.NewApp(t)
	app.Google.Grant("code-x", outbound.OAuthUserInfo{ProviderUserID: "g-x", Email: "x@gmail.com", EmailVerified: true})

	attacker := newBrowser(t, app)
	state := attacker.startGoogle("/api/v1/auth/google", returnToCallback)

	// A victim opens the attacker's callback link: their browser never held the state cookie.
	victim := newBrowser(t, app)
	rec := victim.backFromGoogle("code-x", state)
	expect(t, rec, http.StatusBadRequest)
	assist.Equal(t, false, victim.signedIn())

	// A return address outside the allowed origins is refused up front.
	rec = attacker.get("/api/v1/auth/google?return_to=" + url.QueryEscape("https://evil.example/auth/callback"))
	if rec.Code == http.StatusFound {
		t.Fatalf("a return_to outside the allowed origins must not start a sign-in, got a redirect to %s", rec.Header().Get("Location"))
	}
}

// --- linking Google to a password account -------------------------------------------------------------

func TestHTTP_linkGoogle_aPasswordUserCanConnectGoogleAndThenSignInWithIt_integration(t *testing.T) {
	app := integration.NewApp(t)
	id := insertPasswordUser(t, app, "linda@gmail.com", "linda-password-1", false) // registered by password, not yet verified
	b := newBrowser(t, app)
	b.loginWithPassword("linda@gmail.com", "linda-password-1")
	assist.Equal(t, false, b.me()["google_linked"].(bool))
	assist.Equal(t, false, b.me()["email_verified"].(bool))

	// "Connect Google" on the account page.
	app.Google.Grant("code-link", outbound.OAuthUserInfo{ProviderUserID: "g-linda", Email: "linda@gmail.com", EmailVerified: true})
	state := b.startGoogle("/api/v1/auth/google/link", webOrigin+"/account")
	rec := b.backFromGoogle("code-link", state)

	expect(t, rec, http.StatusFound)
	assist.Equal(t, "linked", locationQuery(t, rec).Get("google"))
	me := b.me()
	assist.Equal(t, true, me["google_linked"].(bool))
	assist.Equal(t, true, me["has_password"].(bool))
	// The matching Google address also proves the person owns the inbox, so the account is now verified.
	assist.Equal(t, true, me["email_verified"].(bool))

	// From a fresh browser, "Continue with Google" now signs in to the SAME account.
	other := newBrowser(t, app)
	app.Google.Grant("code-again", outbound.OAuthUserInfo{ProviderUserID: "g-linda", Email: "linda@gmail.com", EmailVerified: true})
	rec = other.backFromGoogle("code-again", other.startGoogle("/api/v1/auth/google", returnToCallback))
	expect(t, rec, http.StatusFound)
	assist.Equal(t, returnToCallback, rec.Header().Get("Location"))
	assist.Equal(t, id.String(), other.me()["id"].(string))
	assist.Equal(t, 1, countUsersWithEmail(t, app, "linda@gmail.com"))

	// Linking again is harmless.
	again := b.backFromGoogle("code-link", b.startGoogle("/api/v1/auth/google/link", webOrigin+"/account"))
	assist.Equal(t, "linked", locationQuery(t, again).Get("google"))
}

func TestHTTP_linkGoogle_aDifferentGoogleAddressCanBeLinkedButDoesNotVerifyTheAccount_integration(t *testing.T) {
	app := integration.NewApp(t)
	insertPasswordUser(t, app, "work@company.org", "work-password-1", false)
	b := newBrowser(t, app)
	b.loginWithPassword("work@company.org", "work-password-1")

	app.Google.Grant("code-diff", outbound.OAuthUserInfo{ProviderUserID: "g-diff", Email: "personal@gmail.com", EmailVerified: true})
	rec := b.backFromGoogle("code-diff", b.startGoogle("/api/v1/auth/google/link", webOrigin+"/account"))

	assist.Equal(t, "linked", locationQuery(t, rec).Get("google"))
	me := b.me()
	assist.Equal(t, true, me["google_linked"].(bool))
	assist.Equal(t, false, me["email_verified"].(bool)) // a different address proves nothing about this inbox
}

func TestHTTP_linkGoogle_failuresSayWhy_integration(t *testing.T) {
	app := integration.NewApp(t)
	insertPasswordUser(t, app, "one@gmail.com", "one-password-123", true)
	insertPasswordUser(t, app, "two@gmail.com", "two-password-123", true)
	one := newBrowser(t, app)
	one.loginWithPassword("one@gmail.com", "one-password-123")
	two := newBrowser(t, app)
	two.loginWithPassword("two@gmail.com", "two-password-123")

	reason := func(rec *httptest.ResponseRecorder) string {
		expect(t, rec, http.StatusFound)
		q := locationQuery(t, rec)
		assert := q.Get("google")
		if assert != "error" {
			t.Fatalf("want google=error, got %q (%s)", assert, rec.Header().Get("Location"))
		}
		return q.Get("reason")
	}

	// Google says the address is unverified.
	app.Google.Grant("c-unv", outbound.OAuthUserInfo{ProviderUserID: "g-u", Email: "one@gmail.com", EmailVerified: false})
	assist.Equal(t, "not_verified", reason(one.backFromGoogle("c-unv", one.startGoogle("/api/v1/auth/google/link", webOrigin+"/account"))))

	// "One" links a Google account; "Two" then tries to link the same one.
	app.Google.Grant("c-one", outbound.OAuthUserInfo{ProviderUserID: "g-shared", Email: "one@gmail.com", EmailVerified: true})
	assist.Equal(t, "linked", locationQuery(t, one.backFromGoogle("c-one", one.startGoogle("/api/v1/auth/google/link", webOrigin+"/account"))).Get("google"))
	app.Google.Grant("c-two", outbound.OAuthUserInfo{ProviderUserID: "g-shared", Email: "one@gmail.com", EmailVerified: true})
	assist.Equal(t, "already_linked", reason(two.backFromGoogle("c-two", two.startGoogle("/api/v1/auth/google/link", webOrigin+"/account"))))

	// "One" already has a Google account linked, so a different one is refused.
	app.Google.Grant("c-other", outbound.OAuthUserInfo{ProviderUserID: "g-another", Email: "one@gmail.com", EmailVerified: true})
	assist.Equal(t, "already_linked", reason(one.backFromGoogle("c-other", one.startGoogle("/api/v1/auth/google/link", webOrigin+"/account"))))

	// Starting a link needs a session, and a link callback replayed in someone else's session is refused.
	stranger := newBrowser(t, app)
	expect(t, stranger.get("/api/v1/auth/google/link?return_to="+url.QueryEscape(webOrigin+"/account")), http.StatusUnauthorized)
	linkState := two.startGoogle("/api/v1/auth/google/link", webOrigin+"/account")
	app.Google.Grant("c-replay", outbound.OAuthUserInfo{ProviderUserID: "g-replay", Email: "two@gmail.com", EmailVerified: true})
	assist.Equal(t, "not_verified", reason(one.backFromGoogle("c-replay", linkState)))
	expect(t, stranger.backFromGoogle("c-replay", linkState), http.StatusUnauthorized) // no session at all
}

// --- disconnecting ------------------------------------------------------------------------------------

func TestHTTP_unlinkGoogle_needsAPasswordSoNobodyIsLockedOut_integration(t *testing.T) {
	app := integration.NewApp(t)

	// A password user can disconnect, and Google then no longer signs in to the account.
	insertPasswordUser(t, app, "pat@gmail.com", "pat-password-123", true)
	b := newBrowser(t, app)
	b.loginWithPassword("pat@gmail.com", "pat-password-123")
	app.Google.Grant("c-pat", outbound.OAuthUserInfo{ProviderUserID: "g-pat", Email: "pat@gmail.com", EmailVerified: true})
	b.backFromGoogle("c-pat", b.startGoogle("/api/v1/auth/google/link", webOrigin+"/account"))
	assist.Equal(t, true, b.me()["google_linked"].(bool))

	expect(t, b.send(http.MethodDelete, "/api/v1/auth/google/link", ""), http.StatusNoContent)
	assist.Equal(t, false, b.me()["google_linked"].(bool))
	expect(t, b.send(http.MethodDelete, "/api/v1/auth/google/link", ""), http.StatusNotFound)
	fresh := newBrowser(t, app)
	app.Google.Grant("c-pat2", outbound.OAuthUserInfo{ProviderUserID: "g-pat", Email: "pat@gmail.com", EmailVerified: true})
	rec := fresh.backFromGoogle("c-pat2", fresh.startGoogle("/api/v1/auth/google", returnToCallback))
	assist.Equal(t, "account_exists", locationQuery(t, rec).Get("error")) // back to the password-account rule

	// Someone who only has Google has no password to fall back on, so it cannot be disconnected.
	googleOnly := newBrowser(t, app)
	app.Google.Grant("c-go", outbound.OAuthUserInfo{ProviderUserID: "g-go", Email: "googleonly@gmail.com", EmailVerified: true})
	googleOnly.backFromGoogle("c-go", googleOnly.startGoogle("/api/v1/auth/google", returnToCallback))
	rec = googleOnly.send(http.MethodDelete, "/api/v1/auth/google/link", "")
	expect(t, rec, http.StatusConflict)
	assist.Contains(t, rec.Body.String(), "set a password")
	assist.Equal(t, true, googleOnly.me()["google_linked"].(bool))
}
