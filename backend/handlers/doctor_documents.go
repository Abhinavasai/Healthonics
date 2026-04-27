package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/jackc/pgx/v5"
)

// DoctorDocumentsHandler — list/detail/summarize for patient_documents visible to treating doctors (Sprint 3 F4 MVP).
type DoctorDocumentsHandler struct{}

func NewDoctorDocumentsHandler() *DoctorDocumentsHandler {
	return &DoctorDocumentsHandler{}
}

func (h *DoctorDocumentsHandler) canAccessDoc(ctx context.Context, doctorID, patientID uuid.UUID) bool {
	var n int
	err := db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM appointments
		WHERE doctor_id = $1 AND patient_id = $2
	`, doctorID, patientID).Scan(&n)
	return err == nil && n > 0
}

// List GET /api/doctor/documents
func (h *DoctorDocumentsHandler) List(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Doctor role required"})
		return
	}
	ctx := c.Request.Context()
	rows, err := db.Pool.Query(ctx, `
		SELECT d.id, d.filename, u.email, d.created_at,
			COALESCE(NULLIF(d.summary_status, ''), 'none') AS summary_status
		FROM patient_documents d
		JOIN users u ON u.id = d.patient_id
		WHERE d.patient_id IN (
			SELECT DISTINCT patient_id FROM appointments WHERE doctor_id = $1
		)
		ORDER BY d.created_at DESC
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to list documents"})
		return
	}
	defer rows.Close()

	type row struct {
		ID            string    `json:"id"`
		Filename      string    `json:"filename"`
		PatientEmail  string    `json:"patient_email"`
		CreatedAt     time.Time `json:"created_at"`
		SummaryStatus string    `json:"summary_status"`
	}
	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.ID, &r.Filename, &r.PatientEmail, &r.CreatedAt, &r.SummaryStatus); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to list documents"})
			return
		}
		out = append(out, r)
	}
	if out == nil {
		out = []row{}
	}
	c.JSON(http.StatusOK, gin.H{"documents": out})
}

// Get GET /api/doctor/documents/:id
func (h *DoctorDocumentsHandler) Get(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Doctor role required"})
		return
	}
	docID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document id"})
		return
	}
	ctx := c.Request.Context()
	var patientID uuid.UUID
	var filename, contentType string
	var sizeBytes int64
	var createdAt time.Time
	var summary *string
	var summaryStatus, summaryError *string
	err = db.Pool.QueryRow(ctx, `
		SELECT patient_id, filename, content_type, size_bytes, created_at, summary,
			COALESCE(NULLIF(summary_status, ''), 'none'), summary_error
		FROM patient_documents WHERE id = $1
	`, docID).Scan(&patientID, &filename, &contentType, &sizeBytes, &createdAt, &summary, &summaryStatus, &summaryError)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load document"})
		return
	}
	if !h.canAccessDoc(ctx, claims.UserID, patientID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this patient document"})
		return
	}
	var patientEmail string
	_ = db.Pool.QueryRow(ctx, `SELECT email FROM users WHERE id = $1`, patientID).Scan(&patientEmail)

	ss := "none"
	if summaryStatus != nil && *summaryStatus != "" {
		ss = *summaryStatus
	}
	out := gin.H{
		"id":             docID.String(),
		"filename":       filename,
		"patient_id":     patientID.String(),
		"patient_email":  patientEmail,
		"size_bytes":     sizeBytes,
		"content_type":   contentType,
		"created_at":     createdAt.Format(time.RFC3339Nano),
		"summary":        nil,
		"summary_status": ss,
	}
	if summary != nil {
		out["summary"] = *summary
	}
	if summaryError != nil {
		out["summary_error"] = *summaryError
	}
	c.JSON(http.StatusOK, out)
}

// Summarize POST /api/doctor/documents/:id/summarize — stub async: pending then ready (no external LLM).
func (h *DoctorDocumentsHandler) Summarize(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Doctor role required"})
		return
	}
	docID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document id"})
		return
	}
	ctx := c.Request.Context()
	var patientID uuid.UUID
	var filename string
	var sizeBytes int64
	var contentType string
	var body []byte
	err = db.Pool.QueryRow(ctx, `
		SELECT patient_id, filename, size_bytes, content_type, body
		FROM patient_documents
		WHERE id = $1
	`, docID).Scan(&patientID, &filename, &sizeBytes, &contentType, &body)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load document"})
		return
	}
	if !h.canAccessDoc(ctx, claims.UserID, patientID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this patient document"})
		return
	}

	_, _ = db.Pool.Exec(ctx, `
		UPDATE patient_documents
		SET summary_status = 'pending', summary_error = NULL
		WHERE id = $1
	`, docID)

	go h.completeStubSummary(docID, filename, sizeBytes, contentType, body)

	c.JSON(http.StatusAccepted, gin.H{"status": "pending"})
}

func (h *DoctorDocumentsHandler) completeStubSummary(docID uuid.UUID, filename string, sizeBytes int64, contentType string, body []byte) {
	time.Sleep(800 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	extracted, extractErr := extractDocumentText(contentType, filename, body)
	if extractErr != nil {
		msg := extractErr.Error()
		switch {
		case errors.Is(extractErr, errUnsupportedDocument):
			msg = "summary extraction is only supported for PDF and image documents"
		case errors.Is(extractErr, errNoExtractableText):
			msg = "could not extract readable text from document"
		}
		_, _ = db.Pool.Exec(ctx, `
			UPDATE patient_documents
			SET summary_status = 'failed', summary_error = $2
			WHERE id = $1
		`, docID, msg)
		return
	}

	text := buildSummaryFromExtractedText(filename, sizeBytes, extracted)
	_, err := db.Pool.Exec(ctx, `
		UPDATE patient_documents
		SET summary = $2, summary_status = 'ready', summary_error = NULL
		WHERE id = $1 AND summary_status = 'pending'
	`, docID, text)
	if err != nil {
		_, _ = db.Pool.Exec(ctx, `
			UPDATE patient_documents
			SET summary_status = 'failed', summary_error = $2
			WHERE id = $1
		`, docID, err.Error())
	}
}
