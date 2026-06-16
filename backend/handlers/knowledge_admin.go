package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/jackc/pgx/v5"
)

type KnowledgeAdminHandler struct{}

func NewKnowledgeAdminHandler() *KnowledgeAdminHandler {
	return &KnowledgeAdminHandler{}
}

func (h *KnowledgeAdminHandler) List(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}
	ctx := c.Request.Context()
	rows, err := db.Pool.Query(ctx, `
		SELECT id::text, title, left(body, 200) AS excerpt, created_at::text, updated_at::text,
			current_version,
			review_interval_days,
			COALESCE(last_reviewed_at::text, ''),
			GREATEST(0, FLOOR(EXTRACT(EPOCH FROM (NOW() - COALESCE(last_reviewed_at, updated_at))) / 86400))::int
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
		ID                 string `json:"id"`
		Title              string `json:"title"`
		Excerpt            string `json:"excerpt"`
		CreatedAt          string `json:"created_at"`
		UpdatedAt          string `json:"updated_at"`
		CurrentVersion     int    `json:"current_version"`
		ReviewIntervalDays int    `json:"review_interval_days"`
		LastReviewedAt     string `json:"last_reviewed_at,omitempty"`
		DaysSinceReview    int    `json:"days_since_review"`
		HealthScore        int    `json:"health_score"`
		IsStale            bool   `json:"is_stale"`
	}
	var out []row
	for rows.Next() {
		var r row
		var lastRev string
		if err := rows.Scan(&r.ID, &r.Title, &r.Excerpt, &r.CreatedAt, &r.UpdatedAt,
			&r.CurrentVersion, &r.ReviewIntervalDays, &lastRev, &r.DaysSinceReview); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		if lastRev != "" {
			r.LastReviewedAt = lastRev
		}
		r.HealthScore = KnowledgeHealthScore(r.DaysSinceReview, r.ReviewIntervalDays)
		r.IsStale = KnowledgeIsStale(r.DaysSinceReview, r.ReviewIntervalDays)
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if out == nil {
		out = []row{}
	}
	c.JSON(http.StatusOK, gin.H{"documents": out})
}

type knowledgeCreateBody struct {
	Title              string `json:"title" binding:"required"`
	Body               string `json:"body" binding:"required"`
	ReviewIntervalDays *int   `json:"review_interval_days"`
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
	if len([]rune(title)) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title must be 500 characters or fewer"})
		return
	}
	if len([]rune(text)) > 100000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body must be 100000 characters or fewer"})
		return
	}
	interval := 180
	if body.ReviewIntervalDays != nil && *body.ReviewIntervalDays > 0 && *body.ReviewIntervalDays <= 3650 {
		interval = *body.ReviewIntervalDays
	}

	ctx := c.Request.Context()
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var docID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO knowledge_docs (title, body, created_by, current_version, review_interval_days)
		VALUES ($1, $2, $3, 1, $4)
		RETURNING id
	`, title, text, claims.UserID, interval).Scan(&docID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO knowledge_doc_versions (document_id, version, title, body, created_by)
		VALUES ($1, 1, $2, $3, $4)
	`, docID, title, text, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": docID.String(), "current_version": 1})
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
	ctx := c.Request.Context()
	var title, body string
	var updated string
	var curVer, intervalDays int
	var lastRevStr string
	var days int
	err = db.Pool.QueryRow(ctx, `
		SELECT title, body, updated_at::text, current_version, review_interval_days,
			COALESCE(last_reviewed_at::text, ''),
			GREATEST(0, FLOOR(EXTRACT(EPOCH FROM (NOW() - COALESCE(last_reviewed_at, updated_at))) / 86400))::int
		FROM knowledge_docs WHERE id = $1
	`, id).Scan(&title, &body, &updated, &curVer, &intervalDays, &lastRevStr, &days)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	resp := gin.H{
		"id":                   id.String(),
		"title":                title,
		"body":                 body,
		"updated_at":           updated,
		"current_version":      curVer,
		"review_interval_days": intervalDays,
		"days_since_review":    days,
		"health_score":         KnowledgeHealthScore(days, intervalDays),
		"is_stale":             KnowledgeIsStale(days, intervalDays),
	}
	if lastRevStr != "" {
		resp["last_reviewed_at"] = lastRevStr
	}

	c.JSON(http.StatusOK, resp)
}

type knowledgePatchBody struct {
	Title              *string `json:"title"`
	Body               *string `json:"body"`
	ReviewIntervalDays *int    `json:"review_interval_days"`
}

func (h *KnowledgeAdminHandler) Patch(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}
	var body knowledgePatchBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if body.Title == nil && body.Body == nil && body.ReviewIntervalDays == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one of title, body, review_interval_days required"})
		return
	}

	ctx := c.Request.Context()
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var curTitle, curBody string
	var curVer, intervalDays int
	err = tx.QueryRow(ctx, `
		SELECT title, body, current_version, review_interval_days FROM knowledge_docs WHERE id = $1 FOR UPDATE
	`, id).Scan(&curTitle, &curBody, &curVer, &intervalDays)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	newTitle := curTitle
	newBody := curBody
	newInterval := intervalDays
	contentChanged := false
	if body.Title != nil {
		t := strings.TrimSpace(*body.Title)
		if t != "" && t != curTitle {
			newTitle = t
			contentChanged = true
		}
	}
	if body.Body != nil {
		t := strings.TrimSpace(*body.Body)
		if t != "" && t != curBody {
			newBody = t
			contentChanged = true
		}
	}
	if body.ReviewIntervalDays != nil {
		v := *body.ReviewIntervalDays
		if v > 0 && v <= 3650 {
			newInterval = v
		}
	}

	nextVer := curVer
	if contentChanged {
		nextVer = curVer + 1
	}

	if _, err := tx.Exec(ctx, `
		UPDATE knowledge_docs SET
			title = $1,
			body = $2,
			current_version = $3,
			review_interval_days = $4,
			updated_at = NOW()
		WHERE id = $5
	`, newTitle, newBody, nextVer, newInterval, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	if contentChanged {
		if _, err := tx.Exec(ctx, `
			INSERT INTO knowledge_doc_versions (document_id, version, title, body, created_by)
			VALUES ($1, $2, $3, $4, $5)
		`, id, nextVer, newTitle, newBody, claims.UserID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id.String(), "current_version": nextVer})
}

func (h *KnowledgeAdminHandler) Review(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}
	id, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}
	ctx := c.Request.Context()
	tag, err := db.Pool.Exec(ctx, `
		UPDATE knowledge_docs SET last_reviewed_at = NOW(), updated_at = NOW() WHERE id = $1
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *KnowledgeAdminHandler) ListVersions(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}
	docID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}
	ctx := c.Request.Context()
	var exists bool
	if err := db.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM knowledge_docs WHERE id = $1)`, docID).Scan(&exists); err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	rows, err := db.Pool.Query(ctx, `
		SELECT version, title, left(body, 200) AS excerpt, created_at::text, created_by::text
		FROM knowledge_doc_versions
		WHERE document_id = $1
		ORDER BY version DESC
		LIMIT 50
	`, docID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	type vrow struct {
		Version   int    `json:"version"`
		Title     string `json:"title"`
		Excerpt   string `json:"excerpt"`
		CreatedAt string `json:"created_at"`
		CreatedBy string `json:"created_by"`
	}
	var out []vrow
	for rows.Next() {
		var r vrow
		if err := rows.Scan(&r.Version, &r.Title, &r.Excerpt, &r.CreatedAt, &r.CreatedBy); err != nil {
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
		out = []vrow{}
	}
	c.JSON(http.StatusOK, gin.H{"versions": out})
}
