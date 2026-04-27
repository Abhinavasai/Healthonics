package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNotifications_ListMine_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/notifications", nil)

	h := NewNotificationsHandler()
	h.ListMine(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestNotifications_ListPreferences_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/notifications/preferences", nil)

	h := NewNotificationsHandler()
	h.ListPreferences(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestNotifications_UpsertPreferences_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/notifications/preferences", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := NewNotificationsHandler()
	h.UpsertPreferences(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestNormalizeNotificationPreference_ValidAndInvalid(t *testing.T) {
	ok, err := normalizeNotificationPreference(NotificationPreference{
		Category:     "Appointment_Reminders",
		Enabled:      true,
		EmailEnabled: true,
		SmsEnabled:   false,
		InAppEnabled: true,
	})
	if err != nil {
		t.Fatalf("expected valid pref, got error: %v", err)
	}
	if ok.Category != "appointment_reminders" {
		t.Fatalf("expected normalized category appointment_reminders, got %s", ok.Category)
	}

	_, err = normalizeNotificationPreference(NotificationPreference{Category: "unknown"})
	if err == nil {
		t.Fatal("expected invalid category error")
	}
}

func TestNotifications_AdminList_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/admin/notifications", nil)

	h := NewNotificationsHandler()
	h.AdminList(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestNotifications_AdminSummary_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/admin/notifications/summary", nil)

	h := NewNotificationsHandler()
	h.AdminSummary(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestNotifications_RetryFailed_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "d2517c5c-c9aa-40d2-b5ce-60d5bcb3f978"}}
	c.Request = httptest.NewRequest("POST", "/api/admin/notifications/d2517c5c-c9aa-40d2-b5ce-60d5bcb3f978/retry", nil)

	h := NewNotificationsHandler()
	h.RetryFailed(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestNotifications_RetryFailed_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("claims", &Claims{})
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}
	c.Request = httptest.NewRequest("POST", "/api/admin/notifications/not-a-uuid/retry", strings.NewReader(""))

	h := NewNotificationsHandler()
	h.RetryFailed(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
