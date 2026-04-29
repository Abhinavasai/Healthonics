package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/healthonyx/backend/config"
	"github.com/healthonyx/backend/db"
	"golang.org/x/crypto/bcrypt"
)

func buildMinimalPDFBytes(text string) []byte {
	// Minimal PDF fixture that passes `extractTextFromPDF` unit checks.
	// NOTE: keep `text` free of parentheses to avoid breaking PDF literal syntax.
	if text == "" {
		text = "empty"
	}
	text = strings.ReplaceAll(text, "(", "")
	text = strings.ReplaceAll(text, ")", "")
	return []byte(fmt.Sprintf(
		`%%PDF-1.4
1 0 obj
<<>>
stream
BT (%s) Tj ET
endstream
endobj
%%EOF`,
		text,
	))
}

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

	hospitals := []struct {
		name, city, region string
		lat, lng           float64
	}{
		{"UF Health Shands Hospital", "Gainesville", "Florida", 29.6406, -82.3444},
		{"Orlando Health", "Orlando", "Florida", 28.5383, -81.3792},
		{"Mayo Clinic", "Jacksonville", "Florida", 30.2849, -81.3961},
	}

	var shandsID uuid.UUID
	for _, h := range hospitals {
		var id uuid.UUID
		err := db.Pool.QueryRow(ctx, `
			INSERT INTO hospitals (name, city, region, latitude, longitude)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (name, region) DO UPDATE SET city = EXCLUDED.city, latitude = EXCLUDED.latitude, longitude = EXCLUDED.longitude
			RETURNING id
		`, h.name, h.city, h.region, h.lat, h.lng).Scan(&id)
		if err != nil {
			log.Fatalf("hospital %s: %v", h.name, err)
		}
		if h.name == "UF Health Shands Hospital" {
			shandsID = id
		}
		fmt.Printf("Hospital: %s — %s, %s\n", h.name, h.city, h.region)
	}

	users := []struct {
		email string
		pass  string
		role  string
	}{
		{"admin@healthonyx.demo", "admin123", "admin"},
		{"doctor@healthonyx.demo", "doctor123", "doctor"},
		{"patient@healthonyx.demo", "patient123", "patient"},
	}

	for _, u := range users {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.pass), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal(err)
		}
		_, err = db.Pool.Exec(ctx,
			`INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3)
			 ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash, role = EXCLUDED.role`,
			u.email, string(hash), u.role,
		)
		if err != nil {
			log.Fatalf("seed %s: %v", u.email, err)
		}
		fmt.Printf("Seeded: %s / %s (%s)\n", u.email, u.pass, u.role)
	}

	if shandsID != uuid.Nil {
		_, err := db.Pool.Exec(ctx, `
			UPDATE users
			SET specialization = $1, hospital_id = $2
			WHERE email = 'doctor@healthonyx.demo' AND role = 'doctor'
		`, "Internal Medicine", shandsID)
		if err != nil {
			log.Fatalf("doctor profile: %v", err)
		}
		fmt.Println("Doctor profile: specialization + hospital linked.")
	}

	ids, err := loadDemoUserIDs(ctx)
	if err != nil {
		log.Fatalf("load demo user ids: %v", err)
	}

	if err := seedClinicalScenarioData(ctx, ids); err != nil {
		log.Fatalf("seed clinical scenario data: %v", err)
	}

	// One demo open slot for the demo doctor (tomorrow 10:00 UTC window — adjust in app as needed)
	var docID uuid.UUID
	err = db.Pool.QueryRow(ctx, `SELECT id FROM users WHERE email = 'doctor@healthonyx.demo' AND role = 'doctor'`).Scan(&docID)
	if err != nil {
		log.Printf("demo slot: doctor id: %v", err)
	} else {
		now := time.Now().UTC()
		baseDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour)

		// Multiple windows to make the UI feel "alive" during manual testing.
		times := []struct{ h, m int }{
			{9, 0},
			{10, 0},
			{14, 0},
			{16, 0},
		}
		for dayOffset := 0; dayOffset < 4; dayOffset++ {
			day := baseDay.AddDate(0, 0, dayOffset)
			for _, t := range times {
				start := time.Date(day.Year(), day.Month(), day.Day(), t.h, t.m, 0, 0, time.UTC)
				end := start.Add(30 * time.Minute)
				_, err = db.Pool.Exec(ctx, `
					INSERT INTO doctor_slots (doctor_id, start_at, end_at)
					VALUES ($1, $2, $3)
					ON CONFLICT (doctor_id, start_at) DO NOTHING
				`, docID, start, end)
				if err != nil {
					log.Printf("demo slot: %v", err)
				}
			}
		}

		// Mark one slot as booked so the doctor list shows both available/unavailable rows.
		bookedDay := baseDay.AddDate(0, 0, 1)
		bookedStart := time.Date(bookedDay.Year(), bookedDay.Month(), bookedDay.Day(), 11, 0, 0, 0, time.UTC)
		bookedEnd := bookedStart.Add(30 * time.Minute)
		_, err = db.Pool.Exec(ctx, `
			INSERT INTO doctor_slots (doctor_id, start_at, end_at, patient_id)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (doctor_id, start_at) DO NOTHING
		`, docID, bookedStart, bookedEnd, ids.patient)
		if err != nil {
			log.Printf("demo booked slot: %v", err)
		}
	}

	fmt.Println("Demo users seeded. Use these credentials to log in.")
}

