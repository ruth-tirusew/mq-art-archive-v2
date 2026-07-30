package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mq/api/internal/testutil/assist"
	"golang.org/x/oauth2"
)

func TestGoogleProvider_Exchange(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		assist.Equal(t, "auth-code", r.Form.Get("code"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"sub":            "google-user-1",
			"email":          "artist@example.com",
			"email_verified": true,
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	p := NewGoogleProvider("client-id", "client-secret", "http://localhost/callback")
	p.config.Endpoint = oauth2.Endpoint{
		AuthURL:  srv.URL + "/auth",
		TokenURL: srv.URL + "/token",
	}
	p.userInfoURL = srv.URL + "/userinfo"

	info, err := p.Exchange(context.Background(), "auth-code")
	assist.NoError(t, err)
	assist.Equal(t, "google-user-1", info.ProviderUserID)
	assist.Equal(t, "artist@example.com", info.Email)
	assist.Equal(t, true, info.EmailVerified)
}

func TestGoogleProvider_Name(t *testing.T) {
	p := NewGoogleProvider("id", "secret", "http://localhost/cb")
	assist.Equal(t, ProviderGoogle, p.Name())
}

func TestGoogleProvider_AuthCodeURL(t *testing.T) {
	p := NewGoogleProvider("id", "secret", "http://localhost/cb")
	url := p.AuthCodeURL("state-xyz")
	assist.Contains(t, url, "state-xyz")
	assist.Contains(t, url, "client_id=id")
}
