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
			status       TEXT NOT NULL DEFAULT 'pending',
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CHECK (patient_id <> doctor_id)
		);

		ALTER TABLE appointments DROP CONSTRAINT IF EXISTS appointments_status_check;
		ALTER TABLE appointments ADD CONSTRAINT appointments_status_check CHECK (status IN (
			'pending', 'approved', 'rejected', 'cancelled', 'completed', 'no_show', 'reschedule_requested'
		));

		CREATE INDEX IF NOT EXISTS idx_appointments_patient_id ON appointments(patient_id);
		CREATE INDEX IF NOT EXISTS idx_appointments_doctor_id ON appointments(doctor_id);
		CREATE INDEX IF NOT EXISTS idx_appointments_status ON appointments(status);
		CREATE INDEX IF NOT EXISTS idx_appointments_scheduled_at ON appointments(scheduled_at);

		CREATE TABLE IF NOT EXISTS appointment_activities (
			id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			appointment_id  UUID NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
			actor_user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			action          TEXT NOT NULL,
			detail          TEXT NOT NULL DEFAULT '',
			created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_appointment_activities_appt ON appointment_activities(appointment_id);

		CREATE TABLE IF NOT EXISTS hospitals (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name       TEXT NOT NULL,
			city       TEXT NOT NULL DEFAULT '',
			region     TEXT NOT NULL DEFAULT '',
			latitude   DOUBLE PRECISION NOT NULL,
			longitude  DOUBLE PRECISION NOT NULL
		);

		CREATE UNIQUE INDEX IF NOT EXISTS idx_hospitals_name_region ON hospitals(name, region);

		ALTER TABLE users ADD COLUMN IF NOT EXISTS specialization TEXT NOT NULL DEFAULT '';
		ALTER TABLE users ADD COLUMN IF NOT EXISTS hospital_id UUID REFERENCES hospitals(id) ON DELETE SET NULL;
		ALTER TABLE users ADD COLUMN IF NOT EXISTS practice_latitude DOUBLE PRECISION;
		ALTER TABLE users ADD COLUMN IF NOT EXISTS practice_longitude DOUBLE PRECISION;

		CREATE INDEX IF NOT EXISTS idx_users_doctor_spec ON users(role, specialization) WHERE role = 'doctor';

		CREATE TABLE IF NOT EXISTS doctor_slots (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			doctor_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			start_at    TIMESTAMPTZ NOT NULL,
			end_at      TIMESTAMPTZ NOT NULL,
			patient_id  UUID REFERENCES users(id) ON DELETE SET NULL,
			CONSTRAINT doctor_slots_time_order CHECK (end_at > start_at),
			CONSTRAINT doctor_slots_unique_start UNIQUE (doctor_id, start_at)
		);

		CREATE INDEX IF NOT EXISTS idx_doctor_slots_doctor_time ON doctor_slots(doctor_id, start_at);
		CREATE INDEX IF NOT EXISTS idx_doctor_slots_open ON doctor_slots(doctor_id) WHERE patient_id IS NULL;

		ALTER TABLE appointments ADD COLUMN IF NOT EXISTS slot_id UUID UNIQUE REFERENCES doctor_slots(id) ON DELETE SET NULL;

		CREATE TABLE IF NOT EXISTS message_threads (
			id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			patient_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			doctor_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			last_message_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			last_preview    TEXT NOT NULL DEFAULT '',
			created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CHECK (patient_id <> doctor_id),
			UNIQUE (patient_id, doctor_id)
		);

		CREATE INDEX IF NOT EXISTS idx_message_threads_patient ON message_threads(patient_id);
		CREATE INDEX IF NOT EXISTS idx_message_threads_doctor ON message_threads(doctor_id);

		CREATE TABLE IF NOT EXISTS messages (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			thread_id  UUID NOT NULL REFERENCES message_threads(id) ON DELETE CASCADE,
			sender_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			body       TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_messages_thread_time ON messages(thread_id, created_at);

		CREATE TABLE IF NOT EXISTS message_thread_reads (
			thread_id    UUID NOT NULL REFERENCES message_threads(id) ON DELETE CASCADE,
			user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			last_read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (thread_id, user_id)
		);
	`)
	return err
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
}
