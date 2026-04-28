package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/healthonyx/backend/db"
)

type AdminAIRuntimeHandler struct{}

type adminAIRuntimeSettings struct {
	OllamaEnabled      bool   `json:"ollama_enabled"`
	FallbackEnabled    bool   `json:"fallback_enabled"`
	RateLimitEnabled   bool   `json:"rate_limit_enabled"`
	RateLimitPerMinute int    `json:"rate_limit_per_minute"`
	CacheEnabled       bool   `json:"cache_enabled"`
	CacheTTLSeconds    int    `json:"cache_ttl_seconds"`
	OllamaModel        string `json:"ollama_model"`
	UpdatedAt          string `json:"updated_at"`
	UpdatedBy          string `json:"updated_by"`
}

type adminAIObservability struct {
	QueuedJobs24h     int     `json:"queued_jobs_24h"`
	CompletedJobs24h  int     `json:"completed_jobs_24h"`
	FailedJobs24h     int     `json:"failed_jobs_24h"`
	AvgLatencySeconds float64 `json:"avg_latency_seconds"`
	CacheReadyCount   int     `json:"cache_ready_count"`
	PendingDocsCount  int     `json:"pending_docs_count"`
	FailedDocsCount   int     `json:"failed_docs_count"`
}

type adminAIEvalResult struct {
	ModelName        string  `json:"model_name"`
	SamplesEvaluated int     `json:"samples_evaluated"`
	SuccessRate      float64 `json:"success_rate"`
	QualityScore     float64 `json:"quality_score"`
}

type updateAIRuntimeSettingsRequest struct {
	OllamaEnabled      bool   `json:"ollama_enabled"`
	FallbackEnabled    bool   `json:"fallback_enabled"`
	RateLimitEnabled   bool   `json:"rate_limit_enabled"`
	RateLimitPerMinute int    `json:"rate_limit_per_minute"`
	CacheEnabled       bool   `json:"cache_enabled"`
	CacheTTLSeconds    int    `json:"cache_ttl_seconds"`
	OllamaModel        string `json:"ollama_model"`
}

func NewAdminAIRuntimeHandler() *AdminAIRuntimeHandler {
	return &AdminAIRuntimeHandler{}
}

func (h *AdminAIRuntimeHandler) GetObservability(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}

	settings, err := loadAIRuntimeSettings(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	obs, err := loadAIObservability(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settings":      settings,
		"observability": obs,
	})
}

func (h *AdminAIRuntimeHandler) UpdateSettings(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}

	var req updateAIRuntimeSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if req.RateLimitPerMinute < 1 || req.RateLimitPerMinute > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rate_limit_per_minute must be between 1 and 10000"})
		return
	}
	if req.CacheTTLSeconds < 30 || req.CacheTTLSeconds > 604800 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cache_ttl_seconds must be between 30 and 604800"})
		return
	}
	model := strings.TrimSpace(req.OllamaModel)
	if model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ollama_model is required"})
		return
	}

	_, err := db.Pool.Exec(c.Request.Context(), `
		UPDATE admin_ai_runtime_settings
		SET
			ollama_enabled = $1,
			fallback_enabled = $2,
			rate_limit_enabled = $3,
			rate_limit_per_minute = $4,
			cache_enabled = $5,
			cache_ttl_seconds = $6,
			ollama_model = $7,
			updated_by = $8,
			updated_at = NOW()
		WHERE id = TRUE
	`, req.OllamaEnabled, req.FallbackEnabled, req.RateLimitEnabled, req.RateLimitPerMinute, req.CacheEnabled, req.CacheTTLSeconds, model, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	h.GetObservability(c)
}

func (h *AdminAIRuntimeHandler) RunEval(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}

	settings, err := loadAIRuntimeSettings(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	var samples, doneCount int
	if err := db.Pool.QueryRow(c.Request.Context(), `
		SELECT
			COUNT(*)::int,
			COUNT(*) FILTER (WHERE status = 'done')::int
		FROM (
			SELECT status
			FROM document_summary_jobs
			ORDER BY created_at DESC
			LIMIT 50
		) s
	`).Scan(&samples, &doneCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	result := adminAIEvalResult{
		ModelName:        settings.OllamaModel,
		SamplesEvaluated: samples,
		SuccessRate:      0,
		QualityScore:     0,
	}
	if samples > 0 {
		result.SuccessRate = float64(doneCount) / float64(samples)
		result.QualityScore = result.SuccessRate * 100
	}
	c.JSON(http.StatusOK, gin.H{"eval": result})
}

func loadAIRuntimeSettings(c *gin.Context) (adminAIRuntimeSettings, error) {
	var out adminAIRuntimeSettings
	err := db.Pool.QueryRow(c.Request.Context(), `
		SELECT
			ollama_enabled,
			fallback_enabled,
			rate_limit_enabled,
			rate_limit_per_minute,
			cache_enabled,
			cache_ttl_seconds,
			ollama_model,
			updated_at::text,
			COALESCE(updated_by::text, '')
		FROM admin_ai_runtime_settings
		WHERE id = TRUE
	`).Scan(
		&out.OllamaEnabled,
		&out.FallbackEnabled,
		&out.RateLimitEnabled,
		&out.RateLimitPerMinute,
		&out.CacheEnabled,
		&out.CacheTTLSeconds,
		&out.OllamaModel,
		&out.UpdatedAt,
		&out.UpdatedBy,
	)
	return out, err
}

func loadAIObservability(c *gin.Context) (adminAIObservability, error) {
	var out adminAIObservability
	err := db.Pool.QueryRow(c.Request.Context(), `
		SELECT
			COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '24 hours')::int AS queued_24h,
			COUNT(*) FILTER (WHERE status = 'done' AND finished_at >= NOW() - INTERVAL '24 hours')::int AS done_24h,
			COUNT(*) FILTER (WHERE status = 'failed' AND finished_at >= NOW() - INTERVAL '24 hours')::int AS failed_24h,
			COALESCE(AVG(EXTRACT(EPOCH FROM (finished_at - started_at))) FILTER (
				WHERE status IN ('done', 'failed') AND started_at IS NOT NULL AND finished_at IS NOT NULL
			), 0)::float8 AS avg_seconds
		FROM document_summary_jobs
	`).Scan(&out.QueuedJobs24h, &out.CompletedJobs24h, &out.FailedJobs24h, &out.AvgLatencySeconds)
	if err != nil {
		return out, err
	}
	if err := db.Pool.QueryRow(c.Request.Context(), `
		SELECT
			COUNT(*) FILTER (WHERE summary_status = 'ready')::int,
			COUNT(*) FILTER (WHERE summary_status = 'pending')::int,
			COUNT(*) FILTER (WHERE summary_status = 'failed')::int
		FROM patient_documents
	`).Scan(&out.CacheReadyCount, &out.PendingDocsCount, &out.FailedDocsCount); err != nil {
		return out, err
	}
	return out, nil
}
