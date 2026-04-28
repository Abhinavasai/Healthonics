package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestAdminUserLifecycle_GetSettingsKpis_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/admin/user-lifecycle/settings-kpis", nil)

	h := NewAdminUserLifecycleHandler()
	h.GetSettingsKpis(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAdminUserLifecycle_UpdateSettings_BadRange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/admin/user-lifecycle/settings", strings.NewReader(`{"new_user_window_days":0,"inactive_window_days":30}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("claims", &Claims{UserID: uuid.New(), Role: "admin"})

	h := NewAdminUserLifecycleHandler()
	h.UpdateSettings(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
