package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
)

type AppointmentCommentsHandler struct{}

func NewAppointmentCommentsHandler() *AppointmentCommentsHandler {
	return &AppointmentCommentsHandler{}
}

func (h *AppointmentCommentsHandler) canViewAppointment(c *gin.Context, apptID uuid.UUID, claims *Claims) bool {
	var patientID, doctorID uuid.UUID
	err := db.Pool.QueryRow(c.Request.Context(), `
		SELECT patient_id, doctor_id FROM appointments WHERE id = $1
	`, apptID).Scan(&patientID, &doctorID)
	if err != nil {
		return false
	}
	if claims.Role == "admin" {
		return true
	}
	if claims.Role == "patient" && claims.UserID == patientID {
		return true
	}
	return claims.Role == "doctor" && claims.UserID == doctorID
}

func (h *AppointmentCommentsHandler) List(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment id"})
		return
	}
	if !h.canViewAppointment(c, id, claims) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id::text, author_user_id::text, body, created_at::text
		FROM appointment_comments
		WHERE appointment_id = $1
		ORDER BY created_at ASC
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	type row struct {
		ID       string `json:"id"`
		AuthorID string `json:"author_user_id"`
		Body     string `json:"body"`
		Created  string `json:"created_at"`
	}
	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.ID, &r.AuthorID, &r.Body, &r.Created); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		out = append(out, r)
	}
	if out == nil {
		out = []row{}
	}
	c.JSON(http.StatusOK, gin.H{"comments": out})
}

type postCommentBody struct {
	Body string `json:"body" binding:"required"`
}

func (h *AppointmentCommentsHandler) Create(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment id"})
		return
	}
	if !h.canViewAppointment(c, id, claims) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	var body postCommentBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	text := strings.TrimSpace(body.Body)
	if text == "" || len(text) > 4000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body required, max 4000 chars"})
		return
	}

	var cid string
	err = db.Pool.QueryRow(c.Request.Context(), `
		INSERT INTO appointment_comments (appointment_id, author_user_id, body)
		VALUES ($1, $2, $3) RETURNING id::text
	`, id, claims.UserID, text).Scan(&cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": cid})
}
