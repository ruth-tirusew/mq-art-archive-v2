package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimiterRejectsBeyondLimitAndResets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Unix(100, 0)
	limiter := NewRateLimiter(2, time.Minute)
	limiter.now = func() time.Time { return now }
	r := gin.New()
	r.Use(limiter.Middleware())
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	for i, want := range []int{200, 200, 429} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.0.2.1:1234"
		r.ServeHTTP(w, req)
		if w.Code != want {
			t.Fatalf("request %d: got %d want %d", i, w.Code, want)
		}
	}
	now = now.Add(time.Minute)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("after reset: got %d", w.Code)
	}
}

func limitedRouter(t *testing.T, trusted []string, mw gin.HandlerFunc) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if err := r.SetTrustedProxies(trusted); err != nil {
		t.Fatal(err)
	}
	r.POST("/x", mw, func(c *gin.Context) { c.Status(http.StatusNoContent) })
	return r
}

func hit(r *gin.Engine, remote, xff, body string) int {
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(body))
	req.RemoteAddr = remote
	if xff != "" {
		req.Header.Set("X-Forwarded-For", xff)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestRateLimiter_eachRealClientBehindATrustedProxyGetsItsOwnBudget(t *testing.T) {
	r := limitedRouter(t, []string{"10.0.0.0/8"}, NewRateLimiter(2, time.Minute).Middleware())
	// Two visitors arrive through the same platform proxy (10.0.0.5).
	for i := 0; i < 2; i++ {
		if got := hit(r, "10.0.0.5:1111", "203.0.113.1", ""); got != http.StatusNoContent {
			t.Fatalf("visitor A request %d: got %d", i, got)
		}
	}
	if got := hit(r, "10.0.0.5:1111", "203.0.113.1", ""); got != http.StatusTooManyRequests {
		t.Fatalf("visitor A over limit: got %d", got)
	}
	// Visitor B is not punished for A's traffic.
	if got := hit(r, "10.0.0.5:1111", "203.0.113.2", ""); got != http.StatusNoContent {
		t.Fatalf("visitor B should be unaffected: got %d", got)
	}
}

func TestRateLimiter_forgedForwardedHeaderFromTheInternetIsIgnored(t *testing.T) {
	r := limitedRouter(t, []string{"10.0.0.0/8"}, NewRateLimiter(1, time.Minute).Middleware())
	// A public client connects directly and tries to look like a different address each time.
	if got := hit(r, "198.51.100.9:1", "1.1.1.1", ""); got != http.StatusNoContent {
		t.Fatalf("first: %d", got)
	}
	if got := hit(r, "198.51.100.9:1", "2.2.2.2", ""); got != http.StatusTooManyRequests {
		t.Fatalf("a spoofed X-Forwarded-For must not reset the budget: got %d", got)
	}
}

func TestRateLimiter_perEmailAcrossManyIPs(t *testing.T) {
	r := limitedRouter(t, nil, NewRateLimiter(2, time.Minute).ByJSONField("email"))
	body := `{"email":"Victim@Example.com","password":"x"}`
	for i, ip := range []string{"1.1.1.1:1", "2.2.2.2:1"} {
		if got := hit(r, ip, "", body); got != http.StatusNoContent {
			t.Fatalf("attempt %d: %d", i, got)
		}
	}
	if got := hit(r, "3.3.3.3:1", "", `{"email":"victim@example.com"}`); got != http.StatusTooManyRequests {
		t.Fatalf("third IP, same account (case-insensitive): got %d", got)
	}
	if got := hit(r, "3.3.3.3:1", "", `{"email":"other@example.com"}`); got != http.StatusNoContent {
		t.Fatalf("a different account is unaffected: got %d", got)
	}
}

func TestRateLimiter_prunesExpiredWindows(t *testing.T) {
	l := NewRateLimiter(1, time.Minute)
	now := time.Now()
	l.now = func() time.Time { return now }
	for i := 0; i < 50; i++ {
		l.allow(fmt.Sprintf("k%d", i))
	}
	now = now.Add(2 * time.Minute)
	l.allow("fresh")
	if len(l.clients) != 1 {
		t.Fatalf("expected expired entries pruned, have %d", len(l.clients))
	}
}
