package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupRequestIDRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/ping", func(c *gin.Context) {
		id, _ := c.Get(RequestIDKey)
		c.JSON(http.StatusOK, gin.H{"id": id})
	})
	return r
}

func TestRequestID_GeneratesWhenAbsent(t *testing.T) {
	r := setupRequestIDRouter()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	id := w.Header().Get(RequestIDHeader)
	if id == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}
}

func TestRequestID_PropagatesExisting(t *testing.T) {
	r := setupRequestIDRouter()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(RequestIDHeader, "my-custom-id-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	id := w.Header().Get(RequestIDHeader)
	if id != "my-custom-id-123" {
		t.Fatalf("expected propagated id, got %q", id)
	}
}

func TestRequestID_UniquePerRequest(t *testing.T) {
	r := setupRequestIDRouter()
	ids := map[string]bool{}
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		id := w.Header().Get(RequestIDHeader)
		if ids[id] {
			t.Fatalf("duplicate request ID generated: %q", id)
		}
		ids[id] = true
	}
}
