package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestPrescriptions_List_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/patients/"+uuid.New().String()+"/prescriptions", nil)
	c.Params = gin.Params{{Key: "patientId", Value: uuid.New().String()}}

	h := NewPrescriptionsHandler()
	h.ListByPatient(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
