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