type demoIDs struct {
	admin   uuid.UUID
	doctor  uuid.UUID
	patient uuid.UUID
}

func loadDemoUserIDs(ctx context.Context) (demoIDs, error) {
	var out demoIDs
	if err := db.Pool.QueryRow(ctx, `SELECT id FROM users WHERE email='admin@healthonyx.demo'`).Scan(&out.admin); err != nil {
		return out, err
	}
	if err := db.Pool.QueryRow(ctx, `SELECT id FROM users WHERE email='doctor@healthonyx.demo'`).Scan(&out.doctor); err != nil {
		return out, err
	}
	if err := db.Pool.QueryRow(ctx, `SELECT id FROM users WHERE email='patient@healthonyx.demo'`).Scan(&out.patient); err != nil {
		return out, err
	}
	return out, nil
}

func seedClinicalScenarioData(ctx context.Context, ids demoIDs) error {
	now := time.Now().UTC()

	// Appointment scenarios: pending + approved + completed.
	var apptID uuid.UUID
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO appointments (patient_id, doctor_id, scheduled_at, reason, status)
		VALUES ($1, $2, $3, $4, 'approved')
		RETURNING id
	`, ids.patient, ids.doctor, now.Add(48*time.Hour), "Follow-up blood pressure review").Scan(&apptID)
	if err != nil {
		return err
	}
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO appointment_activities (appointment_id, actor_user_id, action, detail)
		VALUES ($1, $2, 'status_changed', 'approved')
	`, apptID, ids.doctor)

	// Prescription + reminder + guardrail-related notification scenarios.
	var rxID uuid.UUID
	err = db.Pool.QueryRow(ctx, `
		INSERT INTO prescriptions (patient_id, doctor_id, medication_name, dosage, frequency, duration_days, instructions, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'active')
		RETURNING id
	`, ids.patient, ids.doctor, "Atorvastatin", "20mg", "once daily", 30, "Take after dinner").Scan(&rxID)
	if err != nil {
		return err
	}
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO notifications (user_id, title, body, channel, status, scheduled_for, provider, attempts, last_error)
		VALUES ($1, $2, $3, 'email', 'pending', $4, 'sendgrid', 0, '')
	`, ids.patient, "Medication reminder", "Take Atorvastatin 20mg tonight after dinner.", now.Add(36*time.Hour))

	// Additional reminder + alert items for inbox realism.
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO notifications (user_id, title, body, channel, status, scheduled_for, provider, attempts, last_error)
		VALUES ($1, $2, $3, 'in_app', 'pending', $4, 'sendgrid', 0, '')
	`, ids.patient, "Appointment reminder", "Upcoming follow-up scheduled. Review your notes before the visit.", now.Add(42*time.Hour))

	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO notifications (user_id, title, body, channel, status, scheduled_for, provider, attempts, last_error)
		VALUES ($1, $2, $3, 'email', 'pending', $4, 'sendgrid', 0, '')
	`, ids.patient, "Lab result alert", "Your latest lab review is available. A clinician will follow up if needed.", now.Add(48*time.Hour))

	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO notifications (user_id, title, body, channel, status, scheduled_for, provider, attempts, last_error)
		VALUES ($1, $2, $3, 'in_app', 'pending', $4, 'sendgrid', 0, '')
	`, ids.patient, "Announcement: clinic updates", "New patient portal updates are live. Please review your preferences.", now.Add(60*time.Hour))

	// AI summary ready documents: normal + critical.
	docBodyNormal := buildMinimalPDFBytes("CBC panel within expected range. Continue current regimen and hydration.")
	docBodyCritical := buildMinimalPDFBytes("Critical alert: potassium level dangerously high. Immediate doctor follow-up required.")

	if err := insertPatientDocument(ctx, ids.patient, "cbc-report.pdf", docBodyNormal, "CBC panel within expected range.", "ready", "ready"); err != nil {
		return err
	}
	if err := insertPatientDocument(ctx, ids.patient, "critical-potassium-lab.pdf", docBodyCritical, "Critical potassium result; urgent follow-up needed.", "ready", "ready"); err != nil {
		return err
	}

	// A non-ready document to exercise AI summarize usability (fallback/local extraction).
	aiUnreadyBody := buildMinimalPDFBytes("A1c test: normal. No urgent abnormalities detected. Continue lifestyle plan.")
	if err := insertPatientDocument(ctx, ids.patient, "a1c-lab.pdf", aiUnreadyBody, "", "none", "pending"); err != nil {
		return err
	}

	// Knowledge base scenario docs for stale/conflict/similarity workflows.
	if err := insertKnowledgeDocIfMissing(ctx, ids.admin, "Hypertension Protocol v1", "Start with low-dose ACE inhibitor; monitor weekly."); err != nil {
		return err
	}
	if err := insertKnowledgeDocIfMissing(ctx, ids.admin, "Hypertension Protocol Legacy", "Begin with beta blocker as first-line for all patients."); err != nil {
		return err
	}
	if err := insertKnowledgeDocIfMissing(ctx, ids.admin, "Lab Escalation SOP", "Escalate critical results within 15 minutes and track acknowledgments."); err != nil {
		return err
	}
	if err := insertKnowledgeDocIfMissing(ctx, ids.admin, "Diabetes Monitoring Checklist", "Check A1c trends quarterly; ensure medication adherence and lifestyle follow-up."); err != nil {
		return err
	}
	if err := insertKnowledgeDocIfMissing(ctx, ids.admin, "High-Risk Action Guardrails", "Before escalating, verify patient identity, consent, and device/source reliability."); err != nil {
		return err
	}

	// Interoperability test fixtures for FHIR/HL7 boundaries.
	hl7Fixture := strings.Join([]string{
		"MSH|^~\\&|LAB|HOSP|HEALTHONYX|APP|" + now.Format("20060102150405") + "||ORU^R01|SEED-MSG-1|P|2.5",
		"PID|1||" + ids.patient.String() + "^^^HEALTHONYX||Demo^Patient",
		"OBX|1|NM|GLU^Glucose||110|mg/dL|70-110|N|||F",
	}, "\r")
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO hl7_lab_ingestion_events (
			message_control_id, patient_identifier, patient_id, observation_code, observation_value, units,
			transform_status, transform_outcome, hl7_message
		) VALUES ($1, $2, $3, $4, $5, $6, 'reconciled', 'seed_fixture_loaded', $7)
	`, "SEED-MSG-1", ids.patient.String(), ids.patient, "GLU", "110", "mg/dL", hl7Fixture)

	// Admin AI runtime seeded for observability/usability checks.
	_, _ = db.Pool.Exec(ctx, `
		UPDATE admin_ai_runtime_settings
		SET fallback_enabled = TRUE,
		    rate_limit_enabled = TRUE,
		    cache_enabled = TRUE,
		    prompt_version = 'v1',
		    updated_by = $1,
		    updated_at = NOW()
		WHERE id = TRUE
	`, ids.admin)

	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO admin_ai_eval_runs (model_name, prompt_version, fixture_count, passed_count, success_rate, quality_score, runtime_mode, ai_enabled, ollama_reachable, model_available, created_by)
		VALUES ('llama3.1:8b', 'v1', 5, 4, 0.8, 0.84, 'fallback_local', FALSE, FALSE, FALSE, $1)
	`, ids.admin)

	// Admin audit log entries to make the UI useful out of the box.
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO audit_logs (actor_user_id, action, entity_type, entity_id, detail)
		VALUES
			($1, 'seeded_demo_data', 'knowledge_docs', 'Hypertension Protocol v1', 'Seeded demo knowledge baseline'),
			($1, 'seeded_demo_data', 'notifications', 'Medication reminder', 'Seeded reminder notification'),
			($1, 'seeded_demo_data', 'doctor_slots', 'availability_windows', 'Seeded multiple doctor availability windows'),
			($1, 'seeded_demo_data', 'ai_runtime', 'cache', 'Seeded AI runtime settings for usability'),
			($1, 'seeded_demo_data', 'patient_documents', 'cbc-report.pdf', 'Seeded patient document fixtures')
	`, ids.admin)

	fmt.Printf("Seeded clinical scenarios for patient=%s doctor=%s rx=%s appointment=%s\n", ids.patient, ids.doctor, rxID, apptID)
	return nil
}

func insertPatientDocument(ctx context.Context, patientID uuid.UUID, filename string, body []byte, summary string, summaryStatus string, docStatus string) error {
	var existing int
	if err := db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM patient_documents WHERE patient_id = $1 AND filename = $2`, patientID, filename).Scan(&existing); err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}
	hasBody, err := tableHasColumn(ctx, "patient_documents", "body")
	if err != nil {
		return err
	}
	hasFileData, err := tableHasColumn(ctx, "patient_documents", "file_data")
	if err != nil {
		return err
	}

	switch {
	case hasBody && hasFileData:
		_, err = db.Pool.Exec(ctx, `
			INSERT INTO patient_documents (patient_id, filename, content_type, size_bytes, status, body, file_data, summary, summary_status)
			VALUES ($1, $2, 'application/pdf', $3, $4, $5, $6, $7, $8)
		`, patientID, filename, len(body), docStatus, bytes.Clone(body), bytes.Clone(body), summary, summaryStatus)
		return err
	case hasBody:
		_, err = db.Pool.Exec(ctx, `
			INSERT INTO patient_documents (patient_id, filename, content_type, size_bytes, status, body, summary, summary_status)
			VALUES ($1, $2, 'application/pdf', $3, $4, $5, $6, $7)
		`, patientID, filename, len(body), docStatus, bytes.Clone(body), summary, summaryStatus)
		return err
	case hasFileData:
		_, err = db.Pool.Exec(ctx, `
			INSERT INTO patient_documents (patient_id, filename, file_data, summary, summary_status)
			VALUES ($1, $2, $3, $4, $5)
		`, patientID, filename, bytes.Clone(body), summary, summaryStatus)
		return err
	default:
		return fmt.Errorf("patient_documents has neither body nor file_data columns")
	}
}

func insertKnowledgeDocIfMissing(ctx context.Context, createdBy uuid.UUID, title string, body string) error {
	var existing int
	if err := db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM knowledge_docs WHERE title = $1`, title).Scan(&existing); err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO knowledge_docs (title, body, created_by)
		VALUES ($1, $2, $3)
	`, title, body, createdBy)
	return err
}

func tableHasColumn(ctx context.Context, tableName string, columnName string) (bool, error) {
	var exists bool
	err := db.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name = $1
			  AND column_name = $2
		)
	`, tableName, columnName).Scan(&exists)
	return exists, err
}
