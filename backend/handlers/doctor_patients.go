package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/healthonyx/backend/db"
)

type DoctorPatientsHandler struct{}

func NewDoctorPatientsHandler() *DoctorPatientsHandler {
	return &DoctorPatientsHandler{}
}

type doctorPatientRow struct {
	PatientID          string  `json:"patient_id"`
	Email              string  `json:"email"`
	LastAppointmentAt  *string `json:"last_appointment_at"`
	ActivePrescriptions int    `json:"active_prescriptions"`
	AdherenceScore     *int    `json:"adherence_score"`
	AdherenceLabel     string  `json:"adherence_label"`
}

// List returns all patients that have ever had an appointment with this doctor.
func (h *DoctorPatientsHandler) List(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	ctx := c.Request.Context()
	rows, err := db.Pool.Query(ctx, `
		SELECT
			u.id::text,
			u.email,
			MAX(a.scheduled_at)::text AS last_appt,
			(SELECT COUNT(*) FROM prescriptions p WHERE p.patient_id = u.id AND p.doctor_id = $1 AND p.status = 'active')::int AS active_rx
		FROM appointments a
		JOIN users u ON u.id = a.patient_id
		WHERE a.doctor_id = $1
		GROUP BY u.id, u.email
		ORDER BY last_appt DESC NULLS LAST
		LIMIT 200
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	var out []doctorPatientRow
	for rows.Next() {
		var r doctorPatientRow
		if err := rows.Scan(&r.PatientID, &r.Email, &r.LastAppointmentAt, &r.ActivePrescriptions); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if out == nil {
		out = []doctorPatientRow{}
	}
	c.JSON(http.StatusOK, gin.H{"patients": out})
}
