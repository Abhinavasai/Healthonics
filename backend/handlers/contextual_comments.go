package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/jackc/pgx/v5"
)

const (
	ContextTypeRecord    = "record"
	ContextTypeDocument  = "document"
	ContextTypeKnowledge = "knowledge_doc"
	ContextVisInternal   = "internal"
	ContextVisCareTeam   = "care_team"
	ContextVisPatient    = "patient_visible"
)

type ContextualCommentsHandler struct{}

func NewContextualCommentsHandler() *ContextualCommentsHandler {
	return &ContextualCommentsHandler{}
}

type postContextualCommentBody struct {
	Body       string `json:"body" binding:"required"`
	Visibility string `json:"visibility"`
}

func normalizeContextType(v string) string {
	return strings.TrimSpace(strings.ToLower(v))
}

func validContextType(v string) bool {
	switch v {
	case ContextTypeRecord, ContextTypeDocument, ContextTypeKnowledge:
		return true
	default:
		return false
	}
}

func resolveContextVisibility(role, requested string) (string, bool) {
	s := strings.TrimSpace(strings.ToLower(requested))
	if s == "" {
		s = ContextVisPatient
	}
	switch s {
	case ContextVisInternal, ContextVisCareTeam, ContextVisPatient:
	default:
		return "", false
	}
	if role == "patient" && s != ContextVisPatient {
		return "", false
	}
	return s, true
}

func (h *ContextualCommentsHandler) canAccessContext(c *gin.Context, claims *Claims, contextType string, contextID uuid.UUID) (bool, error) {
	if claims.Role == "admin" {
		if contextType == ContextTypeKnowledge {
			return true, nil
		}
	}

	switch contextType {
	case ContextTypeKnowledge:
		if claims.Role != "admin" {
			return false, nil
		}
		var exists bool
		err := db.Pool.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM knowledge_docs WHERE id = $1)`, contextID).Scan(&exists)
		return exists, err
	case ContextTypeRecord:
		var patientID, doctorID uuid.UUID
		err := db.Pool.QueryRow(c.Request.Context(), `SELECT patient_id, doctor_id FROM prescriptions WHERE id = $1`, contextID).Scan(&patientID, &doctorID)
		if err != nil {
			if err == pgx.ErrNoRows {
				return false, nil
			}
			return false, err
		}
		if claims.Role == "patient" {
			return claims.UserID == patientID, nil
		}
		if claims.Role == "doctor" {
			return claims.UserID == doctorID, nil
		}
		return claims.Role == "admin", nil
	case ContextTypeDocument:
		var patientID uuid.UUID
		err := db.Pool.QueryRow(c.Request.Context(), `SELECT patient_id FROM patient_documents WHERE id = $1`, contextID).Scan(&patientID)
		if err != nil {
			if err == pgx.ErrNoRows {
				return false, nil
			}
			return false, err
		}
		if claims.Role == "patient" {
			return claims.UserID == patientID, nil
		}
		if claims.Role == "doctor" {
			var linked bool
			if err := db.Pool.QueryRow(c.Request.Context(), `
				SELECT EXISTS (
					SELECT 1 FROM appointments
					WHERE patient_id = $1 AND doctor_id = $2
				)
			`, patientID, claims.UserID).Scan(&linked); err != nil {
				return false, err
			}
			return linked, nil
		}
		return claims.Role == "admin", nil
	default:
		return false, nil
	}
}

func visibilityListForRole(role string) []string {
	if role == "patient" {
		return []string{ContextVisPatient}
	}
	return []string{ContextVisInternal, ContextVisCareTeam, ContextVisPatient}
}

func (h *ContextualCommentsHandler) List(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	contextType := normalizeContextType(c.Param("contextType"))
	if !validContextType(contextType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid context_type"})
		return
	}
	contextID, err := uuid.Parse(strings.TrimSpace(c.Param("contextId")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid context_id"})
		return
	}
	allowed, accessErr := h.canAccessContext(c, claims, contextType, contextID)
	if accessErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	vis := visibilityListForRole(claims.Role)
	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id::text, context_type, context_id::text, author_user_id::text, body, visibility, created_at::text
		FROM contextual_comments
		WHERE context_type = $1
		  AND context_id = $2
		  AND visibility = ANY($3)
		ORDER BY created_at ASC
	`, contextType, contextID, vis)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	type row struct {
		ID          string `json:"id"`
		ContextType string `json:"context_type"`
		ContextID   string `json:"context_id"`
		AuthorID    string `json:"author_user_id"`
		Body        string `json:"body"`
		Visibility  string `json:"visibility"`
		CreatedAt   string `json:"created_at"`
	}
	out := make([]row, 0, 16)
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.ID, &r.ContextType, &r.ContextID, &r.AuthorID, &r.Body, &r.Visibility, &r.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		out = append(out, r)
	}
	c.JSON(http.StatusOK, gin.H{"comments": out})
}

func (h *ContextualCommentsHandler) Create(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	contextType := normalizeContextType(c.Param("contextType"))
	if !validContextType(contextType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid context_type"})
		return
	}
	contextID, err := uuid.Parse(strings.TrimSpace(c.Param("contextId")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid context_id"})
		return
	}
	var req postContextualCommentBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	body := strings.TrimSpace(req.Body)
	if body == "" || len(body) > 4000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body required, max 4000 chars"})
		return
	}
	visibility, okVis := resolveContextVisibility(claims.Role, req.Visibility)
	if !okVis {
		c.JSON(http.StatusBadRequest, gin.H{"error": "visibility must be internal, care_team, or patient_visible"})
		return
	}
	allowed, accessErr := h.canAccessContext(c, claims, contextType, contextID)
	if accessErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	var id string
	if err := db.Pool.QueryRow(c.Request.Context(), `
		INSERT INTO contextual_comments(context_type, context_id, author_user_id, body, visibility)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text
	`, contextType, contextID, claims.UserID, body, visibility).Scan(&id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "visibility": visibility})
}
