package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

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
		if err := db.Pool.QueryRow(c.Request.Context(),
			`SELECT EXISTS(
				SELECT 1 FROM appointments
				WHERE patient_id = $1
				  AND doctor_id = $2
				  AND status IN ('approved', 'completed')
				  AND scheduled_at >= NOW() - INTERVAL '2 years'
			)`,
			patientID, claims.UserID,
		).Scan(&ok); err != nil {
			log.Printf("prescriptions: canAccessPatient DB error for doctor %s: %v", claims.UserID, err)
			return false
		}
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
		ID             uuid.UUID `json:"id"`
		PatientID      uuid.UUID `json:"patient_id"`
		DoctorID       uuid.UUID `json:"doctor_id"`
		MedicationName string    `json:"medication_name"`
		Dosage         string    `json:"dosage"`
		Frequency      string    `json:"frequency"`
		DurationDays   int       `json:"duration_days"`
		Instructions   string    `json:"instructions"`
		Status         string    `json:"status"`
		CreatedAt      string    `json:"created_at"`
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
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
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

type revokePrescriptionBody struct {
	Reason  string `json:"reason"`
	Confirm string `json:"confirm"`
}

func reminderTimesForFrequency(frequency string) []string {
	f := strings.ToLower(strings.TrimSpace(frequency))
	switch {
	case strings.Contains(f, "3"), strings.Contains(f, "thrice"), strings.Contains(f, "three"):
		return []string{"08:00", "14:00", "20:00"}
	case strings.Contains(f, "2"), strings.Contains(f, "twice"), strings.Contains(f, "two"):
		return []string{"08:00", "20:00"}
	default:
		return []string{"08:00"}
	}
}

func buildReminderSchedule(start time.Time, durationDays int, times []string) []time.Time {
	if durationDays <= 0 {
		durationDays = 1
	}
	if len(times) == 0 {
		times = []string{"08:00"}
	}
	base := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	out := make([]time.Time, 0, durationDays*len(times))
	for day := 0; day < durationDays; day++ {
		for _, hhmm := range times {
			h, m := 8, 0
			_, _ = fmt.Sscanf(hhmm, "%d:%d", &h, &m)
			out = append(out, base.AddDate(0, 0, day).Add(time.Duration(h)*time.Hour).Add(time.Duration(m)*time.Minute))
		}
	}
	return out
}

func escapePDFText(s string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "(", "\\(", ")", "\\)")
	return replacer.Replace(s)
}

func buildPrescriptionPDF(medication, dosage, frequency, instructions, status string, durationDays int, prescribedAt time.Time) []byte {
	lines := []string{
		"BT",
		"/F1 18 Tf",
		"50 790 Td",
		fmt.Sprintf("(%s) Tj", escapePDFText("Healthonyx Prescription")),
		"0 -28 Td",
		"/F1 12 Tf",
		fmt.Sprintf("(%s) Tj", escapePDFText("Medication: "+medication)),
		"0 -18 Td",
		fmt.Sprintf("(%s) Tj", escapePDFText("Dosage: "+dosage)),
		"0 -18 Td",
		fmt.Sprintf("(%s) Tj", escapePDFText("Frequency: "+frequency)),
		"0 -18 Td",
		fmt.Sprintf("(%s) Tj", escapePDFText(fmt.Sprintf("Duration: %d day(s)", durationDays))),
		"0 -18 Td",
		fmt.Sprintf("(%s) Tj", escapePDFText("Status: "+status)),
		"0 -18 Td",
		fmt.Sprintf("(%s) Tj", escapePDFText("Prescribed at: "+prescribedAt.UTC().Format(time.RFC3339))),
	}
	if strings.TrimSpace(instructions) != "" {
		lines = append(lines, "0 -18 Td")
		lines = append(lines, fmt.Sprintf("(%s) Tj", escapePDFText("Instructions: "+instructions)))
	}
	lines = append(lines, "ET")
	stream := strings.Join(lines, "\n")

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := []int{}
	writeObj := func(obj string) {
		offsets = append(offsets, buf.Len())
		buf.WriteString(obj)
	}

	writeObj("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")
	writeObj("2 0 obj\n<< /Type /Pages /Count 1 /Kids [3 0 R] >>\nendobj\n")
	writeObj("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>\nendobj\n")
	writeObj("4 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")
	writeObj(fmt.Sprintf("5 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n", len(stream), stream))

	xrefStart := buf.Len()
	buf.WriteString("xref\n")
	buf.WriteString(fmt.Sprintf("0 %d\n", len(offsets)+1))
	buf.WriteString("0000000000 65535 f \n")
	for _, off := range offsets {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", off))
	}
	buf.WriteString("trailer\n")
	buf.WriteString(fmt.Sprintf("<< /Size %d /Root 1 0 R >>\n", len(offsets)+1))
	buf.WriteString("startxref\n")
	buf.WriteString(fmt.Sprintf("%d\n", xrefStart))
	buf.WriteString("%%EOF")
	return buf.Bytes()
}

