package handlers

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/healthonyx/backend/models"
	"github.com/jackc/pgx/v5"
)

type GeoBookingHandler struct{}

func NewGeoBookingHandler() *GeoBookingHandler {
	return &GeoBookingHandler{}
}

// haversineKm returns great-circle distance in kilometers.
func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const earthR = 6371.0
	const deg = math.Pi / 180
	dlat := (lat2 - lat1) * deg
	dlon := (lon2 - lon1) * deg
	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1*deg)*math.Cos(lat2*deg)*math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthR * c
}

type hospitalNearJSON struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	City       string  `json:"city"`
	Region     string  `json:"region"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	DistanceKm float64 `json:"distance_km"`
}

// ListHospitalsNear returns hospitals whose coordinates fall within radius_km of (lat,lng).
func (h *GeoBookingHandler) ListHospitalsNear(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}

	lat, lng, radiusKm, ok := parseLatLngRadius(c)
	if !ok {
		return
	}

	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id, name, city, region, latitude, longitude FROM hospitals
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	out := make([]hospitalNearJSON, 0)
	for rows.Next() {
		var id uuid.UUID
		var name, city, region string
		var hLat, hLng float64
		if err := rows.Scan(&id, &name, &city, &region, &hLat, &hLng); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		d := haversineKm(lat, lng, hLat, hLng)
		if d <= radiusKm {
			out = append(out, hospitalNearJSON{
				ID:         id.String(),
				Name:       name,
				City:       city,
				Region:     region,
				Latitude:   hLat,
				Longitude:  hLng,
				DistanceKm: math.Round(d*100) / 100,
			})
		}
	}
	if rows.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"hospitals": out})
}

func parseLatLngRadius(c *gin.Context) (lat, lng, radiusKm float64, ok bool) {
	latStr := strings.TrimSpace(c.Query("lat"))
	lngStr := strings.TrimSpace(c.Query("lng"))
	rStr := strings.TrimSpace(c.Query("radius_km"))
	if latStr == "" || lngStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat and lng query parameters are required"})
		return 0, 0, 0, false
	}
	var err error
	lat, err = strconv.ParseFloat(latStr, 64)
	if err != nil || lat < -90 || lat > 90 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat must be a number between -90 and 90"})
		return 0, 0, 0, false
	}
	lng, err = strconv.ParseFloat(lngStr, 64)
	if err != nil || lng < -180 || lng > 180 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lng must be a number between -180 and 180"})
		return 0, 0, 0, false
	}
	radiusKm = 25
	if rStr != "" {
		radiusKm, err = strconv.ParseFloat(rStr, 64)
		if err != nil || radiusKm <= 0 || radiusKm > 500 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "radius_km must be a positive number up to 500"})
			return 0, 0, 0, false
		}
	}
	return lat, lng, radiusKm, true
}

type doctorSearchJSON struct {
	ID             string  `json:"id"`
	Email          string  `json:"email"`
	Specialization string  `json:"specialization"`
	HospitalID     *string `json:"hospital_id,omitempty"`
	DistanceKm     float64 `json:"distance_km"`
}

// SearchDoctors returns doctors in radius; optional specialization filters (substring match).
func (h *GeoBookingHandler) SearchDoctors(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}

	lat, lng, radiusKm, ok := parseLatLngRadius(c)
	if !ok {
		return
	}
	spec := strings.TrimSpace(c.Query("specialization"))
	var filterHospitalID *uuid.UUID
	if hs := strings.TrimSpace(c.Query("hospital_id")); hs != "" {
		hid, err := uuid.Parse(hs)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "hospital_id must be a valid UUID"})
			return
		}
		filterHospitalID = &hid
	}

	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT u.id, u.email, u.specialization, u.hospital_id,
			h.latitude, h.longitude, u.practice_latitude, u.practice_longitude
		FROM users u
		LEFT JOIN hospitals h ON h.id = u.hospital_id
		WHERE u.role = 'doctor'
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	out := make([]doctorSearchJSON, 0)
	for rows.Next() {
		var id uuid.UUID
		var email, specialization string
		var hid *uuid.UUID
		var hLat, hLng *float64
		var pLat, pLng *float64
		if err := rows.Scan(&id, &email, &specialization, &hid, &hLat, &hLng, &pLat, &pLng); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		if spec != "" && !strings.Contains(strings.ToLower(specialization), strings.ToLower(spec)) {
			continue
		}
		if filterHospitalID != nil {
			if hid == nil || *hid != *filterHospitalID {
				continue
			}
		}

		var effLat, effLng float64
		var has bool
		if hLat != nil && hLng != nil {
			effLat, effLng, has = *hLat, *hLng, true
		} else if pLat != nil && pLng != nil {
			effLat, effLng, has = *pLat, *pLng, true
		}
		if !has {
			continue
		}

		d := haversineKm(lat, lng, effLat, effLng)
		if d > radiusKm {
			continue
		}

		var hidStr *string
		if hid != nil {
			s := hid.String()
			hidStr = &s
		}
		out = append(out, doctorSearchJSON{
			ID:             id.String(),
			Email:          email,
			Specialization: specialization,
			HospitalID:     hidStr,
			DistanceKm:     math.Round(d*100) / 100,
		})
	}
	if rows.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"doctors": out})
}

type slotJSON struct {
	ID        string    `json:"id"`
	DoctorID  string    `json:"doctor_id"`
	StartAt   time.Time `json:"start_at"`
	EndAt     time.Time `json:"end_at"`
	Available bool      `json:"available"`
}

// ListOpenSlotsForDoctor returns slots for a doctor that are not yet booked.
func (h *GeoBookingHandler) ListOpenSlotsForDoctor(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}

	doctorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid doctor id"})
		return
	}

	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id, doctor_id, start_at, end_at, patient_id
		FROM doctor_slots
		WHERE doctor_id = $1 AND start_at > NOW()
		ORDER BY start_at ASC
	`, doctorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	slots := make([]slotJSON, 0)
	for rows.Next() {
		var id, docID uuid.UUID
		var startAt, endAt time.Time
		var patientID *uuid.UUID
		if err := rows.Scan(&id, &docID, &startAt, &endAt, &patientID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		slots = append(slots, slotJSON{
			ID:        id.String(),
			DoctorID:  docID.String(),
			StartAt:   startAt,
			EndAt:     endAt,
			Available: patientID == nil,
		})
	}
	if rows.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"slots": slots})
}

