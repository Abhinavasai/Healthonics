package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	PromptVersion      string `json:"prompt_version"`
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

type adminSLOMetric struct {
	TargetPercent   float64 `json:"target_percent"`
	SuccessPercent  float64 `json:"success_percent"`
	ErrorBudgetUsed float64 `json:"error_budget_used_percent"`
	Errors24h       int     `json:"errors_24h"`
	Requests24h     int     `json:"requests_24h"`
	AvgLatencyMS    float64 `json:"avg_latency_ms"`
	P95LatencyMS    float64 `json:"p95_latency_ms"`
}

type adminSLOOverview struct {
	Notifications adminSLOMetric `json:"notifications"`
	AI            adminSLOMetric `json:"ai"`
	API           adminSLOMetric `json:"api"`
	Alerts        []gin.H        `json:"alerts"`
}

type adminAIEvalResult struct {
	ModelName        string  `json:"model_name"`
	SamplesEvaluated int     `json:"samples_evaluated"`
	SuccessRate      float64 `json:"success_rate"`
	QualityScore     float64 `json:"quality_score"`
	PromptVersion    string  `json:"prompt_version"`
	RunID            string  `json:"run_id,omitempty"`
}

type adminAIEvalRunSummary struct {
	ID              string  `json:"id"`
	ModelName       string  `json:"model_name"`
	PromptVersion   string  `json:"prompt_version"`
	FixtureCount    int     `json:"fixture_count"`
	PassedCount     int     `json:"passed_count"`
	SuccessRate     float64 `json:"success_rate"`
	QualityScore    float64 `json:"quality_score"`
	RuntimeMode     string  `json:"runtime_mode"`
	OllamaReachable bool    `json:"ollama_reachable"`
	ModelAvailable  bool    `json:"model_available"`
	StartedAt       string  `json:"started_at"`
	FinishedAt      string  `json:"finished_at"`
}

