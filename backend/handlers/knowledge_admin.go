package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
)

type KnowledgeAdminHandler struct{}

func NewKnowledgeAdminHandler() *KnowledgeAdminHandler {
	return &KnowledgeAdminHandler{}
}

func (h *KnowledgeAdminHandler) List(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}
	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id::text, title, left(body, 200) AS excerpt, created_at::text, updated_at::text
		FROM knowledge_docs
		ORDER BY updated_at DESC
		LIMIT 100
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	type row struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Excerpt   string `json:"excerpt"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}
	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.ID, &r.Title, &r.Excerpt, &r.CreatedAt, &r.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		out = append(out, r)
	}
	if out == nil {
		out = []row{}
	}
	c.JSON(http.StatusOK, gin.H{"documents": out})
}

type knowledgeCreateBody struct {
	Title string `json:"title" binding:"required"`
	Body  string `json:"body" binding:"required"`
}

func (h *KnowledgeAdminHandler) Create(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	var body knowledgeCreateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	title := strings.TrimSpace(body.Title)
	text := strings.TrimSpace(body.Body)
	if title == "" || text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title and body required"})
		return
	}

	var id string
	err := db.Pool.QueryRow(c.Request.Context(), `
		INSERT INTO knowledge_docs (title, body, created_by)
		VALUES ($1, $2, $3) RETURNING id::text
	`, title, text, claims.UserID).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *KnowledgeAdminHandler) Get(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}
	id, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}
	var title, body string
	var updated string
	err = db.Pool.QueryRow(c.Request.Context(), `
		SELECT title, body, updated_at::text FROM knowledge_docs WHERE id = $1
	`, id).Scan(&title, &body, &updated)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id.String(), "title": title, "body": body, "updated_at": updated})
}
