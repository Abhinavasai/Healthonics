package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
)

type PreVisitHandler struct {
	ai *AIHealthHandler
}

func NewPreVisitHandler(ai *AIHealthHandler) *PreVisitHandler {
	return &PreVisitHandler{ai: ai}
}

// GetQuestionnaire  GET /api/appointments/:id/questionnaire
// Returns AI-generated questions (or cached ones if already generated).
func (h *PreVisitHandler) GetQuestionnaire(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}

	apptID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment id"})
		return
	}

	ctx := c.Request.Context()

	// Verify access: patient owns it, or doctor assigned.
	var patientID, doctorID string
	var reason, scheduledAt string
	if err := db.Pool.QueryRow(ctx, `
		SELECT patient_id::text, doctor_id::text, reason, scheduled_at::text
		FROM appointments WHERE id = $1
	`, apptID).Scan(&patientID, &doctorID, &reason, &scheduledAt); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}

	uid := claims.UserID.String()
	if claims.Role == "patient" && uid != patientID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}
	if claims.Role == "doctor" && uid != doctorID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Check for existing questionnaire in DB.
	var questionsJSON, answersJSON string
	var submittedAt *string
	err = db.Pool.QueryRow(ctx, `
		SELECT questions_json, COALESCE(answers_json, ''), submitted_at::text
		FROM appointment_questionnaires WHERE appointment_id = $1
	`, apptID).Scan(&questionsJSON, &answersJSON, &submittedAt)

	if err == nil {
		// Already exists.
		c.JSON(http.StatusOK, gin.H{
			"questions":    questionsJSON,
			"answers":      answersJSON,
			"submitted_at": submittedAt,
		})
		return
	}

	// Generate questions via AI.
	sysMsg := `You are a pre-visit medical intake assistant. Based on the appointment reason, generate 4-5 concise intake questions to help the doctor prepare for the visit.

Rules:
- Questions must be specific to the complaint.
- Use simple language the patient can easily answer.
- Include: duration, severity (1-10), associated symptoms, prior treatments tried, impact on daily life.
- Return ONLY a JSON array of strings, no markdown, no explanation. Example: ["How long have you had this symptom?","Rate your pain 1-10."]`

	userMsg := fmt.Sprintf("Appointment reason: %s\nScheduled: %s", reason, scheduledAt[:10])
	aiReply, _ := h.ai.callAI(c, sysMsg, userMsg)

	questions := aiReply
	if questions == "" || !strings.HasPrefix(strings.TrimSpace(questions), "[") {
		// Fallback: generic questions.
		questions = `["How long have you been experiencing this?","On a scale of 1-10, how severe are your symptoms?","Have you tried any treatments or medications for this?","Have you had similar symptoms before?","Is there anything that makes it better or worse?"]`
	}

	// Store in DB.
	if _, err := db.Pool.Exec(ctx, `
		INSERT INTO appointment_questionnaires (appointment_id, questions_json)
		VALUES ($1, $2)
		ON CONFLICT (appointment_id) DO NOTHING
	`, apptID, questions); err != nil {
		slog.Error("previsit: failed to store questionnaire", "error", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"questions":    questions,
		"answers":      "",
		"submitted_at": nil,
	})
}

// SubmitAnswers  POST /api/appointments/:id/questionnaire
func (h *PreVisitHandler) SubmitAnswers(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Patients only"})
		return
	}

	apptID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment id"})
		return
	}

	var req struct {
		Answers string `json:"answers"` // JSON array of strings
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Answers) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "answers required"})
		return
	}
	if len([]rune(req.Answers)) > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "answers too long"})
		return
	}

	ctx := c.Request.Context()

	// Verify patient owns the appointment.
	var patientID string
	if err := db.Pool.QueryRow(ctx, `SELECT patient_id::text FROM appointments WHERE id = $1`, apptID).Scan(&patientID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}
	if claims.UserID.String() != patientID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	now := time.Now()
	res, err := db.Pool.Exec(ctx, `
		UPDATE appointment_questionnaires
		SET answers_json = $1, submitted_at = $2
		WHERE appointment_id = $3
	`, req.Answers, now, apptID)
	if err != nil {
		slog.Error("previsit: failed to store answers", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if res.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Questionnaire not yet generated for this appointment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
