package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/healthonyx/backend/config"
	"github.com/healthonyx/backend/db"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

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

	hospitals := []struct {
		name, city, region string
		lat, lng            float64
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

	_, err := db.Pool.Exec(ctx, `
		INSERT INTO hospital_departments (hospital_id, department_name)
		SELECT h.id, d.dep
		FROM hospitals h
		CROSS JOIN (
			VALUES
				('Emergency Medicine'), ('Internal Medicine'), ('Cardiology'), ('Family Medicine'),
				('Orthopedics'), ('Neurology'), ('Pediatrics'), ('Surgery')
		) AS d(dep)
		ON CONFLICT (hospital_id, department_name) DO NOTHING
	`)
	if err != nil {
		log.Printf("hospital_departments seed: %v", err)
	} else {
		fmt.Println("Seeded hospital departments (per hospital).")
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

	// One demo open slot for the demo doctor (tomorrow 10:00 UTC window — adjust in app as needed)
	var docID uuid.UUID
	err = db.Pool.QueryRow(ctx, `SELECT id FROM users WHERE email = 'doctor@healthonyx.demo' AND role = 'doctor'`).Scan(&docID)
	if err != nil {
		log.Printf("demo slot: doctor id: %v", err)
	} else {
		start := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Hour)
		start = time.Date(start.Year(), start.Month(), start.Day(), 14, 0, 0, 0, time.UTC)
		end := start.Add(30 * time.Minute)
		_, err = db.Pool.Exec(ctx, `
			INSERT INTO doctor_slots (doctor_id, start_at, end_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (doctor_id, start_at) DO NOTHING
		`, docID, start, end)
		if err != nil {
			log.Printf("demo slot: %v", err)
		} else {
			fmt.Printf("Demo doctor slot: %s – %s (UTC)\n", start.Format(time.RFC3339), end.Format(time.RFC3339))
		}
	}

	fmt.Println("Demo users seeded. Use these credentials to log in.")
}