type createSlotRequest struct {
	StartAt time.Time `json:"start_at" binding:"required"`
	EndAt   time.Time `json:"end_at" binding:"required"`
}

// CreateSlot lets a doctor add an availability window.
func (h *GeoBookingHandler) CreateSlot(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	var req createSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if !req.EndAt.After(req.StartAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_at must be after start_at"})
		return
	}

	var id uuid.UUID
	err := db.Pool.QueryRow(c.Request.Context(), `
		INSERT INTO doctor_slots (doctor_id, start_at, end_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`, claims.UserID, req.StartAt, req.EndAt).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusConflict, gin.H{"error": "A slot with this start time already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id.String()})
}

// DeleteOpenSlot removes a slot the doctor owns if it is still open.
func (h *GeoBookingHandler) DeleteOpenSlot(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slot id"})
		return
	}

	cmd, err := db.Pool.Exec(c.Request.Context(), `
		DELETE FROM doctor_slots
		WHERE id = $1 AND doctor_id = $2 AND patient_id IS NULL
	`, id, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if cmd.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Open slot not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type bookSlotRequest struct {
	SlotID string `json:"slot_id" binding:"required"`
	Reason string `json:"reason" binding:"required"`
}

// BookSlot creates an appointment from an open slot and locks the slot.
func (h *GeoBookingHandler) BookSlot(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	var req bookSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	slotID, err := uuid.Parse(strings.TrimSpace(req.SlotID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slot_id must be a valid UUID"})
		return
	}
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reason is required"})
		return
	}

	tx, err := db.Pool.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer tx.Rollback(c.Request.Context())

	var doctorID uuid.UUID
	var startAt time.Time
	var bookedBy *uuid.UUID
	err = tx.QueryRow(c.Request.Context(), `
		SELECT doctor_id, start_at, patient_id FROM doctor_slots
		WHERE id = $1
		FOR UPDATE
	`, slotID).Scan(&doctorID, &startAt, &bookedBy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Slot not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if bookedBy != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Slot is no longer available"})
		return
	}

	if doctorID == claims.UserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slot"})
		return
	}

	var appt models.Appointment
	err = scanAppointment(tx.QueryRow(c.Request.Context(), `
		INSERT INTO appointments (patient_id, doctor_id, scheduled_at, reason, status, slot_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, patient_id, doctor_id, scheduled_at, reason, status, pending_scheduled_at, created_at, updated_at
	`, claims.UserID, doctorID, startAt, reason, models.AppointmentStatusPending, slotID), &appt)
	if err != nil {
		if strings.Contains(err.Error(), "unique") && strings.Contains(err.Error(), "slot_id") {
			c.JSON(http.StatusConflict, gin.H{"error": "This slot is already booked"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	_, err = tx.Exec(c.Request.Context(), `
		UPDATE doctor_slots SET patient_id = $1 WHERE id = $2 AND patient_id IS NULL
	`, claims.UserID, slotID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusCreated, appt)
}

// ListHospitalDepartments returns curated departments for a hospital (for find-care funnel).
func (h *GeoBookingHandler) ListHospitalDepartments(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hospital id"})
		return
	}
	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT department_name FROM hospital_departments
		WHERE hospital_id = $1
		ORDER BY department_name ASC
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()
	names := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		names = append(names, name)
	}
	if rows.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"departments": names})
}
