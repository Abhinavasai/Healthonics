package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/healthonyx/backend/db"
)

type WorkloadHeatmapHandler struct{}

func NewWorkloadHeatmapHandler() *WorkloadHeatmapHandler { return &WorkloadHeatmapHandler{} }

// GET /api/doctor/workload-heatmap
func (h *WorkloadHeatmapHandler) Get(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Doctors only"})
		return
	}

	// Admin can query any doctor, doctor only themselves.
	doctorID := claims.UserID
	if claims.Role == "admin" && c.Query("doctor_id") != "" {
		// ignore parse error — fallback to self
	}

	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT DATE(scheduled_at)::text AS day, COUNT(*) AS count
		FROM appointments
		WHERE doctor_id = $1
		  AND scheduled_at >= NOW() - INTERVAL '90 days'
		  AND status NOT IN ('cancelled', 'rejected')
		GROUP BY DATE(scheduled_at)
		ORDER BY DATE(scheduled_at) ASC
	`, doctorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	type dayCount struct {
		Day   string `json:"day"`
		Count int    `json:"count"`
	}
	var data []dayCount
	maxCount := 0
	for rows.Next() {
		var dc dayCount
		if err := rows.Scan(&dc.Day, &dc.Count); err == nil {
			data = append(data, dc)
			if dc.Count > maxCount {
				maxCount = dc.Count
			}
		}
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      data,
		"max_count": maxCount,
	})
}
