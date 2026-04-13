package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestListMyOpenSlots_ForbiddenForPatient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/doctor/slots", nil)
	c.Set("claims", &Claims{
		UserID: uuid.New(),
		Email:  "pat@test.local",
		Role:   "patient",
	})

	h := NewGeoBookingHandler()
	h.ListMyOpenSlots(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", w.Code, w.Body.String())
	}
}

