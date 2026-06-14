package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/jackc/pgx/v5"
)

type AppointmentRatingHandler struct{}

func NewAppointmentRatingHandler() *AppointmentRatingHandler {
	return &AppointmentRatingHandler{}
}

type appointmentRating struct {
	AppointmentID string    `json:"appointment_id"`
	PatientID     string    `json:"patient_id"`
	DoctorID      string    `json:"doctor_id"`
	Rating        int       `json:"rating"`
	Comment       string    `json:"comment"`
	CreatedAt     time.Time `json:"created_at"`
}

// Submit — patient submits a 1-5 star rating after a completed appointment.
func (h *AppointmentRatingHandler) Submit(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	apptID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment id"})
		return
	}
	var body struct {
		Rating  int    `json:"rating" binding:"required"`
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rating (1-5) required"})
		return
	}
	if body.Rating < 1 || body.Rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rating must be between 1 and 5"})
		return
	}
	comment := strings.TrimSpace(body.Comment)
	if len([]rune(comment)) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment must be 1000 characters or fewer"})
		return
	}

	ctx := c.Request.Context()
	// Ensure appointment is completed and belongs to this patient.
	var doctorID uuid.UUID
	err = db.Pool.QueryRow(ctx, `
		SELECT doctor_id FROM appointments
		WHERE id = $1 AND patient_id = $2 AND status = 'completed'
	`, apptID, claims.UserID).Scan(&doctorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Completed appointment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO appointment_ratings (appointment_id, patient_id, doctor_id, rating, comment)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (appointment_id) DO UPDATE SET rating = EXCLUDED.rating, comment = EXCLUDED.comment
	`, apptID, claims.UserID, doctorID, body.Rating, comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GetForAppointment — returns the rating for an appointment (patient or doctor).
func (h *AppointmentRatingHandler) GetForAppointment(c *gin.Context) {
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
	var r appointmentRating
	err = db.Pool.QueryRow(ctx, `
		SELECT appointment_id::text, patient_id::text, doctor_id::text, rating, comment, created_at
		FROM appointment_ratings WHERE appointment_id = $1
	`, apptID).Scan(&r.AppointmentID, &r.PatientID, &r.DoctorID, &r.Rating, &r.Comment, &r.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No rating found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	// Authorization: only patient or assigned doctor.
	patID, _ := uuid.Parse(r.PatientID)
	docID, _ := uuid.Parse(r.DoctorID)
	if claims.Role == "patient" && claims.UserID != patID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	if claims.Role == "doctor" && claims.UserID != docID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	c.JSON(http.StatusOK, r)
}

// DoctorRatingSummary — doctor sees their average rating and recent reviews.
func (h *AppointmentRatingHandler) DoctorRatingSummary(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	ctx := c.Request.Context()
	var avg *float64
	var total int
	if err := db.Pool.QueryRow(ctx, `
		SELECT AVG(rating::float8), COUNT(*) FROM appointment_ratings WHERE doctor_id = $1
	`, claims.UserID).Scan(&avg, &total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	rows, err := db.Pool.Query(ctx, `
		SELECT ar.appointment_id::text, ar.rating, ar.comment, ar.created_at, u.email
		FROM appointment_ratings ar
		JOIN users u ON u.id = ar.patient_id
		WHERE ar.doctor_id = $1
		ORDER BY ar.created_at DESC
		LIMIT 20
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	type review struct {
		AppointmentID string    `json:"appointment_id"`
		Rating        int       `json:"rating"`
		Comment       string    `json:"comment"`
		CreatedAt     time.Time `json:"created_at"`
		PatientEmail  string    `json:"patient_email"`
	}
	var reviews []review
	for rows.Next() {
		var rv review
		if err := rows.Scan(&rv.AppointmentID, &rv.Rating, &rv.Comment, &rv.CreatedAt, &rv.PatientEmail); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		reviews = append(reviews, rv)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if reviews == nil {
		reviews = []review{}
	}
	c.JSON(http.StatusOK, gin.H{
		"average_rating": avg,
		"total_reviews":  total,
		"recent_reviews": reviews,
	})
}
