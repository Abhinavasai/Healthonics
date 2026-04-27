package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestDoctorDocuments_List_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/doctor/documents", nil)

	h := NewDoctorDocumentsHandler()
	h.List(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestDoctorDocuments_List_ForbiddenForPatient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/doctor/documents", nil)
	c.Set("claims", &Claims{UserID: uuid.New(), Email: "p@b.com", Role: "patient"})

	h := NewDoctorDocumentsHandler()
	h.List(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestDoctorDocuments_Get_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/doctor/documents/"+uuid.New().String(), nil)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}

	h := NewDoctorDocumentsHandler()
	h.Get(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestDoctorDocuments_Summarize_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/doctor/documents/"+uuid.New().String()+"/summarize", nil)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}

	h := NewDoctorDocumentsHandler()
	h.Summarize(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestDoctorDocuments_Summarize_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/doctor/documents/not-a-uuid/summarize", nil)
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}
	c.Set("claims", &Claims{UserID: uuid.New(), Email: "d@b.com", Role: "doctor"})

	h := NewDoctorDocumentsHandler()
	h.Summarize(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
