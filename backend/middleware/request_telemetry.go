package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/healthonyx/backend/db"
)

func RequestTelemetry() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		if db.Pool == nil {
			return
		}
		path := strings.TrimSpace(c.FullPath())
		if path == "" {
			path = strings.TrimSpace(c.Request.URL.Path)
		}
		if !strings.HasPrefix(path, "/api/") {
			return
		}
		latencyMS := int(time.Since(start).Milliseconds())
		if latencyMS < 0 {
			latencyMS = 0
		}
		_, _ = db.Pool.Exec(c.Request.Context(), `
			INSERT INTO api_request_telemetry(method, path, status_code, latency_ms)
			VALUES($1, $2, $3, $4)
		`, c.Request.Method, path, c.Writer.Status(), latencyMS)
	}
}
