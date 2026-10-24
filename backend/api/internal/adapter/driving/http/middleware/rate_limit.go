package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateWindow struct {
	start time.Time
	count int
}

// RateLimiter is a fixed-window limiter. Keys are chosen by the caller (client IP, account email).
type RateLimiter struct {
	mu        sync.Mutex
	limit     int
	window    time.Duration
	clients   map[string]rateWindow
	now       func() time.Time
	lastPrune time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{limit: limit, window: window, clients: make(map[string]rateWindow), now: time.Now}
}

// allow records one hit for key and reports whether it is within the limit.
func (l *RateLimiter) allow(key string) bool {
	now := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()

	// Drop expired windows now and then so the map cannot grow without bound.
	if now.Sub(l.lastPrune) >= l.window {
		for k, w := range l.clients {
			if now.Sub(w.start) >= l.window {
				delete(l.clients, k)
			}
		}
		l.lastPrune = now
	}

	entry := l.clients[key]
	if entry.start.IsZero() || now.Sub(entry.start) >= l.window {
		entry = rateWindow{start: now}
	}
	entry.count++
	l.clients[key] = entry
	return entry.count <= l.limit
}

func deny(c *gin.Context) {
	c.Header("Retry-After", "60")
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
}

// Middleware limits per client IP. c.ClientIP() honours X-Forwarded-For only for the proxies the
// router trusts (see TRUSTED_PROXIES), so behind Render each real visitor gets their own budget
// and a forged header from the open internet cannot dodge or share it.
func (l *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !l.allow(c.ClientIP()) {
			deny(c)
			return
		}
		c.Next()
	}
}

// ByJSONField limits per value of a top-level string field in the JSON body (for example the
// account email on login), regardless of which IPs the requests come from. That is what slows
// credential stuffing from many addresses against one account. The body is restored for the handler.
func (l *RateLimiter) ByJSONField(field string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil {
			c.Next()
			return
		}
		raw, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(raw))

		var body map[string]any
		if json.Unmarshal(raw, &body) == nil {
			if v, ok := body[field].(string); ok && strings.TrimSpace(v) != "" {
				if !l.allow(field + ":" + strings.ToLower(strings.TrimSpace(v))) {
					deny(c)
					return
				}
			}
		}
		c.Next()
	}
}
