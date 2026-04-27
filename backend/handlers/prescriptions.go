package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/jackc/pgx/v5"
)

type PrescriptionsHandler struct{}

func NewPrescriptionsHandler() *PrescriptionsHandler {
	return &PrescriptionsHandler{}
}

func (PrescriptionsHandler) canAccessPatient(c *gin.Context, patientID uuid.UUID, claims *Claims) bool {
	if claims.Role == "admin" {
		return true
	}
	if claims.Role == "patient" && claims.UserID == patientID {
		return true
	}
	if claims.Role == "doctor" {
		var ok bool
		_ = db.Pool.QueryRow(c.Request.Context(),
			`SELECT EXISTS(SELECT 1 FROM appointments WHERE patient_id = $1 AND doctor_id = $2)`,
			patientID, claims.UserID,
		).Scan(&ok)
		return ok
	}
	return false
}

func (h *PrescriptionsHandler) ListByPatient(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	pid, err := uuid.Parse(strings.TrimSpace(c.Param("patientId")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient id"})
		return
	}
	if !h.canAccessPatient(c, pid, claims) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id, patient_id, doctor_id, medication_name, dosage, frequency, duration_days, instructions, status, created_at::text
		FROM prescriptions WHERE patient_id = $1 ORDER BY created_at DESC
	`, pid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	type row struct {
		ID              uuid.UUID `json:"id"`
		PatientID       uuid.UUID `json:"patient_id"`
		DoctorID        uuid.UUID `json:"doctor_id"`
		MedicationName  string    `json:"medication_name"`
		Dosage          string    `json:"dosage"`
		Frequency       string    `json:"frequency"`
		DurationDays    int       `json:"duration_days"`
		Instructions    string    `json:"instructions"`
		Status          string    `json:"status"`
		CreatedAt       string    `json:"created_at"`
	}
	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.ID, &r.PatientID, &r.DoctorID, &r.MedicationName, &r.Dosage, &r.Frequency, &r.DurationDays, &r.Instructions, &r.Status, &r.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		out = append(out, r)
	}
	if out == nil {
		out = []row{}
	}
	c.JSON(http.StatusOK, gin.H{"prescriptions": out})
}

type createPrescriptionBody struct {
	MedicationName string `json:"medication_name" binding:"required"`
	Dosage         string `json:"dosage" binding:"required"`
	Frequency      string `json:"frequency" binding:"required"`
	DurationDays   int    `json:"duration_days"`
	Instructions   string `json:"instructions"`
}

type updatePrescriptionBody struct {
	MedicationName *string `json:"medication_name"`
	Dosage         *string `json:"dosage"`
	Frequency      *string `json:"frequency"`
	DurationDays   *int    `json:"duration_days"`
	Instructions   *string `json:"instructions"`
}

func (h *PrescriptionsHandler) Create(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	pid, err := uuid.Parse(strings.TrimSpace(c.Param("patientId")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient id"})
		return
	}
	if claims.Role == "doctor" && !h.canAccessPatient(c, pid, claims) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	var body createPrescriptionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	docID := claims.UserID

	var id uuid.UUID
	err = db.Pool.QueryRow(c.Request.Context(), `
		INSERT INTO prescriptions (patient_id, doctor_id, medication_name, dosage, frequency, duration_days, instructions)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`, pid, docID, strings.TrimSpace(body.MedicationName), strings.TrimSpace(body.Dosage), strings.TrimSpace(body.Frequency), body.DurationDays, strings.TrimSpace(body.Instructions)).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *PrescriptionsHandler) Update(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	id, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	var body updatePrescriptionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if body.MedicationName == nil && body.Dosage == nil && body.Frequency == nil &&
		body.DurationDays == nil && body.Instructions == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one field must be provided"})
		return
	}

	tx, err := db.Pool.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer func() { _ = tx.Rollback(c.Request.Context()) }()

	var doctorID, patientID uuid.UUID
	var medication, dosage, frequency, instructions, status string
	var duration int
	err = tx.QueryRow(c.Request.Context(), `
		SELECT doctor_id, patient_id, medication_name, dosage, frequency, duration_days, instructions, status
		FROM prescriptions
		WHERE id = $1
	`, id).Scan(&doctorID, &patientID, &medication, &dosage, &frequency, &duration, &instructions, &status)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if claims.Role == "doctor" && doctorID != claims.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	if status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only active prescriptions can be updated"})
		return
	}

	newMedication := strings.TrimSpace(medication)
	newDosage := strings.TrimSpace(dosage)
	newFrequency := strings.TrimSpace(frequency)
	newInstructions := strings.TrimSpace(instructions)
	newDuration := duration

	if body.MedicationName != nil {
		newMedication = strings.TrimSpace(*body.MedicationName)
	}
	if body.Dosage != nil {
		newDosage = strings.TrimSpace(*body.Dosage)
	}
	if body.Frequency != nil {
		newFrequency = strings.TrimSpace(*body.Frequency)
	}
	if body.DurationDays != nil {
		newDuration = *body.DurationDays
	}
	if body.Instructions != nil {
		newInstructions = strings.TrimSpace(*body.Instructions)
	}

	if newMedication == "" || newDosage == "" || newFrequency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "medication_name, dosage, and frequency must be non-empty"})
		return
	}
	if newDuration < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "duration_days must be >= 0"})
		return
	}

	_, err = tx.Exec(c.Request.Context(), `
		UPDATE prescriptions
		SET medication_name = $2,
			dosage = $3,
			frequency = $4,
			duration_days = $5,
			instructions = $6
		WHERE id = $1
	`, id, newMedication, newDosage, newFrequency, newDuration, newInstructions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	detail := fmt.Sprintf("Prescription updated: med=%s dosage=%s frequency=%s duration_days=%d",
		newMedication, newDosage, newFrequency, newDuration)
	if err := writeAudit(c.Request.Context(), tx, claims.UserID, "prescription_updated", "prescription", id.String(), detail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":              id,
		"patient_id":      patientID,
		"medication_name": newMedication,
		"dosage":          newDosage,
		"frequency":       newFrequency,
		"duration_days":   newDuration,
		"instructions":    newInstructions,
		"status":          "active",
	})
}

func (h *PrescriptionsHandler) Revoke(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	id, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	var doctorID, patientID uuid.UUID
	err = db.Pool.QueryRow(c.Request.Context(), `SELECT doctor_id, patient_id FROM prescriptions WHERE id = $1`, id).Scan(&doctorID, &patientID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	if claims.Role == "doctor" && doctorID != claims.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	_, err = db.Pool.Exec(c.Request.Context(), `UPDATE prescriptions SET status = 'revoked' WHERE id = $1 AND status = 'active'`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "patient_id": patientID, "status": "revoked"})
}
