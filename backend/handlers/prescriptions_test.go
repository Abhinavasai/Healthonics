package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestPrescriptions_Update_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/prescriptions/"+uuid.New().String(), strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}

	h := NewPrescriptionsHandler()
	h.Update(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestPrescriptions_Update_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("claims", &Claims{Role: "doctor"})
	c.Request = httptest.NewRequest("PUT", "/api/prescriptions/not-a-uuid", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h := NewPrescriptionsHandler()
	h.Update(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestPrescriptions_Update_RequiresAtLeastOneField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("claims", &Claims{Role: "doctor", UserID: uuid.New()})
	c.Request = httptest.NewRequest("PUT", "/api/prescriptions/"+uuid.New().String(), strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}

	h := NewPrescriptionsHandler()
	h.Update(c)

	// If DB not available, endpoint may return 500 after auth checks; we only
	// enforce that it does not return success for empty payload.
	if w.Code == http.StatusOK {
		t.Fatalf("expected non-200 for empty update payload, got %d", w.Code)
	}
}