type updateAIRuntimeSettingsRequest struct {
	OllamaEnabled      bool   `json:"ollama_enabled"`
	FallbackEnabled    bool   `json:"fallback_enabled"`
	RateLimitEnabled   bool   `json:"rate_limit_enabled"`
	RateLimitPerMinute int    `json:"rate_limit_per_minute"`
	CacheEnabled       bool   `json:"cache_enabled"`
	CacheTTLSeconds    int    `json:"cache_ttl_seconds"`
	OllamaModel        string `json:"ollama_model"`
	PromptVersion      string `json:"prompt_version"`
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
	runtime := getOllamaRuntimeStatus(c.Request.Context(), settings)
	recentRuns, _ := loadRecentEvalRuns(c.Request.Context(), 5)

	c.JSON(http.StatusOK, gin.H{
		"settings":      settings,
		"observability": obs,
		"runtime":       runtime,
		"recent_evals":  recentRuns,
		"slo_overview":  loadSLOOverview(c),
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
	promptVersion := strings.TrimSpace(req.PromptVersion)
	if promptVersion == "" {
		promptVersion = "v1"
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
			prompt_version = $8,
			updated_by = $9,
			updated_at = NOW()
		WHERE id = TRUE
	`, req.OllamaEnabled, req.FallbackEnabled, req.RateLimitEnabled, req.RateLimitPerMinute, req.CacheEnabled, req.CacheTTLSeconds, model, promptVersion, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	h.GetObservability(c)
}

func (h *AdminAIRuntimeHandler) RunEval(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}

	settings, err := loadAIRuntimeSettings(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	runtime := getOllamaRuntimeStatus(c.Request.Context(), settings)
	caseResults := make([]aiEvalCaseResult, 0, len(aiEvalFixtures))
	passedCount := 0
	coverageSum := 0.0
	runtimeMode := "fallback_local"
	for _, fixture := range aiEvalFixtures {
		summary, mode, evalErr := GenerateSummaryWithAIRuntime(c.Request.Context(), settings, fixture.ID+".txt", int64(len(fixture.Input)), fixture.Input)
		if evalErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		if mode == "ollama" {
			runtimeMode = "ollama"
		}
		result := evaluateSummaryAgainstFixture(summary, fixture)
		caseResults = append(caseResults, result)
		coverageSum += result.KeywordCoverage
		if result.Passed {
			passedCount++
		}
	}
	samples := len(caseResults)
	successRate := 0.0
	qualityScore := 0.0
	if samples > 0 {
		successRate = float64(passedCount) / float64(samples)
		qualityScore = (coverageSum / float64(samples)) * 100
	}
	runID, persistErr := persistEvalRun(c.Request.Context(), settings, runtime, runtimeMode, claims.UserID, samples, passedCount, successRate, qualityScore, caseResults)
	if persistErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	result := adminAIEvalResult{
		ModelName:        settings.OllamaModel,
		SamplesEvaluated: samples,
		SuccessRate:      successRate,
		QualityScore:     qualityScore,
		PromptVersion:    settings.PromptVersion,
		RunID:            runID.String(),
	}
	c.JSON(http.StatusOK, gin.H{
		"eval":    result,
		"runtime": runtime,
		"cases":   caseResults,
	})
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
			prompt_version,
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
		&out.PromptVersion,
		&out.UpdatedAt,
		&out.UpdatedBy,
	)
	return out, err
}

func LoadAIRuntimeSettingsForWorker(ctx context.Context) (adminAIRuntimeSettings, error) {
	var out adminAIRuntimeSettings
	err := db.Pool.QueryRow(ctx, `
		SELECT
			ollama_enabled,
			fallback_enabled,
			rate_limit_enabled,
			rate_limit_per_minute,
			cache_enabled,
			cache_ttl_seconds,
			ollama_model,
			prompt_version,
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
		&out.PromptVersion,
		&out.UpdatedAt,
		&out.UpdatedBy,
	)
	return out, err
}

func persistEvalRun(
	ctx context.Context,
	settings adminAIRuntimeSettings,
	runtime ollamaRuntimeStatus,
	runtimeMode string,
	userID uuid.UUID,
	fixtureCount int,
	passedCount int,
	successRate float64,
	qualityScore float64,
	cases []aiEvalCaseResult,
) (uuid.UUID, error) {
	var runID uuid.UUID
	if err := db.Pool.QueryRow(ctx, `
		INSERT INTO admin_ai_eval_runs (
			model_name, prompt_version, fixture_count, passed_count, success_rate, quality_score,
			runtime_mode, ai_enabled, ollama_reachable, model_available, created_by, started_at, finished_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, NOW(), NOW()
		)
		RETURNING id
	`, settings.OllamaModel, settings.PromptVersion, fixtureCount, passedCount, successRate, qualityScore, runtimeMode, runtime.AIEnabled, runtime.OllamaReachable, runtime.ModelAvailable, userID).Scan(&runID); err != nil {
		return uuid.Nil, err
	}
	for _, c := range cases {
		keywordsJSON, err := json.Marshal(c.ExpectedKeywords)
		if err != nil {
			return uuid.Nil, err
		}
		if _, err := db.Pool.Exec(ctx, `
			INSERT INTO admin_ai_eval_cases (
				run_id, fixture_id, input_excerpt, expected_keywords, summary_excerpt, keyword_coverage, passed
			) VALUES ($1, $2, $3, $4::jsonb, $5, $6, $7)
		`, runID, c.FixtureID, c.InputExcerpt, string(keywordsJSON), c.SummaryExcerpt, c.KeywordCoverage, c.Passed); err != nil {
			return uuid.Nil, err
		}
	}
	return runID, nil
}

func loadRecentEvalRuns(ctx context.Context, limit int) ([]adminAIEvalRunSummary, error) {
	if limit <= 0 {
		limit = 5
	}
	rows, err := db.Pool.Query(ctx, `
		SELECT id::text, model_name, prompt_version, fixture_count, passed_count, success_rate, quality_score,
		       runtime_mode, ollama_reachable, model_available, started_at::text, finished_at::text
		FROM admin_ai_eval_runs
		ORDER BY started_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []adminAIEvalRunSummary{}
	for rows.Next() {
		var r adminAIEvalRunSummary
		if err := rows.Scan(&r.ID, &r.ModelName, &r.PromptVersion, &r.FixtureCount, &r.PassedCount, &r.SuccessRate, &r.QualityScore, &r.RuntimeMode, &r.OllamaReachable, &r.ModelAvailable, &r.StartedAt, &r.FinishedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
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

func loadSLOOverview(c *gin.Context) adminSLOOverview {
	out := adminSLOOverview{
		Notifications: adminSLOMetric{TargetPercent: 99.0},
		AI:            adminSLOMetric{TargetPercent: 97.0},
		API:           adminSLOMetric{TargetPercent: 99.5},
		Alerts:        []gin.H{},
	}

	_ = db.Pool.QueryRow(c.Request.Context(), `
		SELECT
			COUNT(*)::int AS total,
			COUNT(*) FILTER (WHERE success = FALSE)::int AS failed,
			COALESCE(AVG(latency_ms)::float8, 0)::float8 AS avg_latency,
			COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY latency_ms), 0)::float8 AS p95_latency
		FROM notification_delivery_attempts
		WHERE created_at >= NOW() - INTERVAL '24 hours'
	`).Scan(&out.Notifications.Requests24h, &out.Notifications.Errors24h, &out.Notifications.AvgLatencyMS, &out.Notifications.P95LatencyMS)
	out.Notifications.SuccessPercent = successPercent(out.Notifications.Requests24h, out.Notifications.Errors24h)
	out.Notifications.ErrorBudgetUsed = errorBudgetUsed(out.Notifications.SuccessPercent, out.Notifications.TargetPercent)

	_ = db.Pool.QueryRow(c.Request.Context(), `
		SELECT
			COUNT(*)::int AS total,
			COUNT(*) FILTER (WHERE status = 'failed')::int AS failed,
			COALESCE(AVG(EXTRACT(EPOCH FROM (finished_at - started_at)) * 1000) FILTER (
				WHERE started_at IS NOT NULL AND finished_at IS NOT NULL
			), 0)::float8 AS avg_latency,
			COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (
				ORDER BY EXTRACT(EPOCH FROM (finished_at - started_at)) * 1000
			) FILTER (WHERE started_at IS NOT NULL AND finished_at IS NOT NULL), 0)::float8 AS p95_latency
		FROM document_summary_jobs
		WHERE created_at >= NOW() - INTERVAL '24 hours'
	`).Scan(&out.AI.Requests24h, &out.AI.Errors24h, &out.AI.AvgLatencyMS, &out.AI.P95LatencyMS)
	out.AI.SuccessPercent = successPercent(out.AI.Requests24h, out.AI.Errors24h)
	out.AI.ErrorBudgetUsed = errorBudgetUsed(out.AI.SuccessPercent, out.AI.TargetPercent)

	_ = db.Pool.QueryRow(c.Request.Context(), `
		SELECT
			COUNT(*)::int AS total,
			COUNT(*) FILTER (WHERE status_code >= 500)::int AS failed,
			COALESCE(AVG(latency_ms)::float8, 0)::float8 AS avg_latency,
			COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY latency_ms), 0)::float8 AS p95_latency
		FROM api_request_telemetry
		WHERE created_at >= NOW() - INTERVAL '24 hours'
	`).Scan(&out.API.Requests24h, &out.API.Errors24h, &out.API.AvgLatencyMS, &out.API.P95LatencyMS)
	out.API.SuccessPercent = successPercent(out.API.Requests24h, out.API.Errors24h)
	out.API.ErrorBudgetUsed = errorBudgetUsed(out.API.SuccessPercent, out.API.TargetPercent)

	out.Alerts = buildSLOAlerts(out)
	return out
}

func successPercent(total, errors int) float64 {
	if total <= 0 {
		return 100
	}
	ok := float64(total - errors)
	if ok < 0 {
		ok = 0
	}
	return (ok / float64(total)) * 100
}

func errorBudgetUsed(successPercentValue, targetPercent float64) float64 {
	if targetPercent >= 100 {
		return 0
	}
	allowed := 100 - targetPercent
	if allowed <= 0 {
		return 0
	}
	actualFailure := 100 - successPercentValue
	if actualFailure <= 0 {
		return 0
	}
	return (actualFailure / allowed) * 100
}

func buildSLOAlerts(overview adminSLOOverview) []gin.H {
	alerts := []gin.H{}
	add := func(domain string, m adminSLOMetric) {
		if m.ErrorBudgetUsed >= 100 {
			alerts = append(alerts, gin.H{
				"severity": "warning",
				"code":     domain + "_error_budget_exhausted",
				"message":  domain + " SLO error budget exhausted in last 24h.",
				"value":    m.ErrorBudgetUsed,
			})
		} else if m.ErrorBudgetUsed >= 50 {
			alerts = append(alerts, gin.H{
				"severity": "info",
				"code":     domain + "_error_budget_burn_high",
				"message":  domain + " SLO error budget burn is above 50% in last 24h.",
				"value":    m.ErrorBudgetUsed,
			})
		}
		if m.P95LatencyMS >= 2000 {
			alerts = append(alerts, gin.H{
				"severity": "warning",
				"code":     domain + "_p95_latency_high",
				"message":  domain + " p95 latency is above 2000ms.",
				"value":    m.P95LatencyMS,
			})
		}
	}
	add("notifications", overview.Notifications)
	add("ai", overview.AI)
	add("api", overview.API)
	return alerts
}
