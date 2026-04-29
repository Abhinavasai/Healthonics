package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
)

type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

// aggregateUnreadMessages matches GET /api/messages/unread logic for embedding in dashboard payloads.
func aggregateUnreadMessages(ctx context.Context, userID uuid.UUID) (int, error) {
	var total int
	err := db.Pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(sub.c), 0)::int
		FROM (
			SELECT (
				SELECT COUNT(*)::int
				FROM messages m
				WHERE m.thread_id = mt.id
				  AND m.sender_id <> $1
				  AND m.created_at > COALESCE(
					(SELECT r.last_read_at FROM message_thread_reads r
					 WHERE r.thread_id = mt.id AND r.user_id = $1),
					'-infinity'::timestamptz
				  )
			) AS c
			FROM message_threads mt
			WHERE mt.patient_id = $1 OR mt.doctor_id = $1
		) sub
	`, userID).Scan(&total)
	return total, err
}

func buildPatientAlerts(upcoming, pending, unread int) []gin.H {
	out := []gin.H{}
	if pending > 0 {
		out = append(out, gin.H{
			"severity": "warning",
			"code":     "pending_appointment_requests",
			"message":  "You have appointment requests awaiting confirmation.",
			"count":    pending,
		})
	}
	if unread > 0 {
		sev := "info"
		if unread >= 10 {
			sev = "warning"
		}
		out = append(out, gin.H{
			"severity": sev,
			"code":     "unread_messages",
			"message":  "Unread messages in your inbox.",
			"count":    unread,
		})
	}
	if upcoming > 3 && pending == 0 {
		out = append(out, gin.H{
			"severity": "info",
			"code":     "busy_schedule",
			"message":  "You have several upcoming appointments on your calendar.",
			"count":    upcoming,
		})
	}
	return out
}

func buildDoctorAlerts(today, pending, unread int) []gin.H {
	out := []gin.H{}
	if pending > 0 {
		out = append(out, gin.H{
			"severity": "warning",
			"code":     "pending_queue",
			"message":  "Appointment requests need your review.",
			"count":    pending,
		})
	}
	if unread > 0 {
		sev := "info"
		if unread >= 10 {
			sev = "warning"
		}
		out = append(out, gin.H{
			"severity": sev,
			"code":     "unread_messages",
			"message":  "Unread messages from patients.",
			"count":    unread,
		})
	}
	if today >= 8 {
		out = append(out, gin.H{
			"severity": "warning",
			"code":     "heavy_clinic_day",
			"message":  "Many appointments scheduled for today.",
			"count":    today,
		})
	}
	return out
}

func (h *DashboardHandler) PatientSummary(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	ctx := c.Request.Context()
	var upcoming, pending int
	err := db.Pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status IN ('pending','approved') AND scheduled_at >= NOW())::int,
			COUNT(*) FILTER (WHERE status = 'pending')::int
		FROM appointments WHERE patient_id = $1
	`, claims.UserID).Scan(&upcoming, &pending)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	unread, err := aggregateUnreadMessages(ctx, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	var rxCount int
	if err := db.Pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM prescriptions WHERE patient_id = $1 AND status = 'active'
	`, claims.UserID).Scan(&rxCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	var docCount int
	if err := db.Pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM patient_documents WHERE patient_id = $1
	`, claims.UserID).Scan(&docCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	alerts := buildPatientAlerts(upcoming, pending, unread)

	c.JSON(http.StatusOK, gin.H{
		"role":                     "patient",
		"upcoming_or_active_count": upcoming,
		"pending_requests":         pending,
		"aggregations": gin.H{
			"unread_messages":      unread,
			"active_prescriptions": rxCount,
			"documents_uploaded":   docCount,
		},
		"alerts": alerts,
	})
}

func (h *DashboardHandler) DoctorSummary(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	ctx := c.Request.Context()
	var today, pending, week int
	err := db.Pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (
				WHERE scheduled_at >= CURRENT_DATE
				  AND scheduled_at < CURRENT_DATE + INTERVAL '1 day'
				  AND status IN ('pending','approved')
			)::int,
			COUNT(*) FILTER (WHERE status = 'pending')::int,
			COUNT(*) FILTER (
				WHERE scheduled_at >= date_trunc('week', CURRENT_TIMESTAMP)
				  AND scheduled_at < date_trunc('week', CURRENT_TIMESTAMP) + INTERVAL '7 days'
				  AND status IN ('pending','approved','completed','no_show','reschedule_requested')
			)::int
		FROM appointments WHERE doctor_id = $1
	`, claims.UserID).Scan(&today, &pending, &week)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	unread, err := aggregateUnreadMessages(ctx, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	if err := refreshDoctorCriticalEscalations(ctx, claims.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	criticalEscalations, err := listDoctorCriticalEscalations(ctx, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	alerts := buildDoctorAlerts(today, pending, unread)
	if len(criticalEscalations) > 0 {
		alerts = append([]gin.H{
			{
				"severity": "warning",
				"code":     "critical_result_escalations",
				"message":  "Critical findings require acknowledgement.",
				"count":    len(criticalEscalations),
			},
		}, alerts...)
	}

	c.JSON(http.StatusOK, gin.H{
		"role":               "doctor",
		"appointments_today": today,
		"pending_queue":      pending,
		"aggregations": gin.H{
			"unread_messages":        unread,
			"appointments_this_week": week,
			"critical_open_count":    len(criticalEscalations),
		},
		"alerts":               alerts,
		"critical_escalations": criticalEscalations,
	})
}
