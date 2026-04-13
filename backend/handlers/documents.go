package handlers

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/jackc/pgx/v5"
)

const maxDocumentBytes = 10 << 20 // 10 MiB

type DocumentsHandler struct{}

func NewDocumentsHandler() *DocumentsHandler {
	return &DocumentsHandler{}
}

type patientDocumentJSON struct {
	ID          uuid.UUID `json:"id"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type,omitempty"`
	SizeBytes   int64     `json:"size_bytes"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// List GET /api/documents — patient only; own documents.
func (h *DocumentsHandler) List(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only patients can access documents"})
		return
	}
	ctx := c.Request.Context()
	rows, err := db.Pool.Query(ctx, `
		SELECT id, filename, content_type, size_bytes, status, created_at
		FROM patient_documents
		WHERE patient_id = $1
		ORDER BY created_at DESC
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to list documents"})
		return
	}
	defer rows.Close()

	var out []patientDocumentJSON
	for rows.Next() {
		var r patientDocumentJSON
		if err := rows.Scan(&r.ID, &r.Filename, &r.ContentType, &r.SizeBytes, &r.Status, &r.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to list documents"})
			return
		}
		out = append(out, r)
	}
	if out == nil {
		out = []patientDocumentJSON{}
	}
	c.JSON(http.StatusOK, gin.H{"documents": out})
}

// Upload POST /api/documents — multipart field "file".
func (h *DocumentsHandler) Upload(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only patients can upload documents"})
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required (multipart field name: file)"})
		return
	}
	if fh.Size > maxDocumentBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("file too large (max %d bytes)", maxDocumentBytes)})
		return
	}

	src, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read file"})
		return
	}
	defer src.Close()

	data, err := io.ReadAll(io.LimitReader(src, maxDocumentBytes+1))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to read file"})
		return
	}
	if int64(len(data)) > maxDocumentBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("file too large (max %d bytes)", maxDocumentBytes)})
		return
	}

	filename := sanitizeFilename(fh.Filename)
	if filename == "" {
		filename = "upload"
	}
	contentType := fh.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = sniffContentType(filename, data)
	}

	ctx := c.Request.Context()
	var doc patientDocumentJSON
	err = db.Pool.QueryRow(ctx, `
		INSERT INTO patient_documents (patient_id, filename, content_type, size_bytes, status, body)
		VALUES ($1, $2, $3, $4, 'ready', $5)
		RETURNING id, filename, content_type, size_bytes, status, created_at
	`, claims.UserID, filename, contentType, int64(len(data)), data).Scan(
		&doc.ID, &doc.Filename, &doc.ContentType, &doc.SizeBytes, &doc.Status, &doc.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save document"})
		return
	}

	c.JSON(http.StatusCreated, doc)
}

// Download GET /api/documents/:id/download — patient only; owner only.
func (h *DocumentsHandler) Download(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only patients can download documents"})
		return
	}
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document id"})
		return
	}
	ctx := c.Request.Context()
	var filename, contentType string
	var body []byte
	err = db.Pool.QueryRow(ctx, `
		SELECT filename, content_type, body
		FROM patient_documents
		WHERE id = $1 AND patient_id = $2
	`, docID, claims.UserID).Scan(&filename, &contentType, &body)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load document"})
		return
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	disposition := "inline"
	if !isInlineFriendly(contentType) {
		disposition = "attachment"
	}
	c.Header("Content-Disposition", fmt.Sprintf(`%s; filename="%s"`, disposition, escapeFilename(filename)))
	c.Data(http.StatusOK, contentType, body)
}

func sanitizeFilename(name string) string {
	name = path.Base(strings.TrimSpace(name))
	if name == "." || name == "/" {
		return ""
	}
	// prevent path tricks
	name = strings.ReplaceAll(name, "..", "")
	if utf8.RuneCountInString(name) > 255 {
		name = string([]rune(name)[:255])
	}
	return name
}

func escapeFilename(s string) string {
	s = strings.ReplaceAll(s, `"`, `'`)
	return s
}

func sniffContentType(filename string, head []byte) string {
	ext := strings.ToLower(path.Ext(filename))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		if len(head) >= 4 && head[0] == 0x25 && head[1] == 0x50 && head[2] == 0x44 && head[3] == 0x46 {
			return "application/pdf"
		}
		return "application/octet-stream"
	}
}

func isInlineFriendly(ct string) bool {
	switch ct {
	case "application/pdf", "image/png", "image/jpeg", "image/gif", "image/webp", "text/plain":
		return true
	default:
		return strings.HasPrefix(ct, "image/")
	}
}
