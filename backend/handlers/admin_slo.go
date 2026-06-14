package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/healthonyx/backend/db"
)

type AdminSLOHandler struct{}

func NewAdminSLOHandler() *AdminSLOHandler {
	return &AdminSLOHandler{}
}

// Dashboard returns SLO metrics: p50/p95/p99 latency, error rate, active users.
func (h *AdminSLOHandler) Dashboard(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}
	ctx := c.Request.Context()

	// Latency percentiles (last 24h).
	var p50, p95, p99 *float64
	var totalRequests, errorRequests int64
	if err := db.Pool.QueryRow(ctx, `
		SELECT
			PERCENTILE_CONT(0.50) WITHIN GROUP (ORDER BY latency_ms),
			PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY latency_ms),
			PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY latency_ms),
			COUNT(*),
			COALESCE(SUM(CASE WHEN status_code >= 500 THEN 1 ELSE 0 END), 0)
		FROM api_request_telemetry
		WHERE created_at > NOW() - INTERVAL '24 hours'
	`).Scan(&p50, &p95, &p99, &totalRequests, &errorRequests); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	// Error rate.
	var errorRate float64
	if totalRequests > 0 {
		errorRate = float64(errorRequests) / float64(totalRequests) * 100
	}

	// Active users: distinct users in last 24h (via audit log).
	var activeUsers int64
	if err := db.Pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT actor_user_id) FROM audit_logs
		WHERE created_at > NOW() - INTERVAL '24 hours' AND actor_user_id IS NOT NULL
	`).Scan(&activeUsers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	// Recent errors (last 50).
	rows, err := db.Pool.Query(ctx, `
		SELECT method, path, status_code, latency_ms, created_at::text
		FROM api_request_telemetry
		WHERE status_code >= 400
		  AND created_at > NOW() - INTERVAL '24 hours'
		ORDER BY created_at DESC
		LIMIT 50
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	type errorEntry struct {
		Method     string `json:"method"`
		Path       string `json:"path"`
		StatusCode int    `json:"status_code"`
		LatencyMS  int    `json:"latency_ms"`
		CreatedAt  string `json:"created_at"`
	}
	var recentErrors []errorEntry
	for rows.Next() {
		var e errorEntry
		if err := rows.Scan(&e.Method, &e.Path, &e.StatusCode, &e.LatencyMS, &e.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		recentErrors = append(recentErrors, e)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if recentErrors == nil {
		recentErrors = []errorEntry{}
	}

	// Requests per hour (last 24 buckets).
	rows2, err := db.Pool.Query(ctx, `
		SELECT
			DATE_TRUNC('hour', created_at)::text AS hour,
			COUNT(*) AS requests
		FROM api_request_telemetry
		WHERE created_at > NOW() - INTERVAL '24 hours'
		GROUP BY 1
		ORDER BY 1
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows2.Close()

	type hourBucket struct {
		Hour     string `json:"hour"`
		Requests int64  `json:"requests"`
	}
	var hourly []hourBucket
	for rows2.Next() {
		var hb hourBucket
		if err := rows2.Scan(&hb.Hour, &hb.Requests); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		hourly = append(hourly, hb)
	}
	if err := rows2.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if hourly == nil {
		hourly = []hourBucket{}
	}

	c.JSON(http.StatusOK, gin.H{
		"latency_p50_ms":    p50,
		"latency_p95_ms":    p95,
		"latency_p99_ms":    p99,
		"total_requests_24h": totalRequests,
		"error_requests_24h": errorRequests,
		"error_rate_pct":    errorRate,
		"active_users_24h":  activeUsers,
		"recent_errors":     recentErrors,
		"hourly_requests":   hourly,
	})
}
