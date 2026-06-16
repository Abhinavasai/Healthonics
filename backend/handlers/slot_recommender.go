package handlers

import (
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
)

// SlotRecommender scores available doctor slots and returns the top 3.
type SlotRecommenderHandler struct{}

func NewSlotRecommenderHandler() *SlotRecommenderHandler { return &SlotRecommenderHandler{} }

type slotScore struct {
	SlotID    string  `json:"slot_id"`
	DoctorID  string  `json:"doctor_id"`
	StartTime string  `json:"start_time"`
	EndTime   string  `json:"end_time"`
	Score     float64 `json:"score"`
	Reason    string  `json:"reason"`
}

// Recommend  GET /api/appointments/recommend?doctor_id=<uuid>
func (h *SlotRecommenderHandler) Recommend(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Patients only"})
		return
	}

	doctorIDStr := c.Query("doctor_id")
	if doctorIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "doctor_id query param required"})
		return
	}
	doctorID, err := uuid.Parse(doctorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid doctor_id"})
		return
	}

	ctx := c.Request.Context()

	// Fetch available slots for the next 30 days.
	rows, err := db.Pool.Query(ctx, `
		SELECT id::text, start_time, end_time
		FROM doctor_slots
		WHERE doctor_id = $1
		  AND start_time > NOW()
		  AND start_time < NOW() + INTERVAL '30 days'
		  AND id NOT IN (
		      SELECT slot_id FROM appointments
		      WHERE slot_id IS NOT NULL
		      AND status NOT IN ('cancelled', 'rejected')
		  )
		ORDER BY start_time ASC
		LIMIT 50
	`, doctorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	type rawSlot struct {
		ID        string
		StartTime time.Time
		EndTime   time.Time
	}
	var slots []rawSlot
	for rows.Next() {
		var s rawSlot
		if err := rows.Scan(&s.ID, &s.StartTime, &s.EndTime); err == nil {
			slots = append(slots, s)
		}
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	if len(slots) == 0 {
		c.JSON(http.StatusOK, gin.H{"recommendations": []slotScore{}, "message": "No available slots in the next 30 days."})
		return
	}

	// Patient's preferred hours from booking history.
	preferredHours := map[int]int{}
	pRows, err := db.Pool.Query(ctx, `
		SELECT EXTRACT(HOUR FROM scheduled_at)::int
		FROM appointments
		WHERE patient_id = $1 AND status IN ('approved', 'completed')
		ORDER BY created_at DESC LIMIT 20
	`, claims.UserID)
	if err == nil {
		defer pRows.Close()
		for pRows.Next() {
			var h int
			if err := pRows.Scan(&h); err == nil {
				preferredHours[h]++
			}
		}
	}

	// Doctor's no-show rate by hour (lower = better).
	noShowRate := map[int]float64{}
	nRows, err := db.Pool.Query(ctx, `
		SELECT EXTRACT(HOUR FROM scheduled_at)::int AS hr,
		       COUNT(*) FILTER (WHERE status = 'cancelled') * 1.0 / NULLIF(COUNT(*), 0) AS rate
		FROM appointments
		WHERE doctor_id = $1
		GROUP BY hr
	`, doctorID)
	if err == nil {
		defer nRows.Close()
		for nRows.Next() {
			var hr int
			var rate float64
			if err := nRows.Scan(&hr, &rate); err == nil {
				noShowRate[hr] = rate
			}
		}
	}

	maxPref := 1
	for _, v := range preferredHours {
		if v > maxPref {
			maxPref = v
		}
	}

	scored := make([]slotScore, 0, len(slots))
	for _, s := range slots {
		hr := s.StartTime.Hour()
		prefScore := float64(preferredHours[hr]) / float64(maxPref) // 0-1
		noshowPenalty := noShowRate[hr]                             // 0-1
		score := prefScore*0.6 + (1-noshowPenalty)*0.4

		reason := buildSlotReason(prefScore, noshowPenalty, hr)
		scored = append(scored, slotScore{
			SlotID:    s.ID,
			DoctorID:  doctorID.String(),
			StartTime: s.StartTime.Format(time.RFC3339),
			EndTime:   s.EndTime.Format(time.RFC3339),
			Score:     score,
			Reason:    reason,
		})
	}

	sort.Slice(scored, func(i, j int) bool { return scored[i].Score > scored[j].Score })
	top := scored
	if len(top) > 3 {
		top = top[:3]
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": top,
		"total_available": len(slots),
	})
}

func buildSlotReason(prefScore, noshowRate float64, hour int) string {
	timeLabel := "morning"
	switch {
	case hour >= 12 && hour < 17:
		timeLabel = "afternoon"
	case hour >= 17:
		timeLabel = "evening"
	}

	switch {
	case prefScore > 0.6 && noshowRate < 0.1:
		return "Matches your preferred " + timeLabel + " booking time with high doctor availability"
	case prefScore > 0.4:
		return "Aligns with your past " + timeLabel + " appointment preferences"
	case noshowRate < 0.05:
		return "High doctor reliability at this " + timeLabel + " time slot"
	default:
		return "Available " + timeLabel + " slot with reasonable scheduling history"
	}
}
