package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Connect(databaseURL string) error {
	var err error
	Pool, err = pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return err
	}
	return nil
}

func Migrate(ctx context.Context) error {
	_, err := Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email      TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role       TEXT NOT NULL CHECK (role IN ('patient', 'doctor', 'admin')),
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS appointments (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			patient_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			doctor_id    UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
			scheduled_at TIMESTAMPTZ NOT NULL,
			reason       TEXT NOT NULL,
			status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CHECK (patient_id <> doctor_id)
		);

		CREATE INDEX IF NOT EXISTS idx_appointments_patient_id ON appointments(patient_id);
		CREATE INDEX IF NOT EXISTS idx_appointments_doctor_id ON appointments(doctor_id);
		CREATE INDEX IF NOT EXISTS idx_appointments_status ON appointments(status);
		CREATE INDEX IF NOT EXISTS idx_appointments_scheduled_at ON appointments(scheduled_at);
	`)
	return err
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
}
