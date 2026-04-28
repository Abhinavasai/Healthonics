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

func TestAdminUserLifecycle_ListUsers_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/admin/user-lifecycle/users", nil)

	h := NewAdminUserLifecycleHandler()
	h.ListUsers(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAdminUserLifecycle_CreateUser_BadPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/admin/user-lifecycle/users", strings.NewReader(`{"email":"x","password":"123","role":"nope"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("claims", &Claims{UserID: uuid.New(), Role: "admin"})

	h := NewAdminUserLifecycleHandler()
	h.CreateUser(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAdminUserLifecycle_UpdateUser_RequiresFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.NewString()}}
	c.Request = httptest.NewRequest("PUT", "/api/admin/user-lifecycle/users/id", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("claims", &Claims{UserID: uuid.New(), Role: "admin"})

	h := NewAdminUserLifecycleHandler()
	h.UpdateUser(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAdminUserLifecycle_DeactivateUser_CannotDeactivateSelf(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	adminID := uuid.New()
	c.Params = gin.Params{{Key: "id", Value: adminID.String()}}
	c.Request = httptest.NewRequest("PATCH", "/api/admin/user-lifecycle/users/id/deactivate", nil)
	c.Set("claims", &Claims{UserID: adminID, Role: "admin"})

	h := NewAdminUserLifecycleHandler()
	h.DeactivateUser(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAdminUserLifecycle_ResetPassword_BadPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.NewString()}}
	c.Request = httptest.NewRequest("POST", "/api/admin/user-lifecycle/users/id/reset-password", strings.NewReader(`{"new_password":"123"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("claims", &Claims{UserID: uuid.New(), Role: "admin"})

	h := NewAdminUserLifecycleHandler()
	h.ResetPassword(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
