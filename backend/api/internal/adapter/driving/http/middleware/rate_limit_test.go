package middleware

import (
	"net/http"
	"net/http/httptest"
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
