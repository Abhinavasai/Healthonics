package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/jackc/pgx/v5"
)

type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

// Patient returns aggregated counts for the patient home view.
func (h *DashboardHandler) Patient(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok || claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	ctx := c.Request.Context()
	uid := claims.UserID
	now := time.Now().UTC()

	var upcoming, pending, unread int
	err := db.Pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*)::int FROM appointments
			 WHERE patient_id = $1 AND scheduled_at > $2
			   AND status IN ('pending','approved','reschedule_requested')),
			(SELECT COUNT(*)::int FROM appointments
			 WHERE patient_id = $1 AND status = 'pending')
	`, uid, now).Scan(&upcoming, &pending)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	err = db.Pool.QueryRow(ctx, `
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
	`, uid).Scan(&unread)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	var nextID uuid.UUID
	var nextAt time.Time
	var nextStatus string
	var docEmail string
	err = db.Pool.QueryRow(ctx, `
		SELECT a.id, a.scheduled_at, a.status, u.email
		FROM appointments a
		JOIN users u ON u.id = a.doctor_id
		WHERE a.patient_id = $1 AND a.scheduled_at >= $2
		  AND a.status IN ('pending','approved','reschedule_requested')
		ORDER BY a.scheduled_at ASC
		LIMIT 1
	`, uid, now).Scan(&nextID, &nextAt, &nextStatus, &docEmail)

	out := gin.H{
		"upcoming_count":  upcoming,
		"pending_count":   pending,
		"unread_messages": unread,
	}
	if err == nil {
		out["next_appointment"] = gin.H{
			"id":           nextID.String(),
			"scheduled_at": nextAt,
			"status":       nextStatus,
			"doctor_email": docEmail,
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, out)
}

// Doctor returns aggregated counts for the doctor home view.
func (h *DashboardHandler) Doctor(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok || claims.Role != "doctor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	ctx := c.Request.Context()
	uid := claims.UserID
	now := time.Now().UTC()
	startDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endDay := startDay.Add(24 * time.Hour)

	var todayCount, pendingQueue, unread int
	err := db.Pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*)::int FROM appointments
			 WHERE doctor_id = $1 AND scheduled_at >= $2 AND scheduled_at < $3
			   AND status NOT IN ('cancelled','rejected')),
			(SELECT COUNT(*)::int FROM appointments
			 WHERE doctor_id = $1 AND status = 'pending')
	`, uid, startDay, endDay).Scan(&todayCount, &pendingQueue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	err = db.Pool.QueryRow(ctx, `
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
	`, uid).Scan(&unread)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"today_appointments_count": todayCount,
		"pending_queue_count":      pendingQueue,
		"unread_messages":          unread,
	})
}

// Admin returns system health and usage counts.
func (h *DashboardHandler) Admin(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok || claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	ctx := c.Request.Context()

	var patients, doctors, admins int
	err := db.Pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*)::int FROM users WHERE role = 'patient' AND COALESCE(active, TRUE)),
			(SELECT COUNT(*)::int FROM users WHERE role = 'doctor' AND COALESCE(active, TRUE)),
			(SELECT COUNT(*)::int FROM users WHERE role = 'admin' AND COALESCE(active, TRUE))
	`).Scan(&patients, &doctors, &admins)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	if err := db.Pool.Ping(ctx); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"db_ok":          false,
			"patients_count": patients,
			"doctors_count":  doctors,
			"admins_count":   admins,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"db_ok":          true,
		"patients_count": patients,
		"doctors_count":  doctors,
		"admins_count":   admins,
	})
}
