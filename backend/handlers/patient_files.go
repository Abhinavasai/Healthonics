package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
)

const patientFileMaxBytes = 5 << 20 // 5 MiB (matches DB check)

var allowedPatientFileMimes = map[string]bool{
	"application/pdf": true,
	"image/jpeg":      true,
	"image/png":       true,
}

type PatientFilesHandler struct {
	Root string
}

func NewPatientFilesHandler(root string) *PatientFilesHandler {
	if root == "" {
		root = "data/uploads"
	}
	_ = os.MkdirAll(root, 0o750)
	return &PatientFilesHandler{Root: root}
}

func (h *PatientFilesHandler) canAccessPatient(c *gin.Context, patientID uuid.UUID, claims *Claims) bool {
	if claims.Role == "admin" {
		return true
	}
	if claims.Role == "patient" && claims.UserID == patientID {
		return true
	}
	if claims.Role == "doctor" {
		var ok bool
		err := db.Pool.QueryRow(c.Request.Context(),
			`SELECT EXISTS(SELECT 1 FROM appointments WHERE patient_id = $1 AND doctor_id = $2)`,
			patientID, claims.UserID,
		).Scan(&ok)
		return err == nil && ok
	}
	return false
}

func (h *PatientFilesHandler) List(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	patientID, err := uuid.Parse(strings.TrimSpace(c.Param("patientId")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient id"})
		return
	}
	if !h.canAccessPatient(c, patientID, claims) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id, patient_id, uploaded_by, description, original_name, mime_type, byte_size, created_at::text
		FROM patient_files
		WHERE patient_id = $1
		ORDER BY created_at DESC
	`, patientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	type row struct {
		ID           uuid.UUID `json:"id"`
		PatientID    uuid.UUID `json:"patient_id"`
		UploadedBy   uuid.UUID `json:"uploaded_by"`
		Description  string    `json:"description"`
		OriginalName string    `json:"original_name"`
		MimeType     string    `json:"mime_type"`
		ByteSize     int64     `json:"byte_size"`
		CreatedAt    string    `json:"created_at"`
	}
	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.ID, &r.PatientID, &r.UploadedBy, &r.Description, &r.OriginalName, &r.MimeType, &r.ByteSize, &r.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		out = append(out, r)
	}
	if out == nil {
		out = []row{}
	}
	c.JSON(http.StatusOK, gin.H{"files": out})
}

func (h *PatientFilesHandler) Upload(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	patientID, err := uuid.Parse(strings.TrimSpace(c.Param("patientId")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient id"})
		return
	}
	if !h.canAccessPatient(c, patientID, claims) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file field required"})
		return
	}
	src, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not read file"})
		return
	}
	defer src.Close()

	body, err := io.ReadAll(io.LimitReader(src, patientFileMaxBytes+1))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not read file"})
		return
	}
	if len(body) == 0 || len(body) > patientFileMaxBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file must be 1 byte to 5MB"})
		return
	}
	mime := http.DetectContentType(body)
	if !allowedPatientFileMimes[mime] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only PDF, JPEG, or PNG allowed"})
		return
	}

	stored := uuid.New().String()
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if ext != ".pdf" && ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		switch mime {
		case "application/pdf":
			ext = ".pdf"
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		default:
			ext = ""
		}
	}
	storedName := stored + ext
	destPath := filepath.Join(h.Root, storedName)
	if err := os.WriteFile(destPath, body, 0o640); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not store file"})
		return
	}

	desc := strings.TrimSpace(c.PostForm("description"))
	if len(desc) > 2000 {
		desc = desc[:2000]
	}

	size := int64(len(body))
	var id uuid.UUID
	err = db.Pool.QueryRow(c.Request.Context(), `
		INSERT INTO patient_files (patient_id, uploaded_by, description, original_name, stored_name, mime_type, byte_size)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, patientID, claims.UserID, desc, fh.Filename, storedName, mime, size).Scan(&id)
	if err != nil {
		_ = os.Remove(destPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":            id,
		"patient_id":    patientID,
		"uploaded_by":   claims.UserID,
		"description":   desc,
		"original_name": fh.Filename,
		"mime_type":     mime,
		"byte_size":     size,
	})
}

func (h *PatientFilesHandler) Download(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file id"})
		return
	}

	var patientID uuid.UUID
	var storedName, orig, mime string
	err = db.Pool.QueryRow(c.Request.Context(), `
		SELECT patient_id, stored_name, original_name, mime_type
		FROM patient_files WHERE id = $1
	`, id).Scan(&patientID, &storedName, &orig, &mime)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	if !h.canAccessPatient(c, patientID, claims) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	path := filepath.Join(h.Root, storedName)
	absRoot, err := filepath.Abs(h.Root)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	absPath, err := filepath.Abs(path)
	if err != nil || !strings.HasPrefix(absPath, absRoot) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid path"})
		return
	}
	c.Header("Content-Type", mime)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, strings.ReplaceAll(orig, `"`, ``)))
	c.File(path)
}
