package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestPrescriptions_Revoke_RequiresGuardrailFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.NewString()}}
	c.Request = httptest.NewRequest("PATCH", "/api/prescriptions/id/revoke", strings.NewReader(`{"reason":"tiny","confirm":"no"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("claims", &Claims{UserID: uuid.New(), Role: "doctor"})

	h := NewPrescriptionsHandler()
	h.Revoke(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
