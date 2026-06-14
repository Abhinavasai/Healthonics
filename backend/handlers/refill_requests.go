package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/jackc/pgx/v5"
)

type RefillRequestHandler struct{}

func NewRefillRequestHandler() *RefillRequestHandler {
	return &RefillRequestHandler{}
}

type refillRequest struct {
	ID             string     `json:"id"`
	PrescriptionID string     `json:"prescription_id"`
	PatientID      string     `json:"patient_id"`
	DoctorID       string     `json:"doctor_id"`
	Note           string     `json:"note"`
	Status         string     `json:"status"`
	ReviewedAt     *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	// Enriched fields
	MedicationName string `json:"medication_name,omitempty"`
	PatientEmail   string `json:"patient_email,omitempty"`
}

// PatientCreate — patient requests a refill on one of their prescriptions.
func (h *RefillRequestHandler) PatientCreate(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	prescriptionID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prescription id"})
		return
	}
	var body struct {
		Note string `json:"note"`
	}
	_ = c.ShouldBindJSON(&body)
	note := strings.TrimSpace(body.Note)
	if len([]rune(note)) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "note must be 500 characters or fewer"})
		return
	}

	ctx := c.Request.Context()
	// Validate prescription belongs to this patient and is active.
	var doctorID uuid.UUID
	var medicationName string
	err = db.Pool.QueryRow(ctx, `
		SELECT doctor_id, medication_name FROM prescriptions
		WHERE id = $1 AND patient_id = $2 AND status = 'active'
	`, prescriptionID, claims.UserID).Scan(&doctorID, &medicationName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Active prescription not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	// Prevent duplicate pending requests.
	var existing bool
	if err := db.Pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM prescription_refill_requests
		WHERE prescription_id = $1 AND patient_id = $2 AND status = 'pending')
	`, prescriptionID, claims.UserID).Scan(&existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if existing {
		c.JSON(http.StatusConflict, gin.H{"error": "A refill request is already pending for this prescription"})
		return
	}

	var id uuid.UUID
	if err := db.Pool.QueryRow(ctx, `
		INSERT INTO prescription_refill_requests (prescription_id, patient_id, doctor_id, note)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, prescriptionID, claims.UserID, doctorID, note).Scan(&id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id.String(), "medication_name": medicationName})
}

// PatientList — patient lists their own refill requests.
func (h *RefillRequestHandler) PatientList(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	ctx := c.Request.Context()
	rows, err := db.Pool.Query(ctx, `
		SELECT r.id::text, r.prescription_id::text, r.patient_id::text, r.doctor_id::text,
		       r.note, r.status, r.reviewed_at, r.created_at, p.medication_name
		FROM prescription_refill_requests r
		JOIN prescriptions p ON p.id = r.prescription_id
		WHERE r.patient_id = $1
		ORDER BY r.created_at DESC
		LIMIT 50
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	var out []refillRequest
	for rows.Next() {
		var r refillRequest
		if err := rows.Scan(&r.ID, &r.PrescriptionID, &r.PatientID, &r.DoctorID, &r.Note, &r.Status, &r.ReviewedAt, &r.CreatedAt, &r.MedicationName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if out == nil {
		out = []refillRequest{}
	}
	c.JSON(http.StatusOK, gin.H{"requests": out})
}

// DoctorList — doctor lists pending refill requests for their patients.
func (h *RefillRequestHandler) DoctorList(c *gin.Context) {
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
		SELECT r.id::text, r.prescription_id::text, r.patient_id::text, r.doctor_id::text,
		       r.note, r.status, r.reviewed_at, r.created_at, p.medication_name, u.email
		FROM prescription_refill_requests r
		JOIN prescriptions p ON p.id = r.prescription_id
		JOIN users u ON u.id = r.patient_id
		WHERE r.doctor_id = $1 AND r.status = 'pending'
		ORDER BY r.created_at ASC
		LIMIT 100
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	var out []refillRequest
	for rows.Next() {
		var r refillRequest
		if err := rows.Scan(&r.ID, &r.PrescriptionID, &r.PatientID, &r.DoctorID, &r.Note, &r.Status, &r.ReviewedAt, &r.CreatedAt, &r.MedicationName, &r.PatientEmail); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if out == nil {
		out = []refillRequest{}
	}
	c.JSON(http.StatusOK, gin.H{"requests": out})
}

// DoctorReview — doctor approves or denies a refill request.
func (h *RefillRequestHandler) DoctorReview(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	id, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid refill request id"})
		return
	}
	var body struct {
		Action string `json:"action" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "action required"})
		return
	}
	action := strings.ToLower(strings.TrimSpace(body.Action))
	if action != "approved" && action != "denied" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "action must be 'approved' or 'denied'"})
		return
	}

	ctx := c.Request.Context()
	res, err := db.Pool.Exec(ctx, `
		UPDATE prescription_refill_requests
		SET status = $1, reviewed_at = NOW(), reviewed_by = $2
		WHERE id = $3 AND doctor_id = $4 AND status = 'pending'
	`, action, claims.UserID, id, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if res.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pending refill request not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "status": action})
}
