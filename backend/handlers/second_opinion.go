package handlers

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
)

type SecondOpinionHandler struct{}

func NewSecondOpinionHandler() *SecondOpinionHandler { return &SecondOpinionHandler{} }

// List  GET /api/second-opinions
func (h *SecondOpinionHandler) List(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Doctors only"})
		return
	}

	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id::text, title, anonymized, specialty, status, created_at::text,
		       (SELECT COUNT(*) FROM second_opinion_replies r WHERE r.case_id = so.id) AS reply_count
		FROM second_opinions so
		ORDER BY created_at DESC
		LIMIT 50
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	type item struct {
		ID         string `json:"id"`
		Title      string `json:"title"`
		Anonymized string `json:"anonymized"`
		Specialty  string `json:"specialty"`
		Status     string `json:"status"`
		CreatedAt  string `json:"created_at"`
		ReplyCount int    `json:"reply_count"`
	}
	var items []item
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.ID, &it.Title, &it.Anonymized, &it.Specialty, &it.Status, &it.CreatedAt, &it.ReplyCount); err == nil {
			items = append(items, it)
		}
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cases": items})
}

// Create  POST /api/second-opinions
func (h *SecondOpinionHandler) Create(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Doctors only"})
		return
	}

	var req struct {
		Title      string `json:"title"`
		Anonymized string `json:"anonymized"`
		Specialty  string `json:"specialty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	title := strings.TrimSpace(req.Title)
	anonymized := strings.TrimSpace(req.Anonymized)
	if title == "" || len([]rune(title)) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title required (max 200 chars)"})
		return
	}
	if anonymized == "" || len([]rune(anonymized)) > 5000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "anonymized case description required (max 5000 chars)"})
		return
	}

	var id string
	if err := db.Pool.QueryRow(c.Request.Context(), `
		INSERT INTO second_opinions (author_id, title, anonymized, specialty)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text
	`, claims.UserID, title, anonymized, strings.TrimSpace(req.Specialty)).Scan(&id); err != nil {
		slog.Error("second opinion: create failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// GetReplies  GET /api/second-opinions/:id/replies
func (h *SecondOpinionHandler) GetReplies(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Doctors only"})
		return
	}

	caseID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case id"})
		return
	}

	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT r.id::text, r.body, r.created_at::text,
		       CASE WHEN r.author_id = $2 THEN 'You' ELSE 'Dr. ' || LEFT(r.author_id::text, 8) END AS author_label
		FROM second_opinion_replies r
		WHERE r.case_id = $1
		ORDER BY r.created_at ASC
	`, caseID, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	type reply struct {
		ID          string `json:"id"`
		Body        string `json:"body"`
		CreatedAt   string `json:"created_at"`
		AuthorLabel string `json:"author_label"`
	}
	var replies []reply
	for rows.Next() {
		var r reply
		if err := rows.Scan(&r.ID, &r.Body, &r.CreatedAt, &r.AuthorLabel); err == nil {
			replies = append(replies, r)
		}
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"replies": replies})
}

// PostReply  POST /api/second-opinions/:id/replies
func (h *SecondOpinionHandler) PostReply(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Doctors only"})
		return
	}

	caseID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case id"})
		return
	}

	var req struct {
		Body string `json:"body"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Body) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body required"})
		return
	}
	if len([]rune(req.Body)) > 3000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reply must be 3000 characters or fewer"})
		return
	}

	var id string
	if err := db.Pool.QueryRow(c.Request.Context(), `
		INSERT INTO second_opinion_replies (case_id, author_id, body)
		VALUES ($1, $2, $3)
		RETURNING id::text
	`, caseID, claims.UserID, strings.TrimSpace(req.Body)).Scan(&id); err != nil {
		slog.Error("second opinion: reply failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// Resolve  PATCH /api/second-opinions/:id/resolve
func (h *SecondOpinionHandler) Resolve(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Doctors only"})
		return
	}

	caseID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case id"})
		return
	}

	// Only the author can resolve.
	res, err := db.Pool.Exec(c.Request.Context(), `
		UPDATE second_opinions SET status = 'resolved'
		WHERE id = $1 AND author_id = $2
	`, caseID, claims.UserID)
	if err != nil {
		slog.Error("second opinion: resolve failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if res.RowsAffected() == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot resolve: not the author or case not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
