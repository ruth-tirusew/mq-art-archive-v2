package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func csrfRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CSRF([]string{"https://app.example.com"}))
	ok := func(c *gin.Context) { c.Status(http.StatusNoContent) }
	r.GET("/r", ok)
	r.POST("/w", ok)
	r.PATCH("/w", ok)
	r.DELETE("/w", ok)
	return r
}

func send(method, body string, headers map[string]string) int {
	var rd *strings.Reader
	if body != "" {
		rd = strings.NewReader(body)
	} else {
		rd = strings.NewReader("")
	}
	path := "/w"
	if method == http.MethodGet {
		path = "/r"
	}
	req := httptest.NewRequest(method, path, rd)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	csrfRouter().ServeHTTP(w, req)
	return w.Code
}

func TestCSRF(t *testing.T) {
	json := map[string]string{"Content-Type": "application/json"}
	cases := []struct {
		name    string
		method  string
		body    string
		headers map[string]string
		want    int
	}{
		{"allowed origin with json", http.MethodPost, `{"a":1}`, map[string]string{"Origin": "https://app.example.com", "Content-Type": "application/json"}, http.StatusNoContent},
		{"json with charset", http.MethodPost, `{"a":1}`, map[string]string{"Origin": "https://app.example.com", "Content-Type": "application/json; charset=utf-8"}, http.StatusNoContent},
		{"foreign origin blocked", http.MethodPost, `{"a":1}`, map[string]string{"Origin": "https://evil.example", "Content-Type": "application/json"}, http.StatusForbidden},
		{"null origin blocked", http.MethodPost, `{"a":1}`, map[string]string{"Origin": "null", "Content-Type": "application/json"}, http.StatusForbidden},
		{"foreign origin blocked on delete", http.MethodDelete, "", map[string]string{"Origin": "https://evil.example"}, http.StatusForbidden},
		{"foreign origin blocked on patch", http.MethodPatch, `{"a":1}`, map[string]string{"Origin": "https://evil.example", "Content-Type": "application/json"}, http.StatusForbidden},
		{"no origin is not a browser csrf (curl, server-to-server)", http.MethodPost, `{"a":1}`, json, http.StatusNoContent},
		{"text/plain body is a simple request: rejected", http.MethodPost, `{"a":1}`, map[string]string{"Origin": "https://app.example.com", "Content-Type": "text/plain"}, http.StatusUnsupportedMediaType},
		{"form body rejected", http.MethodPost, "a=1", map[string]string{"Content-Type": "application/x-www-form-urlencoded"}, http.StatusUnsupportedMediaType},
		{"missing content type with body rejected", http.MethodPost, `{"a":1}`, nil, http.StatusUnsupportedMediaType},
		{"empty-body post needs no content type", http.MethodPost, "", map[string]string{"Origin": "https://app.example.com"}, http.StatusNoContent},
		{"reads are never blocked", http.MethodGet, "", map[string]string{"Origin": "https://evil.example"}, http.StatusNoContent},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := send(tc.method, tc.body, tc.headers); got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}
