package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func setupRateLimitRouter(limit int, window time.Duration) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	rl := NewRateLimiter(context.Background(), limit, window)
	r.POST("/login", rl.Middleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	r := setupRateLimitRouter(5, time.Minute)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}
}

func TestRateLimit_BlocksOverLimit(t *testing.T) {
	r := setupRateLimitRouter(3, time.Minute)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = "10.0.0.1:9999"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
	// 4th request must be rejected
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = "10.0.0.1:9999"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on 4th request, got %d", w.Code)
	}
}

func TestRateLimit_DifferentIPsIsolated(t *testing.T) {
	r := setupRateLimitRouter(2, time.Minute)
	// Exhaust IP A
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = "192.168.1.1:80"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
	// IP B should still be allowed
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = "192.168.1.2:80"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected IP B to be allowed, got %d", w.Code)
	}
}

func TestRateLimit_WindowExpiry(t *testing.T) {
	// Very short window so we can test expiry
	r := setupRateLimitRouter(1, 50*time.Millisecond)
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = "5.5.5.5:1"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("first request should pass, got %d", w.Code)
	}
	// Second immediately should be blocked
	req2 := httptest.NewRequest(http.MethodPost, "/login", nil)
	req2.RemoteAddr = "5.5.5.5:1"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request should be blocked, got %d", w2.Code)
	}
	// After window expires, should be allowed again
	time.Sleep(60 * time.Millisecond)
	req3 := httptest.NewRequest(http.MethodPost, "/login", nil)
	req3.RemoteAddr = "5.5.5.5:1"
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("request after window should pass, got %d", w3.Code)
	}
}
