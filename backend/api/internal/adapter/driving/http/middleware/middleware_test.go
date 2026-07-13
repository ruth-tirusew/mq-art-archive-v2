package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/testutil/assist"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRequestID_generatesWhenMissing(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/", func(c *gin.Context) {
		id, ok := c.Get("request_id")
		if !ok {
			t.Fatal("request_id not set")
		}
		if id == "" {
			t.Fatal("request_id is empty")
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assist.Equal(t, http.StatusOK, w.Code)
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID response header")
	}
}

func TestRequestID_preservesExisting(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "fixed-id")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assist.Equal(t, "fixed-id", w.Header().Get("X-Request-ID"))
}

func TestCORS_allowsConfiguredOrigin(t *testing.T) {
	r := gin.New()
	r.Use(CORS([]string{"http://localhost:5173"}))
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assist.Equal(t, "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_optionsPreflight(t *testing.T) {
	r := gin.New()
	r.Use(CORS([]string{"http://localhost:5173"}))
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assist.Equal(t, http.StatusNoContent, w.Code)
}

func TestRequireRole_setsContext(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(requestauth.ContextUserRole, identity.RoleAdmin)
		c.Next()
	}, RequireRole("admin"))
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	assist.Equal(t, http.StatusOK, w.Code)
}
