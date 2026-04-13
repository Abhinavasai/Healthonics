package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/healthonyx/backend/db"
)

type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

func (h *DashboardHandler) PatientSummary(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	var upcoming, pending int
	_ = db.Pool.QueryRow(c.Request.Context(), `
		SELECT
			COUNT(*) FILTER (WHERE status IN ('pending','approved') AND scheduled_at >= NOW()),
			COUNT(*) FILTER (WHERE status = 'pending')
		FROM appointments WHERE patient_id = $1
	`, claims.UserID).Scan(&upcoming, &pending)

	c.JSON(http.StatusOK, gin.H{
		"role":                     "patient",
		"upcoming_or_active_count": upcoming,
		"pending_requests":         pending,
	})
}

func (h *DashboardHandler) DoctorSummary(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	var today, pending int
	_ = db.Pool.QueryRow(c.Request.Context(), `
		SELECT
			COUNT(*) FILTER (WHERE scheduled_at >= CURRENT_DATE AND scheduled_at < CURRENT_DATE + INTERVAL '1 day' AND status IN ('pending','approved')),
			COUNT(*) FILTER (WHERE status = 'pending')
		FROM appointments WHERE doctor_id = $1
	`, claims.UserID).Scan(&today, &pending)

	c.JSON(http.StatusOK, gin.H{
		"role":            "doctor",
		"appointments_today": today,
		"pending_queue":      pending,
	})
}
