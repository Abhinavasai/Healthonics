package main

import (
	"context"
	"fmt"
	"log"

	"github.com/healthonyx/backend/config"
	"github.com/healthonyx/backend/db"
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
		_, err = db.Pool.Exec(context.Background(),
			`INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3)
			 ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash, role = EXCLUDED.role`,
			u.email, string(hash), u.role,
		)
		if err != nil {
			log.Fatalf("seed %s: %v", u.email, err)
		}
		fmt.Printf("Seeded: %s / %s (%s)\n", u.email, u.pass, u.role)
	}

	fmt.Println("Demo users seeded. Use these credentials to log in.")
}
