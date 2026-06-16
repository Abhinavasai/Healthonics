package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/jackc/pgx/v5"
)

type LabResultsHandler struct {
	ai *AIHealthHandler
}

func NewLabResultsHandler(ai *AIHealthHandler) *LabResultsHandler {
	return &LabResultsHandler{ai: ai}
}

type labResultRow struct {
	ID             string    `json:"id"`
	PatientID      string    `json:"patient_id"`
	OriginalName   string    `json:"original_name"`
	MimeType       string    `json:"mime_type"`
	ByteSize       int64     `json:"byte_size"`
	Summary        *string   `json:"summary,omitempty"`
	SummaryStatus  string    `json:"summary_status"`
	CreatedAt      time.Time `json:"created_at"`
}

// PatientList returns lab-related patient files (PDFs).
func (h *LabResultsHandler) PatientList(c *gin.Context) {
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
		SELECT id::text, patient_id::text, filename, content_type, size_bytes,
		       summary, summary_status, created_at
		FROM patient_documents
		WHERE patient_id = $1 AND (content_type = 'application/pdf' OR filename ILIKE '%.pdf')
		ORDER BY created_at DESC
		LIMIT 50
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	var out []labResultRow
	for rows.Next() {
		var r labResultRow
		if err := rows.Scan(&r.ID, &r.PatientID, &r.OriginalName, &r.MimeType, &r.ByteSize, &r.Summary, &r.SummaryStatus, &r.CreatedAt); err != nil {
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
		out = []labResultRow{}
	}
	c.JSON(http.StatusOK, gin.H{"lab_results": out})
}

// Interpret — runs AI interpretation on the text of an existing patient_document.
func (h *LabResultsHandler) Interpret(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	docIDRaw := strings.TrimSpace(c.Param("id"))
	docID, err := uuid.Parse(docIDRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document id"})
		return
	}
	ctx := c.Request.Context()

	// Load the document text (extracted summary is stored in summary field from document_summary_jobs).
	var summary *string
	var summaryStatus string
	var originalName string
	if err := db.Pool.QueryRow(ctx, `
		SELECT filename, summary, summary_status
		FROM patient_documents WHERE id = $1 AND patient_id = $2
	`, docID, claims.UserID).Scan(&originalName, &summary, &summaryStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	if summary == nil || *summary == "" {
		if summaryStatus == "pending" || summaryStatus == "processing" {
			c.JSON(http.StatusAccepted, gin.H{"message": "Document text extraction is still in progress, try again shortly"})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "No text could be extracted from this document"})
		return
	}

	systemMsg := `You are a medical AI assistant helping patients understand their lab results.
Given extracted text from a lab report:
1. Provide a plain-language summary of what the results mean.
2. Identify any values flagged as HIGH, LOW, or ABNORMAL and explain what they may indicate.
3. Note any values that require prompt attention.
4. Remind the patient to discuss results with their doctor.
Keep your response clear, compassionate, and non-alarmist.`

	userMsg := fmt.Sprintf("Lab report file: %s\n\nExtracted text:\n%s", originalName, truncate(*summary, 3000))

	reply, provider := h.ai.callAI(c, systemMsg, userMsg)
	if reply == "" {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"document_id":   docID,
		"original_name": originalName,
		"interpretation": reply,
		"provider":      provider,
	})
}

func truncate(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "…"
}
