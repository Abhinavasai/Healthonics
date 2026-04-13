package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestPatientFiles_List_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/patients/"+uuid.New().String()+"/files", nil)
	c.Params = gin.Params{{Key: "patientId", Value: uuid.New().String()}}

	h := NewPatientFilesHandler(t.TempDir())
	h.List(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestPatientFiles_Download_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/files/not-uuid", nil)
	c.Params = gin.Params{{Key: "id", Value: "not-uuid"}}
	c.Set("claims", &Claims{
		UserID: uuid.New(),
		Email:  "a@b.com",
		Role:   "patient",
	})

	h := NewPatientFilesHandler(t.TempDir())
	h.Download(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
