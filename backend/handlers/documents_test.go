package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestDocuments_List_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/documents", nil)

	h := NewDocumentsHandler()
	h.List(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestDocuments_Upload_MissingFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("claims", &Claims{
		UserID: uuid.New(),
		Email:  "p@test.com",
		Role:   "patient",
	})
	c.Request = httptest.NewRequest("POST", "/api/documents", nil)

	h := NewDocumentsHandler()
	h.Upload(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDocuments_Download_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/documents/x/download", nil)
	c.Params = gin.Params{{Key: "id", Value: "not-uuid"}}
	c.Set("claims", &Claims{
		UserID: uuid.New(),
		Email:  "p@test.com",
		Role:   "patient",
	})

	h := NewDocumentsHandler()
	h.Download(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSanitizeFilename(t *testing.T) {
	if sanitizeFilename("../../etc/passwd") != "passwd" {
		t.Fatalf("expected basename only")
	}
	if sanitizeFilename("") != "" {
		t.Fatalf("empty")
	}
}

func TestSniffContentType(t *testing.T) {
	if sniffContentType("x.pdf", nil) != "application/pdf" {
		t.Fatalf("pdf ext")
	}
	if sniffContentType("x.png", []byte{0x89, 'P', 'N', 'G'}) != "image/png" {
		t.Fatalf("png ext")
	}
}
