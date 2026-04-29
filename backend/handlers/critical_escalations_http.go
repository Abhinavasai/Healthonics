package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/healthonyx/backend/db"
)

type CriticalEscalationsHandler struct{}

func NewCriticalEscalationsHandler() *CriticalEscalationsHandler {
	return &CriticalEscalationsHandler{}
}

func (h *CriticalEscalationsHandler) ListMine(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Doctor role required"})
		return
	}
	if err := refreshDoctorCriticalEscalations(c.Request.Context(), claims.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	rows, err := listDoctorCriticalEscalations(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"escalations": rows})
}

func (h *CriticalEscalationsHandler) Ack(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Doctor role required"})
		return
	}
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid escalation id"})
		return
	}
	now := time.Now().UTC()
	cmd, err := db.Pool.Exec(c.Request.Context(), `
		UPDATE critical_result_escalations
		SET status = 'acknowledged',
		    acknowledged_at = $2,
		    acknowledged_by = $3,
		    last_evaluated_at = $2
		WHERE id::text = $1
		  AND doctor_id = $3
		  AND status = 'open'
	`, id, now, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if cmd.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Open escalation not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "status": "acknowledged"})
}
