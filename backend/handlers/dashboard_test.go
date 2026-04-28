package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestDashboard_PatientSummary_ForbiddenForDoctor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/patient/dashboard/summary", nil)
	c.Set("claims", &Claims{UserID: uuid.New(), Email: "d@b.com", Role: "doctor"})

	h := NewDashboardHandler()
	h.PatientSummary(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestBuildPatientAlerts_PendingUnread(t *testing.T) {
	a := buildPatientAlerts(1, 2, 5)
	if len(a) < 2 {
		t.Fatalf("expected pending + unread alerts, got %d", len(a))
	}
}

func TestBuildDoctorAlerts_HeavyDay(t *testing.T) {
	a := buildDoctorAlerts(10, 0, 0)
	found := false
	for _, x := range a {
		if x["code"] == "heavy_clinic_day" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected heavy_clinic_day alert")
	}
}
