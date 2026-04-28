package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestContextualComments_List_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "contextType", Value: "record"}, {Key: "contextId", Value: uuid.NewString()}}
	c.Request = httptest.NewRequest("GET", "/api/comments/record/id", nil)

	h := NewContextualCommentsHandler()
	h.List(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestContextualComments_List_BadContextType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("claims", &Claims{UserID: uuid.New(), Role: "admin"})
	c.Params = gin.Params{{Key: "contextType", Value: "bad"}, {Key: "contextId", Value: uuid.NewString()}}
	c.Request = httptest.NewRequest("GET", "/api/comments/bad/id", nil)

	h := NewContextualCommentsHandler()
	h.List(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestContextualComments_Create_BadBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("claims", &Claims{UserID: uuid.New(), Role: "admin"})
	c.Params = gin.Params{{Key: "contextType", Value: "record"}, {Key: "contextId", Value: uuid.NewString()}}
	c.Request = httptest.NewRequest("POST", "/api/comments/record/id", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := NewContextualCommentsHandler()
	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestContextualComments_Create_PatientInternalRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("claims", &Claims{UserID: uuid.New(), Role: "patient"})
	c.Params = gin.Params{{Key: "contextType", Value: "record"}, {Key: "contextId", Value: uuid.NewString()}}
	c.Request = httptest.NewRequest("POST", "/api/comments/record/id", strings.NewReader(`{"body":"x","visibility":"internal"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := NewContextualCommentsHandler()
	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
