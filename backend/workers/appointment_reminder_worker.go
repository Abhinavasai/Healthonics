package workers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/healthonyx/backend/db"
)

// AppointmentReminderWorker sends in-app notifications 24h and 1h before appointments.
type AppointmentReminderWorker struct {
	interval time.Duration
}

func NewAppointmentReminderWorker(interval time.Duration) *AppointmentReminderWorker {
	return &AppointmentReminderWorker{interval: interval}
}

func (w *AppointmentReminderWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.tick(ctx); err != nil {
				slog.Error("appointment reminder worker error", "error", err)
			}
		}
	}
}

func (w *AppointmentReminderWorker) tick(ctx context.Context) error {
	for _, reminder := range []struct {
		reminderType string
		window       time.Duration
		label        string
	}{
		{"24h", 24 * time.Hour, "24 hours"},
		{"1h", 1 * time.Hour, "1 hour"},
	} {
		if err := w.sendReminders(ctx, reminder.reminderType, reminder.window, reminder.label); err != nil {
			slog.Error("reminder batch failed", "type", reminder.reminderType, "error", err)
		}
	}
	return nil
}

func (w *AppointmentReminderWorker) sendReminders(ctx context.Context, reminderType string, window time.Duration, label string) error {
	// Find appointments scheduled within [window, window+interval] that haven't had this reminder yet.
	now := time.Now()
	windowStart := now.Add(window)
	windowEnd := now.Add(window + w.interval)

	rows, err := db.Pool.Query(ctx, `
		SELECT a.id::text, a.patient_id::text, u.email, a.scheduled_at, a.reason
		FROM appointments a
		JOIN users u ON u.id = a.patient_id
		WHERE a.status = 'approved'
		  AND a.scheduled_at >= $1
		  AND a.scheduled_at < $2
		  AND NOT EXISTS (
			SELECT 1 FROM appointment_reminder_log l
			WHERE l.appointment_id = a.id AND l.reminder_type = $3
		  )
	`, windowStart, windowEnd, reminderType)
	if err != nil {
		return err
	}
	defer rows.Close()

	type apptRow struct {
		ID          string
		PatientID   string
		Email       string
		ScheduledAt time.Time
		Reason      string
	}
	var appts []apptRow
	for rows.Next() {
		var r apptRow
		if err := rows.Scan(&r.ID, &r.PatientID, &r.Email, &r.ScheduledAt, &r.Reason); err != nil {
			return err
		}
		appts = append(appts, r)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, appt := range appts {
		title := fmt.Sprintf("Appointment in %s", label)
		body := fmt.Sprintf("Your appointment is scheduled for %s. Reason: %s",
			appt.ScheduledAt.Format("Jan 2, 2006 at 3:04 PM"), appt.Reason)

		// Insert in-app notification.
		_, err := db.Pool.Exec(ctx, `
			INSERT INTO notifications (user_id, title, body, channel, status)
			VALUES ($1::uuid, $2, $3, 'in_app', 'pending')
		`, appt.PatientID, title, body)
		if err != nil {
			slog.Error("reminder: failed to insert notification", "appointment_id", appt.ID, "error", err)
			continue
		}

		// Mark this reminder as sent to avoid duplicates.
		_, err = db.Pool.Exec(ctx, `
			INSERT INTO appointment_reminder_log (appointment_id, reminder_type)
			VALUES ($1::uuid, $2)
			ON CONFLICT DO NOTHING
		`, appt.ID, reminderType)
		if err != nil {
			slog.Error("reminder: failed to log reminder", "appointment_id", appt.ID, "error", err)
		}
		slog.Info("appointment reminder sent", "type", reminderType, "appointment_id", appt.ID)
	}
	return nil
}
