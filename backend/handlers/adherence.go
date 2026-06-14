package handlers

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
)

// AdherenceHandler computes a rolling prescription adherence score for a patient.
// Score 0-100 derived purely from existing prescriptions data — no new data collection needed.
type AdherenceHandler struct{}

func NewAdherenceHandler() *AdherenceHandler { return &AdherenceHandler{} }

type adherenceResult struct {
	PatientID    string  `json:"patient_id"`
	Score        int     `json:"score"`
	Label        string  `json:"label"`
	Total        int     `json:"total"`
	Active       int     `json:"active"`
	Revoked      int     `json:"revoked"`
	Completed    int     `json:"completed"`
	AvgDaysUsed  float64 `json:"avg_days_used"`
}

// GET /api/patients/:patientId/adherence-score
func (h *AdherenceHandler) Score(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" && claims.Role != "admin" && claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	patientIDStr := c.Param("patientId")
	patientID, err := uuid.Parse(patientIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient id"})
		return
	}

	// Patients can only see their own score.
	if claims.Role == "patient" && claims.UserID != patientID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	ctx := c.Request.Context()

	var total, active, revoked int
	var avgDays float64

	// Count by status.
	rows, err := db.Pool.Query(ctx, `
		SELECT status, COUNT(*) as cnt,
		       AVG(EXTRACT(EPOCH FROM (NOW() - created_at)) / 86400)
		FROM prescriptions
		WHERE patient_id = $1
		GROUP BY status
	`, patientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	completedCount := 0
	totalDays := 0.0
	totalPrescriptions := 0

	for rows.Next() {
		var status string
		var cnt int
		var avgD *float64 // AVG() can return NULL when group is empty
		if err := rows.Scan(&status, &cnt, &avgD); err != nil {
			continue
		}
		if avgD == nil {
			d := 0.0
			avgD = &d
		}
		total += cnt
		totalPrescriptions += cnt
		totalDays += *avgD * float64(cnt)
		switch status {
		case "active":
			active += cnt
		case "revoked":
			revoked += cnt
		case "completed":
			completedCount += cnt
		}
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	if totalPrescriptions > 0 {
		avgDays = totalDays / float64(totalPrescriptions)
	}

	if total == 0 {
		c.JSON(http.StatusOK, gin.H{
			"patient_id":    patientID.String(),
			"score":         nil,
			"label":         "No data",
			"total":         0,
			"active":        0,
			"revoked":       0,
			"completed":     0,
			"avg_days_used": 0,
		})
		return
	}

	// Score formula:
	// Base: (active + completed) / total * 80, penalty for revoked, bonus for duration.
	score := 0
	if total > 0 {
		base := float64(active+completedCount) / float64(total) * 80.0
		revokedPenalty := math.Min(float64(revoked)/float64(total)*40.0, 20.0)
		durationBonus := math.Min(avgDays/30.0*10.0, 10.0)
		raw := base - revokedPenalty + durationBonus
		score = int(math.Round(math.Max(0, math.Min(100, raw))))
	}

	label := adherenceLabel(score)

	c.JSON(http.StatusOK, adherenceResult{
		PatientID:   patientID.String(),
		Score:       score,
		Label:       label,
		Total:       total,
		Active:      active,
		Revoked:     revoked,
		Completed:   completedCount,
		AvgDaysUsed: math.Round(avgDays*10) / 10,
	})
}

func adherenceLabel(score int) string {
	switch {
	case score >= 80:
		return "Good"
	case score >= 60:
		return "Fair"
	case score >= 40:
		return "Poor"
	default:
		return "Very Poor"
	}
}
