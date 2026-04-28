package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestAdminAIRuntime_GetObservability_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/admin/ai/observability", nil)

	h := NewAdminAIRuntimeHandler()
	h.GetObservability(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAdminAIRuntime_UpdateSettings_BadRange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/admin/ai/settings", strings.NewReader(`{"rate_limit_per_minute":0,"cache_ttl_seconds":900,"ollama_model":"llama3.1:8b"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("claims", &Claims{UserID: uuid.New(), Role: "admin"})

	h := NewAdminAIRuntimeHandler()
	h.UpdateSettings(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
