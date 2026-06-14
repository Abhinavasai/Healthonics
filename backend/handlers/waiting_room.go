package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
)

type WaitingRoomHandler struct{}

func NewWaitingRoomHandler() *WaitingRoomHandler {
	return &WaitingRoomHandler{}
}

type waitingRoomEntry struct {
	AppointmentID string     `json:"appointment_id"`
	PatientEmail  string     `json:"patient_email"`
	ScheduledAt   time.Time  `json:"scheduled_at"`
	CheckedInAt   *time.Time `json:"checked_in_at,omitempty"`
	Reason        string     `json:"reason"`
	Status        string     `json:"status"`
	WaitMinutes   *int       `json:"wait_minutes,omitempty"`
}

// Queue — doctor sees today's approved appointments sorted by check-in time then scheduled time.
func (h *WaitingRoomHandler) Queue(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	ctx := c.Request.Context()
	rows, err := db.Pool.Query(ctx, `
		SELECT a.id::text, u.email, a.scheduled_at, a.checked_in_at, a.reason, a.status
		FROM appointments a
		JOIN users u ON u.id = a.patient_id
		WHERE a.doctor_id = $1
		  AND a.scheduled_at::date = NOW()::date
		  AND a.status IN ('approved', 'completed')
		ORDER BY a.checked_in_at ASC NULLS LAST, a.scheduled_at ASC
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	var out []waitingRoomEntry
	now := time.Now()
	for rows.Next() {
		var e waitingRoomEntry
		if err := rows.Scan(&e.AppointmentID, &e.PatientEmail, &e.ScheduledAt, &e.CheckedInAt, &e.Reason, &e.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		if e.CheckedInAt != nil {
			mins := int(now.Sub(*e.CheckedInAt).Minutes())
			if mins < 0 {
				mins = 0
			}
			e.WaitMinutes = &mins
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if out == nil {
		out = []waitingRoomEntry{}
	}
	c.JSON(http.StatusOK, gin.H{"queue": out})
}

// CheckIn — patient checks in for their appointment.
func (h *WaitingRoomHandler) CheckIn(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	apptID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment id"})
		return
	}
	ctx := c.Request.Context()
	res, err := db.Pool.Exec(ctx, `
		UPDATE appointments
		SET checked_in_at = COALESCE(checked_in_at, NOW())
		WHERE id = $1 AND patient_id = $2 AND status = 'approved'
		  AND scheduled_at::date = NOW()::date
	`, apptID, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if res.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Approved appointment for today not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "checked_in_at": time.Now().UTC()})
}