func schedulePrescriptionReminders(ctx *gin.Context, patientID uuid.UUID, medication, dosage, frequency string, durationDays int) error {
	times := reminderTimesForFrequency(frequency)
	schedule := buildReminderSchedule(time.Now().UTC(), durationDays, times)
	title := fmt.Sprintf("Medication reminder: %s", medication)
	body := fmt.Sprintf("Take %s %s as prescribed (%s).", dosage, medication, frequency)
	for _, at := range schedule {
		_, err := db.Pool.Exec(ctx.Request.Context(), `
			INSERT INTO notifications (user_id, title, body, channel, status, scheduled_for)
			VALUES ($1, $2, $3, 'in_app', 'pending', $4)
		`, patientID, title, body, at)
		if err != nil {
			return err
		}
	}
	return nil
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
	if err := schedulePrescriptionReminders(c, pid, strings.TrimSpace(body.MedicationName), strings.TrimSpace(body.Dosage), strings.TrimSpace(body.Frequency), body.DurationDays); err != nil {
		log.Printf("prescriptions: failed to schedule reminders for patient %s: %v", pid, err)
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
	if _, notifErr := db.Pool.Exec(c.Request.Context(), `
		INSERT INTO notifications (user_id, title, body, channel, status, scheduled_for)
		VALUES ($1, $2, $3, 'in_app', 'pending', NOW())
	`, patientID, "Prescription updated", fmt.Sprintf("Your prescription for %s has been updated.", newMedication)); notifErr != nil {
		log.Printf("prescriptions: failed to queue update notification for patient %s: %v", patientID, notifErr)
	}
	if err := schedulePrescriptionReminders(c, patientID, newMedication, newDosage, newFrequency, newDuration); err != nil {
		log.Printf("prescriptions: failed to schedule reminders after update for patient %s: %v", patientID, err)
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
	var body revokePrescriptionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if err := validateHighRiskGuardrail(body.Reason, body.Confirm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	tx, err := db.Pool.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer func() { _ = tx.Rollback(c.Request.Context()) }()

	_, err = tx.Exec(c.Request.Context(), `UPDATE prescriptions SET status = 'revoked' WHERE id = $1 AND status = 'active'`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if err := writeAudit(c.Request.Context(), tx, claims.UserID, "prescription_revoked", "prescription", id.String(), appendAuditReason("prescription revoked", body.Reason)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	// Queue notification outside the transaction so a delivery failure doesn't roll back the revocation.
	if _, notifErr := db.Pool.Exec(c.Request.Context(), `
		INSERT INTO notifications (user_id, title, body, channel, status, scheduled_for)
		VALUES ($1, $2, $3, 'in_app', 'pending', NOW())
	`, patientID, "Prescription revoked", "One of your prescriptions has been revoked by your care team."); notifErr != nil {
		log.Printf("prescriptions: failed to queue revocation notification for patient %s: %v", patientID, notifErr)
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "patient_id": patientID, "status": "revoked"})
}

func (h *PrescriptionsHandler) DownloadPDF(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	var patientID uuid.UUID
	var medication, dosage, frequency, instructions, status string
	var durationDays int
	var createdAt time.Time
	err = db.Pool.QueryRow(c.Request.Context(), `
		SELECT patient_id, medication_name, dosage, frequency, duration_days, instructions, status, created_at
		FROM prescriptions
		WHERE id = $1
	`, id).Scan(&patientID, &medication, &dosage, &frequency, &durationDays, &instructions, &status, &createdAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	if !h.canAccessPatient(c, patientID, claims) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	pdf := buildPrescriptionPDF(medication, dosage, frequency, instructions, status, durationDays, createdAt)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"prescription-%s.pdf\"", id.String()))
	c.Data(http.StatusOK, "application/pdf", pdf)
}
