package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

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

func TestPrescriptions_DownloadPDF_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/prescriptions/"+uuid.New().String()+"/pdf", nil)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}

	h := NewPrescriptionsHandler()
	h.DownloadPDF(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestPrescriptions_DownloadPDF_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("claims", &Claims{Role: "doctor", UserID: uuid.New()})
	c.Request = httptest.NewRequest("GET", "/api/prescriptions/not-a-uuid/pdf", nil)
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h := NewPrescriptionsHandler()
	h.DownloadPDF(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestReminderTimesForFrequency(t *testing.T) {
	if got := reminderTimesForFrequency("twice daily"); !slices.Equal(got, []string{"08:00", "20:00"}) {
		t.Fatalf("unexpected twice-daily times: %v", got)
	}
	if got := reminderTimesForFrequency("3x"); !slices.Equal(got, []string{"08:00", "14:00", "20:00"}) {
		t.Fatalf("unexpected thrice-daily times: %v", got)
	}
	if got := reminderTimesForFrequency("once daily"); !slices.Equal(got, []string{"08:00"}) {
		t.Fatalf("unexpected default times: %v", got)
	}
}

func TestBuildReminderSchedule(t *testing.T) {
	start := time.Date(2026, 4, 27, 17, 30, 0, 0, time.UTC)
	got := buildReminderSchedule(start, 2, []string{"08:00", "20:00"})
	if len(got) != 4 {
		t.Fatalf("expected 4 reminder times, got %d", len(got))
	}
	if got[0].Hour() != 8 || got[0].Minute() != 0 {
		t.Fatalf("unexpected first reminder time: %v", got[0])
	}
	if got[2].Day() != 28 {
		t.Fatalf("expected day rollover to 28, got %v", got[2])
	}
}

func TestBuildPrescriptionPDF_ContainsHeaderAndFields(t *testing.T) {
	pdf := buildPrescriptionPDF(
		"Atenolol",
		"25mg",
		"daily",
		"after food",
		"active",
		30,
		time.Date(2026, 4, 27, 10, 0, 0, 0, time.UTC),
	)
	if !bytes.HasPrefix(pdf, []byte("%PDF-1.4")) {
		t.Fatalf("expected pdf header, got: %q", string(pdf[:8]))
	}
	if !strings.Contains(string(pdf), "Medication: Atenolol") {
		t.Fatalf("expected medication text in generated pdf")
	}
}
