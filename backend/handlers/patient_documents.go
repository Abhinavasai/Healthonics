package handlers

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
)

const maxDocumentUploadBytes = 10 * 1024 * 1024

// PatientDocumentsHandler handles patient uploads of clinical documents for doctors to review.
type PatientDocumentsHandler struct{}

func NewPatientDocumentsHandler() *PatientDocumentsHandler {
	return &PatientDocumentsHandler{}
}

type patientUploadResponse struct {
	ID       uuid.UUID `json:"id"`
	Filename string    `json:"filename"`
}

type patientDocumentRow struct {
	ID          uuid.UUID `json:"id"`
	Filename    string    `json:"filename"`
	SizeBytes   int64     `json:"size_bytes"`
	ContentType string    `json:"content_type"`
	CreatedAt   time.Time `json:"created_at"`
}

// Upload accepts multipart form field "file" and stores metadata (and bytes) for doctor visibility
// when an appointment exists between this patient and a doctor.
func (h *PatientDocumentsHandler) Upload(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}

	if err := c.Request.ParseMultipartForm(maxDocumentUploadBytes + 1024); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid multipart form"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil || file == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	if file.Size > maxDocumentUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large (max 10MB)"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read file"})
		return
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, maxDocumentUploadBytes+1))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read file"})
		return
	}
	if len(data) > maxDocumentUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large (max 10MB)"})
		return
	}

	filename := filepath.Base(file.Filename)
	if filename == "" || filename == "." {
		filename = "upload.bin"
	}
	contentType := file.Header.Get("Content-Type")
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}

	var id uuid.UUID
	err = db.Pool.QueryRow(c.Request.Context(), `
		INSERT INTO patient_documents (patient_id, filename, content_type, size_bytes, file_data, summary_status)
		VALUES ($1, $2, $3, $4, $5, 'none')
		RETURNING id
	`, claims.UserID, filename, contentType, len(data), data).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save document"})
		return
	}

	c.JSON(http.StatusCreated, patientUploadResponse{ID: id, Filename: filename})
}

// List returns all documents uploaded by the authenticated patient.
func (h *PatientDocumentsHandler) List(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id, filename, size_bytes, content_type, created_at
		FROM patient_documents
		WHERE patient_id = $1
		ORDER BY created_at DESC
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load documents"})
		return
	}
	defer rows.Close()

	out := make([]patientDocumentRow, 0)
	for rows.Next() {
		var r patientDocumentRow
		if err := rows.Scan(&r.ID, &r.Filename, &r.SizeBytes, &r.ContentType, &r.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load documents"})
			return
		}
		out = append(out, r)
	}
	c.JSON(http.StatusOK, gin.H{"documents": out})
}

// Download streams a patient's own uploaded file.
func (h *PatientDocumentsHandler) Download(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document id"})
		return
	}
	var filename, contentType string
	var payload []byte
	err = db.Pool.QueryRow(c.Request.Context(), `
		SELECT filename, content_type, file_data
		FROM patient_documents
		WHERE id = $1 AND patient_id = $2
	`, id, claims.UserID).Scan(&filename, &contentType, &payload)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Data(http.StatusOK, contentType, payload)
}
