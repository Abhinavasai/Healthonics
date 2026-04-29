package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var criticalKeywords = []string{
	"critical",
	"urgent",
	"abnormal",
	"high risk",
	"severe",
}

func hasCriticalSignal(summary, summaryStatus, summaryErr string) bool {
	if strings.EqualFold(strings.TrimSpace(summaryStatus), "failed") {
		return true
	}
	s := strings.ToLower(strings.TrimSpace(summary + " " + summaryErr))
	for _, k := range criticalKeywords {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}

func escalationStepDuration(nextTier int) time.Duration {
	switch nextTier {
	case 2:
		return 15 * time.Minute
	case 3:
		return 30 * time.Minute
	default:
		return 60 * time.Minute
	}
}

func nextEscalation(tier int, now time.Time) (int, time.Time) {
	if tier < 3 {
		tier++
	}
	return tier, now.Add(escalationStepDuration(tier))
}

func escalationMessage(tier int, filename, patientEmail string) string {
	return fmt.Sprintf("Tier %d critical result: %s for patient %s requires review.", tier, filename, patientEmail)
}

func isPgUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}
	// Some driver/code paths surface this as plain text; guard for the known index.
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(msg, "idx_critical_result_escalations_active_doc") &&
		strings.Contains(msg, "duplicate key value")
}

func enqueueEscalationNotification(ctx context.Context, doctorID uuid.UUID, title, body string) {
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO notifications (user_id, title, body, channel, status, scheduled_for)
		VALUES ($1, $2, $3, 'in_app', 'pending', NOW())
	`, doctorID, title, body)
}

type criticalEscalationRow struct {
	ID               string `json:"id"`
	DocumentID       string `json:"document_id"`
	PatientID        string `json:"patient_id"`
	PatientEmail     string `json:"patient_email"`
	Filename         string `json:"filename"`
	Tier             int    `json:"tier"`
	Status           string `json:"status"`
	RuleCode         string `json:"rule_code"`
	Message          string `json:"message"`
	FirstDetectedAt  string `json:"first_detected_at"`
	NextEscalationAt string `json:"next_escalation_at"`
	AcknowledgedAt   string `json:"acknowledged_at,omitempty"`
}

func refreshDoctorCriticalEscalations(ctx context.Context, doctorID uuid.UUID) error {
	rows, err := db.Pool.Query(ctx, `
		SELECT DISTINCT d.id::text, d.patient_id::text, COALESCE(u.email, ''), d.filename,
		       COALESCE(d.summary, ''), COALESCE(d.summary_status, ''), COALESCE(d.summary_error, '')
		FROM patient_documents d
		JOIN appointments a ON a.patient_id = d.patient_id
		LEFT JOIN users u ON u.id = d.patient_id
		WHERE a.doctor_id = $1
	`, doctorID)
	if err != nil {
		return err
	}
	defer rows.Close()

	now := time.Now().UTC()
	for rows.Next() {
		var docID, patientID, patientEmail, filename, summary, summaryStatus, summaryErr string
		if err := rows.Scan(&docID, &patientID, &patientEmail, &filename, &summary, &summaryStatus, &summaryErr); err != nil {
			return err
		}
		if !hasCriticalSignal(summary, summaryStatus, summaryErr) {
			continue
		}

		var escID string
		var status string
		var tier int
		var nextAt time.Time
		err = db.Pool.QueryRow(ctx, `
			SELECT id::text, status, tier, next_escalation_at
			FROM critical_result_escalations
			WHERE doctor_id = $1 AND document_id::text = $2 AND status IN ('open', 'acknowledged')
			ORDER BY first_detected_at DESC
			LIMIT 1
		`, doctorID, docID).Scan(&escID, &status, &tier, &nextAt)
		if err == pgx.ErrNoRows {
			msg := escalationMessage(1, filename, patientEmail)
			_, insertErr := db.Pool.Exec(ctx, `
				INSERT INTO critical_result_escalations (
					doctor_id, patient_id, document_id, status, tier, rule_code, message,
					first_detected_at, last_evaluated_at, next_escalation_at
				)
				VALUES ($1, $2::uuid, $3::uuid, 'open', 1, 'critical_summary_keyword', $4, $5, $5, $6)
			`, doctorID, patientID, docID, msg, now, now.Add(escalationStepDuration(2)))
			if insertErr != nil {
				// Soak/perf tests run this refresh concurrently. If another goroutine
				// inserted the same escalation first, treat it as a benign race.
				if isPgUniqueViolation(insertErr) {
					continue
				}
				return insertErr
			}
			enqueueEscalationNotification(ctx, doctorID, "Critical result escalation (tier 1)", msg)
			continue
		}
		if err != nil {
			return err
		}

		if status == "open" && (nextAt.Before(now) || nextAt.Equal(now)) {
			newTier, newNext := nextEscalation(tier, now)
			msg := escalationMessage(newTier, filename, patientEmail)
			_, updateErr := db.Pool.Exec(ctx, `
				UPDATE critical_result_escalations
				SET tier = $2, message = $3, next_escalation_at = $4, last_evaluated_at = $5
				WHERE id::text = $1
			`, escID, newTier, msg, newNext, now)
			if updateErr != nil {
				return updateErr
			}
			enqueueEscalationNotification(ctx, doctorID, fmt.Sprintf("Critical result escalation (tier %d)", newTier), msg)
		} else {
			_, _ = db.Pool.Exec(ctx, `
				UPDATE critical_result_escalations
				SET last_evaluated_at = $2
				WHERE id::text = $1
			`, escID, now)
		}
	}
	return rows.Err()
}

func listDoctorCriticalEscalations(ctx context.Context, doctorID uuid.UUID) ([]criticalEscalationRow, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT e.id::text, e.document_id::text, e.patient_id::text, COALESCE(u.email, ''), COALESCE(d.filename, ''),
		       e.tier, e.status, e.rule_code, e.message,
		       e.first_detected_at::text, e.next_escalation_at::text, COALESCE(e.acknowledged_at::text, '')
		FROM critical_result_escalations e
		LEFT JOIN users u ON u.id = e.patient_id
		LEFT JOIN patient_documents d ON d.id = e.document_id
		WHERE e.doctor_id = $1
		  AND e.status IN ('open', 'acknowledged')
		ORDER BY e.first_detected_at DESC
		LIMIT 100
	`, doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []criticalEscalationRow{}
	for rows.Next() {
		var r criticalEscalationRow
		if err := rows.Scan(
			&r.ID, &r.DocumentID, &r.PatientID, &r.PatientEmail, &r.Filename,
			&r.Tier, &r.Status, &r.RuleCode, &r.Message, &r.FirstDetectedAt, &r.NextEscalationAt, &r.AcknowledgedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
