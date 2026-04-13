package handlers

import (
	"database/sql"
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

func scanAppointment(scanner interface {
	Scan(dest ...any) error
}, appt *models.Appointment) error {
	var pending sql.NullTime
	err := scanner.Scan(
		&appt.ID, &appt.PatientID, &appt.DoctorID, &appt.ScheduledAt, &appt.Reason, &appt.Status,
		&pending, &appt.CreatedAt, &appt.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if pending.Valid {
		t := pending.Time.UTC()
		appt.PendingScheduledAt = &t
	} else {
		appt.PendingScheduledAt = nil
	}
	return nil
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
	if !req.ScheduledAt.After(time.Now().UTC()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scheduled_at must be in the future"})
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
	err = scanAppointment(db.Pool.QueryRow(c.Request.Context(), `
		INSERT INTO appointments (patient_id, doctor_id, scheduled_at, reason, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, patient_id, doctor_id, scheduled_at, reason, status, pending_scheduled_at, created_at, updated_at
	`, claims.UserID, doctorID, req.ScheduledAt, reason, models.AppointmentStatusPending), &appt)
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
		SELECT id, patient_id, doctor_id, scheduled_at, reason, status, pending_scheduled_at, created_at, updated_at
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
		if err := scanAppointment(rows, &appt); err != nil {
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
		SELECT id, patient_id, doctor_id, scheduled_at, reason, status, pending_scheduled_at, created_at, updated_at
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
		if err := scanAppointment(rows, &appt); err != nil {
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
	err = scanAppointment(db.Pool.QueryRow(c.Request.Context(), `
		SELECT id, patient_id, doctor_id, scheduled_at, reason, status, pending_scheduled_at, created_at, updated_at
		FROM appointments
		WHERE id = $1
	`, id), &appt)
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
// Response items include actor_email when the actor user row exists (joined from users; empty string if missing).
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
	err = scanAppointment(db.Pool.QueryRow(c.Request.Context(), `
		SELECT id, patient_id, doctor_id, scheduled_at, reason, status, pending_scheduled_at, created_at, updated_at
		FROM appointments
		WHERE id = $1
	`, id), &appt)
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
		SELECT a.id, a.appointment_id, a.actor_user_id, COALESCE(u.email, ''), a.action, a.detail, a.created_at
		FROM appointment_activities a
		LEFT JOIN users u ON u.id = a.actor_user_id
		WHERE a.appointment_id = $1
		ORDER BY a.created_at ASC
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	activities := make([]models.AppointmentActivity, 0)
	for rows.Next() {
		var row models.AppointmentActivity
		if err := rows.Scan(&row.ID, &row.AppointmentID, &row.ActorUserID, &row.ActorEmail, &row.Action, &row.Detail, &row.CreatedAt); err != nil {
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
		WHERE role = 'doctor' AND COALESCE(active, TRUE)
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
	if claims.Role != "doctor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
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

	next := strings.ToLower(strings.TrimSpace(req.Status))
	allowed := map[string]bool{
		models.AppointmentStatusApproved:  true,
		models.AppointmentStatusRejected:  true,
		models.AppointmentStatusCompleted: true,
		models.AppointmentStatusNoShow:    true,
		models.AppointmentStatusCancelled: true,
	}
	if !allowed[next] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status for doctor update"})
		return
	}

	tx, err := db.Pool.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer tx.Rollback(c.Request.Context())

	var cur models.Appointment
	err = scanAppointment(tx.QueryRow(c.Request.Context(), `
		SELECT id, patient_id, doctor_id, scheduled_at, reason, status, pending_scheduled_at, created_at, updated_at
		FROM appointments
		WHERE id = $1 AND doctor_id = $2
		FOR UPDATE
	`, id, claims.UserID), &cur)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	detail := next
	var appt models.Appointment

	switch cur.Status {
	case models.AppointmentStatusRescheduleRequested:
		if next == models.AppointmentStatusApproved && cur.PendingScheduledAt != nil {
			err = scanAppointment(tx.QueryRow(c.Request.Context(), `
				UPDATE appointments
				SET scheduled_at = $1, pending_scheduled_at = NULL, status = $2, updated_at = NOW()
				WHERE id = $3 AND doctor_id = $4
				RETURNING id, patient_id, doctor_id, scheduled_at, reason, status, pending_scheduled_at, created_at, updated_at
			`, *cur.PendingScheduledAt, models.AppointmentStatusApproved, id, claims.UserID), &appt)
			detail = "reschedule_approved"
		} else if next == models.AppointmentStatusRejected {
			err = scanAppointment(tx.QueryRow(c.Request.Context(), `
				UPDATE appointments
				SET pending_scheduled_at = NULL, status = $1, updated_at = NOW()
				WHERE id = $2 AND doctor_id = $3
				RETURNING id, patient_id, doctor_id, scheduled_at, reason, status, pending_scheduled_at, created_at, updated_at
			`, models.AppointmentStatusApproved, id, claims.UserID), &appt)
			detail = "reschedule_rejected"
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "approve or reject the reschedule request"})
			return
		}
	case models.AppointmentStatusPending:
		if next != models.AppointmentStatusApproved && next != models.AppointmentStatusRejected {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status must be approved or rejected"})
			return
		}
		err = scanAppointment(tx.QueryRow(c.Request.Context(), `
			UPDATE appointments SET status = $1, updated_at = NOW()
			WHERE id = $2 AND doctor_id = $3
			RETURNING id, patient_id, doctor_id, scheduled_at, reason, status, pending_scheduled_at, created_at, updated_at
		`, next, id, claims.UserID), &appt)
	case models.AppointmentStatusApproved:
		if next != models.AppointmentStatusCompleted && next != models.AppointmentStatusNoShow && next != models.AppointmentStatusCancelled {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status must be completed, no_show, or cancelled"})
			return
		}
		err = scanAppointment(tx.QueryRow(c.Request.Context(), `
			UPDATE appointments SET status = $1, updated_at = NOW()
			WHERE id = $2 AND doctor_id = $3
			RETURNING id, patient_id, doctor_id, scheduled_at, reason, status, pending_scheduled_at, created_at, updated_at
		`, next, id, claims.UserID), &appt)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "appointment cannot be updated in this state"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	_, err = tx.Exec(c.Request.Context(), `
		INSERT INTO appointment_activities (appointment_id, actor_user_id, action, detail)
		VALUES ($1, $2, $3, $4)
	`, appt.ID, claims.UserID, "status_changed", detail)
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

type cancelAppointmentRequest struct {
	Reason string `json:"reason"`
}

// PatientCancel sets status to cancelled for pending or approved future appointments.
func (h *AppointmentHandler) PatientCancel(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok || claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment id"})
		return
	}
	var req cancelAppointmentRequest
	_ = c.ShouldBindJSON(&req)

	var cur models.Appointment
	err = scanAppointment(db.Pool.QueryRow(c.Request.Context(), `
		SELECT id, patient_id, doctor_id, scheduled_at, reason, status, pending_scheduled_at, created_at, updated_at
		FROM appointments WHERE id = $1 AND patient_id = $2
	`, id, claims.UserID), &cur)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if cur.Status != models.AppointmentStatusPending && cur.Status != models.AppointmentStatusApproved && cur.Status != models.AppointmentStatusRescheduleRequested {
		c.JSON(http.StatusBadRequest, gin.H{"error": "appointment cannot be cancelled"})
		return
	}
	if !cur.ScheduledAt.After(time.Now().UTC()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot cancel a past appointment"})
		return
	}

	detail := strings.TrimSpace(req.Reason)
	if detail == "" {
		detail = "cancelled_by_patient"
	} else {
		detail = "cancelled_by_patient: " + detail
	}

	tx, err := db.Pool.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer tx.Rollback(c.Request.Context())

	var appt models.Appointment
	err = scanAppointment(tx.QueryRow(c.Request.Context(), `
		UPDATE appointments
		SET status = $1, pending_scheduled_at = NULL, updated_at = NOW()
		WHERE id = $2 AND patient_id = $3
		RETURNING id, patient_id, doctor_id, scheduled_at, reason, status, pending_scheduled_at, created_at, updated_at
	`, models.AppointmentStatusCancelled, id, claims.UserID), &appt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	_, err = tx.Exec(c.Request.Context(), `
		INSERT INTO appointment_activities (appointment_id, actor_user_id, action, detail)
		VALUES ($1, $2, $3, $4)
	`, appt.ID, claims.UserID, "status_changed", detail)
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

type requestRescheduleRequest struct {
	ScheduledAt time.Time `json:"scheduled_at" binding:"required"`
	Reason      string    `json:"reason"`
}

// PatientRequestReschedule sets reschedule_requested and stores proposed time in pending_scheduled_at.
func (h *AppointmentHandler) PatientRequestReschedule(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok || claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment id"})
		return
	}
	var req requestRescheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if !req.ScheduledAt.After(time.Now().UTC()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scheduled_at must be in the future"})
		return
	}

	var cur models.Appointment
	err = scanAppointment(db.Pool.QueryRow(c.Request.Context(), `
		SELECT id, patient_id, doctor_id, scheduled_at, reason, status, pending_scheduled_at, created_at, updated_at
		FROM appointments WHERE id = $1 AND patient_id = $2
	`, id, claims.UserID), &cur)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if cur.Status != models.AppointmentStatusPending && cur.Status != models.AppointmentStatusApproved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot reschedule this appointment"})
		return
	}
	if !cur.ScheduledAt.After(time.Now().UTC()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot reschedule a past appointment"})
		return
	}

	detail := "reschedule_requested"
	if r := strings.TrimSpace(req.Reason); r != "" {
		detail += ": " + r
	}

	tx, err := db.Pool.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer tx.Rollback(c.Request.Context())

	var appt models.Appointment
	err = scanAppointment(tx.QueryRow(c.Request.Context(), `
		UPDATE appointments
		SET status = $1, pending_scheduled_at = $2, updated_at = NOW()
		WHERE id = $3 AND patient_id = $4
		RETURNING id, patient_id, doctor_id, scheduled_at, reason, status, pending_scheduled_at, created_at, updated_at
	`, models.AppointmentStatusRescheduleRequested, req.ScheduledAt, id, claims.UserID), &appt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	_, err = tx.Exec(c.Request.Context(), `
		INSERT INTO appointment_activities (appointment_id, actor_user_id, action, detail)
		VALUES ($1, $2, $3, $4)
	`, appt.ID, claims.UserID, "status_changed", detail)
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
