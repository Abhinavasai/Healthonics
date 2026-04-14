package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestAppointmentComments_List_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/appointments/x/comments", nil)
	c.Params = gin.Params{{Key: "id", Value: "x"}}
	c.Set("claims", &Claims{UserID: uuid.New(), Email: "a@b.com", Role: "patient"})

	h := NewAppointmentCommentsHandler()
	h.List(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
