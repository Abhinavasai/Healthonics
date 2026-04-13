package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/healthonyx/backend/config"
	"github.com/healthonyx/backend/db"
)

// Seed deterministic test rows for local/manual QA.
// Safe: intended only for local DB; it only INSERTs with fixed UUIDs (idempotent via ON CONFLICT DO NOTHING).
func main() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if err := db.Connect(cfg.DatabaseURL); err != nil {
		log.Fatalf("database connection: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		log.Fatalf("migration: %v", err)
	}

	// These UUIDs are stable so you can paste them in Postman/Cypress.
	patientID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	doctorID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	adminID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	threadID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	msgID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	apptID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	actID := uuid.MustParse("77777777-7777-7777-7777-777777777777")

	// Users are created by cmd/seed already; here we just ensure IDs exist for quick referencing.
	// If emails already exist, we don't attempt to overwrite them (avoid breaking existing demo creds).
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, role)
		VALUES
			($1, 'patient-fixed@healthonyx.demo', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZagvrTqGQwQxI6pGxQ1vYv7zR1W7a', 'patient'),
			($2, 'doctor-fixed@healthonyx.demo',  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZagvrTqGQwQxI6pGxQ1vYv7zR1W7a', 'doctor'),
			($3, 'admin-fixed@healthonyx.demo',   '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZagvrTqGQwQxI6pGxQ1vYv7zR1W7a', 'admin')
		ON CONFLICT (id) DO NOTHING
	`, patientID, doctorID, adminID)

	// Appointment + activity
	when := time.Now().UTC().Add(72 * time.Hour).Truncate(time.Minute)
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO appointments (id, patient_id, doctor_id, scheduled_at, reason, status)
		VALUES ($1, $2, $3, $4, $5, 'approved')
		ON CONFLICT (id) DO NOTHING
	`, apptID, patientID, doctorID, when, "Seeded test appointment")

	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO appointment_activities (id, appointment_id, actor_user_id, action, detail, created_at)
		VALUES ($1, $2, $3, 'status_changed', 'approved', NOW())
		ON CONFLICT (id) DO NOTHING
	`, actID, apptID, doctorID)

	// Messaging thread + one message
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO message_threads (id, patient_id, doctor_id, last_message_at, last_preview)
		VALUES ($1, $2, $3, NOW(), 'Hello from seeded thread')
		ON CONFLICT (id) DO NOTHING
	`, threadID, patientID, doctorID)

	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO messages (id, thread_id, sender_id, body, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (id) DO NOTHING
	`, msgID, threadID, patientID, "Hello from seeded message")

	fmt.Println("Seeded deterministic QA IDs:")
	fmt.Printf("  patient_id:      %s\n", patientID)
	fmt.Printf("  doctor_id:       %s\n", doctorID)
	fmt.Printf("  admin_id:        %s\n", adminID)
	fmt.Printf("  appointment_id:  %s\n", apptID)
	fmt.Printf("  activity_id:     %s\n", actID)
	fmt.Printf("  thread_id:       %s\n", threadID)
	fmt.Printf("  message_id:      %s\n", msgID)
	fmt.Println("Note: fixed users use password hash for 'password' (bcrypt demo hash). Prefer cmd/seed demo users for login creds.")
}
