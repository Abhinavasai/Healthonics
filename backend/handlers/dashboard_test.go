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
