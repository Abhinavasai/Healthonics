package workers

import (
	"context"
	"log/slog"
	"time"

	"github.com/healthonyx/backend/db"
)

// StartAnomalyWorker runs every 10 minutes and checks for audit access spikes.
// If any user accessed more than 50 audit events in the last 10 minutes
// AND that's ≥3× their rolling hourly average, it fires admin notifications.
func StartAnomalyWorker(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runAnomalyCheck(ctx)
			}
		}
	}()
}

func runAnomalyCheck(ctx context.Context) {
	// Find users who accessed many audit records in the last 10 min.
	rows, err := db.Pool.Query(ctx, `
		WITH recent AS (
		    SELECT actor_user_id AS user_id, COUNT(*) AS recent_count
		    FROM audit_logs
		    WHERE created_at >= NOW() - INTERVAL '10 minutes'
		      AND actor_user_id IS NOT NULL
		    GROUP BY actor_user_id
		    HAVING COUNT(*) >= 50
		),
		rolling AS (
		    SELECT actor_user_id AS user_id,
		           COUNT(*) / NULLIF(EXTRACT(EPOCH FROM (MAX(created_at) - MIN(created_at))) / 3600, 0) AS hourly_rate
		    FROM audit_logs
		    WHERE created_at >= NOW() - INTERVAL '24 hours'
		      AND actor_user_id IS NOT NULL
		    GROUP BY actor_user_id
		)
		SELECT r.user_id, r.recent_count, COALESCE(ro.hourly_rate, 0)
		FROM recent r
		LEFT JOIN rolling ro ON ro.user_id = r.user_id
		WHERE COALESCE(ro.hourly_rate, 0) = 0
		   OR r.recent_count >= 3 * COALESCE(ro.hourly_rate, 1)
	`)
	if err != nil {
		slog.Error("anomaly worker: query failed", "error", err)
		return
	}
	defer rows.Close()

	type spike struct {
		UserID      string
		RecentCount int
		HourlyRate  float64
	}
	var spikes []spike
	for rows.Next() {
		var s spike
		if err := rows.Scan(&s.UserID, &s.RecentCount, &s.HourlyRate); err == nil {
			spikes = append(spikes, s)
		}
	}

	if len(spikes) == 0 {
		return
	}

	// Fetch all admin user IDs.
	adminRows, err := db.Pool.Query(ctx, `SELECT id FROM users WHERE role = 'admin' AND is_active = TRUE`)
	if err != nil {
		slog.Error("anomaly worker: admin fetch failed", "error", err)
		return
	}
	defer adminRows.Close()

	var adminIDs []string
	for adminRows.Next() {
		var id string
		if err := adminRows.Scan(&id); err == nil {
			adminIDs = append(adminIDs, id)
		}
	}

	for _, s := range spikes {
		msg := buildAnomalyMessage(s.UserID, s.RecentCount, s.HourlyRate)
		for _, adminID := range adminIDs {
		if _, err := db.Pool.Exec(ctx, `
				INSERT INTO notifications (user_id, title, body, channel, status, scheduled_for)
				VALUES ($1, 'Unusual Access Pattern Detected', $2, 'in_app', 'pending', NOW())
			`, adminID, msg); err != nil {
				slog.Error("anomaly worker: notification failed", "admin", adminID, "error", err)
			}
		}
		slog.Warn("anomaly worker: access spike detected", "user_id", s.UserID, "count", s.RecentCount, "avg_hourly", s.HourlyRate)
	}
}

func buildAnomalyMessage(userID string, count int, hourlyRate float64) string {
	end := len(userID)
	if end > 8 {
		end = 8
	}
	return "User " + userID[:end] + "... accessed " + itoa(count) + " audit records in 10 minutes" +
		" (normal hourly rate: " + ftoa(hourlyRate, 1) + "). Review audit log for suspicious activity."
}

func itoa(n int) string {
	return itoa10(n)
}

func itoa10(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

func ftoa(f float64, decimals int) string {
	if decimals == 0 {
		return itoa(int(f))
	}
	whole := int(f)
	frac := int((f - float64(whole)) * 10)
	if frac < 0 {
		frac = -frac
	}
	return itoa(whole) + "." + itoa(frac)
}
