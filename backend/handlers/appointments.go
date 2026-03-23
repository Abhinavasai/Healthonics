package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/healthonyx/backend/models"
)

type AppointmentHandler struct{}

type CreateAppointmentRequest struct {
	DoctorID    string    `json:"doctor_id" binding:"required"`
	ScheduledAt time.Time `json:"scheduled_at" binding:"required"`
	Reason      string    `json:"reason" binding:"required"`
}

func NewAppointmentHandler() *AppointmentHandler {
	return &AppointmentHandler{}
}

func getClaims(c *gin.Context) (*Claims, bool) {
	claimsVal, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return nil, false
	}
	claims, ok := claimsVal.(*Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return nil, false
	}
	return claims, true
}

func (h *AppointmentHandler) Create(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}

	var req CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.ScheduledAt.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scheduled_at is required"})
		return
	}

	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reason is required"})
		return
	}

	doctorID, err := uuid.Parse(req.DoctorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "doctor_id must be a valid UUID"})
		return
	}
	if doctorID == claims.UserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "doctor_id cannot match patient user"})
		return
	}

	var doctorExists bool
	err = db.Pool.QueryRow(c.Request.Context(),
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND role = 'doctor')`,
		doctorID,
	).Scan(&doctorExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if !doctorExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "doctor_id must reference a doctor account"})
		return
	}

	var appt models.Appointment
	err = db.Pool.QueryRow(c.Request.Context(), `
		INSERT INTO appointments (patient_id, doctor_id, scheduled_at, reason, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, patient_id, doctor_id, scheduled_at, reason, status, created_at, updated_at
	`, claims.UserID, doctorID, req.ScheduledAt, reason, models.AppointmentStatusPending).
		Scan(&appt.ID, &appt.PatientID, &appt.DoctorID, &appt.ScheduledAt, &appt.Reason, &appt.Status, &appt.CreatedAt, &appt.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusCreated, appt)
}

func (h *AppointmentHandler) ListPatient(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}

	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id, patient_id, doctor_id, scheduled_at, reason, status, created_at, updated_at
		FROM appointments
		WHERE patient_id = $1
		ORDER BY scheduled_at DESC, created_at DESC
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	appointments := make([]models.Appointment, 0)
	for rows.Next() {
		var appt models.Appointment
		if err := rows.Scan(&appt.ID, &appt.PatientID, &appt.DoctorID, &appt.ScheduledAt, &appt.Reason, &appt.Status, &appt.CreatedAt, &appt.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		appointments = append(appointments, appt)
	}
	if rows.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"appointments": appointments})
}

func (h *AppointmentHandler) ListDoctor(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}

	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id, patient_id, doctor_id, scheduled_at, reason, status, created_at, updated_at
		FROM appointments
		WHERE doctor_id = $1
		ORDER BY status ASC, scheduled_at ASC, created_at DESC
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	appointments := make([]models.Appointment, 0)
	for rows.Next() {
		var appt models.Appointment
		if err := rows.Scan(&appt.ID, &appt.PatientID, &appt.DoctorID, &appt.ScheduledAt, &appt.Reason, &appt.Status, &appt.CreatedAt, &appt.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		appointments = append(appointments, appt)
	}
	if rows.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"appointments": appointments})
}
