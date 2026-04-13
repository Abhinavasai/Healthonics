package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/healthonyx/backend/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// TestAppointmentActivity_flow exercises PATCH status → GET activity against a real DB.
// Enable with: RUN_APPOINTMENT_ACTIVITY_INTEGRATION=1 DATABASE_URL=postgres://... go test ./handlers -run TestAppointmentActivity_flow -count=1
func TestAppointmentActivity_flow(t *testing.T) {
	if os.Getenv("RUN_APPOINTMENT_ACTIVITY_INTEGRATION") != "1" {
		t.Skip("set RUN_APPOINTMENT_ACTIVITY_INTEGRATION=1 and DATABASE_URL to run")
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL required")
	}

	gin.SetMode(gin.TestMode)
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	old := db.Pool
	db.Pool = pool
	defer func() { db.Pool = old }()

	if err := db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	suffix := uuid.New().String()[:8]
	patientEmail := "pat-act-" + suffix + "@test.local"
	doctorEmail := "doc-act-" + suffix + "@test.local"

	ph, _ := bcrypt.GenerateFromPassword([]byte("testpass123"), bcrypt.DefaultCost)
	var patientID, doctorID uuid.UUID
	err = pool.QueryRow(ctx, `INSERT INTO users (email, password_hash, role) VALUES ($1, $2, 'patient') RETURNING id`,
		patientEmail, string(ph)).Scan(&patientID)
	if err != nil {
		t.Fatal(err)
	}
	err = pool.QueryRow(ctx, `INSERT INTO users (email, password_hash, role) VALUES ($1, $2, 'doctor') RETURNING id`,
		doctorEmail, string(ph)).Scan(&doctorID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM appointment_activities WHERE appointment_id IN (SELECT id FROM appointments WHERE patient_id = $1)`, patientID)
		_, _ = pool.Exec(ctx, `DELETE FROM appointments WHERE patient_id = $1`, patientID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id IN ($1, $2)`, patientID, doctorID)
	})

	when := time.Date(2026, 7, 1, 15, 0, 0, 0, time.UTC)
	var apptID uuid.UUID
	err = pool.QueryRow(ctx, `
		INSERT INTO appointments (patient_id, doctor_id, scheduled_at, reason, status)
		VALUES ($1, $2, $3, $4, 'pending')
		RETURNING id
	`, patientID, doctorID, when, "integration activity").Scan(&apptID)
	if err != nil {
		t.Fatal(err)
	}

	secret := []byte("test-secret-for-activity-flow")
	h := NewAuthHandler(string(secret))
	apptHandler := NewAppointmentHandler()

	token := func(uid uuid.UUID, email, role string) string {
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
			UserID: uid,
			Email:  email,
			Role:   role,
		})
		s, err := tok.SignedString(secret)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}

	r := gin.New()
	r.PATCH("/api/appointments/:id/status", h.RequireAuth(), h.RequireRole("doctor"), apptHandler.UpdateStatus)
	r.GET("/api/appointments/:id/activity", h.RequireAuth(), apptHandler.ListActivity)

	doctorTok := token(doctorID, doctorEmail, "doctor")

	body := []byte(`{"status":"approved"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/appointments/"+apptID.String()+"/status", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+doctorTok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("PATCH status: %d %s", w.Code, w.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/appointments/"+apptID.String()+"/activity", nil)
	req2.Header.Set("Authorization", "Bearer "+doctorTok)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("GET activity: %d %s", w2.Code, w2.Body.String())
	}
	var resp struct {
		Activities []models.AppointmentActivity `json:"activities"`
	}
	if err := json.Unmarshal(w2.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Activities) != 1 {
		t.Fatalf("activities len: got %d body=%s", len(resp.Activities), w2.Body.String())
	}
	if resp.Activities[0].Action != "status_changed" || resp.Activities[0].Detail != "approved" {
		t.Fatalf("unexpected row: %+v", resp.Activities[0])
	}
	if resp.Activities[0].ActorEmail != doctorEmail {
		t.Fatalf("actor_email: got %q want %q", resp.Activities[0].ActorEmail, doctorEmail)
	}

	patientTok := token(patientID, patientEmail, "patient")
	req3 := httptest.NewRequest(http.MethodGet, "/api/appointments/"+apptID.String()+"/activity", nil)
	req3.Header.Set("Authorization", "Bearer "+patientTok)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("patient GET activity: %d %s", w3.Code, w3.Body.String())
	}
}
