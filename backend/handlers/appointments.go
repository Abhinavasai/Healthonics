package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/healthonyx/backend/models"
	"github.com/jackc/pgx/v5"
)

type AppointmentHandler struct{}

type CreateAppointmentRequest struct {
	DoctorID    string    `json:"doctor_id" binding:"required"`
	ScheduledAt time.Time `json:"scheduled_at" binding:"required"`
	Reason      string    `json:"reason" binding:"required"`
}

type UpdateAppointmentStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type DoctorOption struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
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

// GetByID returns one appointment when the caller is the patient or assigned doctor.
func (h *AppointmentHandler) GetByID(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "patient" && claims.Role != "doctor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment id"})
		return
	}

	var appt models.Appointment
	err = db.Pool.QueryRow(c.Request.Context(), `
		SELECT id, patient_id, doctor_id, scheduled_at, reason, status, created_at, updated_at
		FROM appointments
		WHERE id = $1
	`, id).Scan(&appt.ID, &appt.PatientID, &appt.DoctorID, &appt.ScheduledAt, &appt.Reason, &appt.Status, &appt.CreatedAt, &appt.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	switch claims.Role {
	case "patient":
		if appt.PatientID != claims.UserID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
			return
		}
	case "doctor":
		if appt.DoctorID != claims.UserID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
			return
		}
	}

	c.JSON(http.StatusOK, appt)
}

// ListActivity returns audit rows for an appointment when the caller is the patient or assigned doctor.
func (h *AppointmentHandler) ListActivity(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "patient" && claims.Role != "doctor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment id"})
		return
	}

	var appt models.Appointment
	err = db.Pool.QueryRow(c.Request.Context(), `
		SELECT id, patient_id, doctor_id, scheduled_at, reason, status, created_at, updated_at
		FROM appointments
		WHERE id = $1
	`, id).Scan(&appt.ID, &appt.PatientID, &appt.DoctorID, &appt.ScheduledAt, &appt.Reason, &appt.Status, &appt.CreatedAt, &appt.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	switch claims.Role {
	case "patient":
		if appt.PatientID != claims.UserID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
			return
		}
	case "doctor":
		if appt.DoctorID != claims.UserID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
			return
		}
	}

	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id, appointment_id, actor_user_id, action, detail, created_at
		FROM appointment_activities
		WHERE appointment_id = $1
		ORDER BY created_at ASC
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	activities := make([]models.AppointmentActivity, 0)
	for rows.Next() {
		var row models.AppointmentActivity
		if err := rows.Scan(&row.ID, &row.AppointmentID, &row.ActorUserID, &row.Action, &row.Detail, &row.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		activities = append(activities, row)
	}
	if rows.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"activities": activities})
}

func (h *AppointmentHandler) ListAvailableDoctors(c *gin.Context) {
	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id, email
		FROM users
		WHERE role = 'doctor'
		ORDER BY email ASC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	doctors := make([]DoctorOption, 0)
	for rows.Next() {
		var doctor DoctorOption
		if err := rows.Scan(&doctor.ID, &doctor.Email); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		doctors = append(doctors, doctor)
	}
	if rows.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"doctors": doctors})
}

func (h *AppointmentHandler) UpdateStatus(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment id"})
		return
	}

	var req UpdateAppointmentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status != models.AppointmentStatusApproved && status != models.AppointmentStatusRejected {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be approved or rejected"})
		return
	}

	tx, err := db.Pool.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer tx.Rollback(c.Request.Context())

	var appt models.Appointment
	err = tx.QueryRow(c.Request.Context(), `
		UPDATE appointments
		SET status = $1, updated_at = NOW()
		WHERE id = $2 AND doctor_id = $3
		RETURNING id, patient_id, doctor_id, scheduled_at, reason, status, created_at, updated_at
	`, status, id, claims.UserID).
		Scan(&appt.ID, &appt.PatientID, &appt.DoctorID, &appt.ScheduledAt, &appt.Reason, &appt.Status, &appt.CreatedAt, &appt.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	_, err = tx.Exec(c.Request.Context(), `
		INSERT INTO appointment_activities (appointment_id, actor_user_id, action, detail)
		VALUES ($1, $2, $3, $4)
	`, appt.ID, claims.UserID, "status_changed", status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, appt)
}
