package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
