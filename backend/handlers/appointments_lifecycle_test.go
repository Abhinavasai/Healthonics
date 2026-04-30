package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestPatientCancel_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/patient/appointments/"+uuid.New().String()+"/cancel", nil)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}

	h := NewAppointmentHandler()
	h.PatientCancel(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestCreate_RejectsPastScheduledAt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"doctor_id":"` + uuid.New().String() + `","scheduled_at":"` + time.Now().UTC().Add(-1*time.Hour).Format(time.RFC3339) + `","reason":"follow up"}`
	c.Request = httptest.NewRequest("POST", "/api/appointments", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("claims", &Claims{UserID: uuid.New(), Role: "patient"})

	h := NewAppointmentHandler()
	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
