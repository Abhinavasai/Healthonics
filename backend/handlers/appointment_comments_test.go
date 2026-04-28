package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestResolveCommentVisibility_DefaultPatientVisible(t *testing.T) {
	v, err := resolveCommentVisibility("patient", "")
	if err != nil || v != CommentVisibilityPatientVisible {
		t.Fatalf("got %q err=%v", v, err)
	}
}

func TestResolveCommentVisibility_PatientCannotUseInternal(t *testing.T) {
	_, err := resolveCommentVisibility("patient", "internal")
	if !errors.Is(err, errPatientInternalCommentForbidden) {
		t.Fatalf("expected forbidden: %v", err)
	}
}

func TestResolveCommentVisibility_DoctorInternalOK(t *testing.T) {
	v, err := resolveCommentVisibility("doctor", "internal")
	if err != nil || v != CommentVisibilityInternal {
		t.Fatalf("got %q err=%v", v, err)
	}
}

func TestResolveCommentVisibility_AdminPatientVisibleAlias(t *testing.T) {
	v, err := resolveCommentVisibility("admin", "patient_visible")
	if err != nil || v != CommentVisibilityPatientVisible {
		t.Fatalf("got %q err=%v", v, err)
	}
}

func TestResolveCommentVisibility_Invalid(t *testing.T) {
	_, err := resolveCommentVisibility("doctor", "secret")
	if !errors.Is(err, errInvalidAppointmentCommentVisibility) {
		t.Fatalf("expected invalid: %v", err)
	}
}

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
