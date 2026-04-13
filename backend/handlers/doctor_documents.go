package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/jackc/pgx/v5"
)

// DoctorDocumentsHandler serves GET list/detail and POST summarize for patient-uploaded docs.
type DoctorDocumentsHandler struct{}

func NewDoctorDocumentsHandler() *DoctorDocumentsHandler {
	return &DoctorDocumentsHandler{}
}

type doctorDocumentListRow struct {
	ID            uuid.UUID `json:"id"`
	Filename      string    `json:"filename"`
	PatientEmail  string    `json:"patient_email"`
	CreatedAt     time.Time `json:"created_at"`
	SummaryStatus string    `json:"summary_status"`
}

type doctorDocumentsListResponse struct {
	Documents []doctorDocumentListRow `json:"documents"`
}

type doctorDocumentDetailResponse struct {
	ID            uuid.UUID `json:"id"`
	Filename      string    `json:"filename"`
	PatientID     uuid.UUID `json:"patient_id"`
	PatientEmail  string    `json:"patient_email"`
	SizeBytes     int64     `json:"size_bytes"`
	ContentType   string    `json:"content_type"`
	CreatedAt     time.Time `json:"created_at"`
	Summary       *string   `json:"summary"`
	SummaryStatus string    `json:"summary_status"`
	SummaryError  *string   `json:"summary_error,omitempty"`
}

type summarizeResponse struct {
	Status string `json:"status"`
}

func (h *DoctorDocumentsHandler) doctorSeesDocument(ctx context.Context, doctorID, docID uuid.UUID) (bool, error) {
	var ok bool
	err := db.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM patient_documents pd
			WHERE pd.id = $1
			  AND EXISTS (
				SELECT 1 FROM appointments a
				WHERE a.patient_id = pd.patient_id AND a.doctor_id = $2
			  )
		)
	`, docID, doctorID).Scan(&ok)
	return ok, err
}

// List returns documents for patients who share an appointment with the logged-in doctor.
func (h *DoctorDocumentsHandler) List(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}

	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT pd.id, pd.filename, u.email, pd.created_at, pd.summary_status
		FROM patient_documents pd
		JOIN users u ON u.id = pd.patient_id
		WHERE EXISTS (
			SELECT 1 FROM appointments a
			WHERE a.patient_id = pd.patient_id AND a.doctor_id = $1
		)
		ORDER BY pd.created_at DESC
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	var out []doctorDocumentListRow
	for rows.Next() {
		var r doctorDocumentListRow
		if err := rows.Scan(&r.ID, &r.Filename, &r.PatientEmail, &r.CreatedAt, &r.SummaryStatus); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		out = append(out, r)
	}
	if out == nil {
		out = []doctorDocumentListRow{}
	}
	c.JSON(http.StatusOK, doctorDocumentsListResponse{Documents: out})
}

// Get returns one document if the doctor is allowed to view it.
func (h *DoctorDocumentsHandler) Get(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}

	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document id"})
		return
	}

	sees, err := h.doctorSeesDocument(c.Request.Context(), claims.UserID, docID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if !sees {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	var res doctorDocumentDetailResponse
	var summary, summaryErr *string
	err = db.Pool.QueryRow(c.Request.Context(), `
		SELECT pd.id, pd.filename, pd.patient_id, u.email, pd.size_bytes, pd.content_type, pd.created_at,
		       pd.summary, pd.summary_status, pd.summary_error
		FROM patient_documents pd
		JOIN users u ON u.id = pd.patient_id
		WHERE pd.id = $1
	`, docID).Scan(
		&res.ID, &res.Filename, &res.PatientID, &res.PatientEmail, &res.SizeBytes, &res.ContentType, &res.CreatedAt,
		&summary, &res.SummaryStatus, &summaryErr,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	res.Summary = summary
	res.SummaryError = summaryErr
	c.JSON(http.StatusOK, res)
}

// Summarize triggers async stub summarization (or returns current state).
func (h *DoctorDocumentsHandler) Summarize(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}

	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document id"})
		return
	}

	sees, err := h.doctorSeesDocument(c.Request.Context(), claims.UserID, docID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if !sees {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	var status, filename string
	err = db.Pool.QueryRow(c.Request.Context(), `
		SELECT summary_status, filename FROM patient_documents WHERE id = $1
	`, docID).Scan(&status, &filename)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	if status == "ready" {
		c.JSON(http.StatusOK, summarizeResponse{Status: "ready"})
		return
	}
	if status == "pending" {
		c.JSON(http.StatusOK, summarizeResponse{Status: "pending"})
		return
	}

	_, err = db.Pool.Exec(c.Request.Context(), `
		UPDATE patient_documents
		SET summary_status = 'pending', summary = NULL, summary_error = NULL
		WHERE id = $1 AND summary_status NOT IN ('pending', 'ready')
	`, docID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	docIDCopy := docID
	filenameCopy := filename
	go h.runStubSummarize(docIDCopy, filenameCopy)

	c.JSON(http.StatusOK, summarizeResponse{Status: "pending"})
}

func (h *DoctorDocumentsHandler) runStubSummarize(docID uuid.UUID, filename string) {
	ctx := context.Background()
	time.Sleep(1500 * time.Millisecond)

	// Demo: filenames containing "fail" produce a failed summary for UI testing.
	if strings.Contains(strings.ToLower(filename), "fail") {
		msg := "Stub summarizer could not process this file (demo: filename contains 'fail')."
		_, err := db.Pool.Exec(ctx, `
			UPDATE patient_documents
			SET summary_status = 'failed', summary = NULL, summary_error = $2
			WHERE id = $1 AND summary_status = 'pending'
		`, docID, msg)
		if err != nil {
			return
		}
		return
	}

	stub := "Summary (stub): This document appears to be a patient upload named \"" + filename +
		"\". Replace this text with real AI output when integrated."
	_, err := db.Pool.Exec(ctx, `
		UPDATE patient_documents
		SET summary_status = 'ready', summary = $2, summary_error = NULL
		WHERE id = $1 AND summary_status = 'pending'
	`, docID, stub)
	if err != nil {
		return
	}
}
