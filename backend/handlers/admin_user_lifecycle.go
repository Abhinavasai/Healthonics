package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/healthonyx/backend/db"
)

type AdminUserLifecycleHandler struct{}

type updateLifecycleSettingsRequest struct {
	NewUserWindowDays  int `json:"new_user_window_days"`
	InactiveWindowDays int `json:"inactive_window_days"`
}

func NewAdminUserLifecycleHandler() *AdminUserLifecycleHandler {
	return &AdminUserLifecycleHandler{}
}

func (h *AdminUserLifecycleHandler) GetSettingsKpis(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}

	type settingsRow struct {
		NewUserWindowDays  int    `json:"new_user_window_days"`
		InactiveWindowDays int    `json:"inactive_window_days"`
		UpdatedAt          string `json:"updated_at"`
		UpdatedBy          string `json:"updated_by"`
	}
	type kpisRow struct {
		TotalUsers         int `json:"total_users"`
		TotalPatients      int `json:"total_patients"`
		TotalDoctors       int `json:"total_doctors"`
		TotalAdmins        int `json:"total_admins"`
		NewUsersInWindow   int `json:"new_users_in_window"`
		InactiveUsersCount int `json:"inactive_users_count"`
	}

	var settings settingsRow
	if err := db.Pool.QueryRow(c.Request.Context(), `
		SELECT
			new_user_window_days,
			inactive_window_days,
			updated_at::text,
			COALESCE(updated_by::text, '')
		FROM admin_user_lifecycle_settings
		WHERE id = TRUE
	`).Scan(&settings.NewUserWindowDays, &settings.InactiveWindowDays, &settings.UpdatedAt, &settings.UpdatedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	var kpis kpisRow
	if err := db.Pool.QueryRow(c.Request.Context(), `
		WITH cfg AS (
			SELECT new_user_window_days, inactive_window_days
			FROM admin_user_lifecycle_settings
			WHERE id = TRUE
		),
		last_activity AS (
			SELECT user_id, MAX(event_at) AS last_event_at
			FROM (
				SELECT id AS user_id, created_at AS event_at FROM users
				UNION ALL
				SELECT patient_id AS user_id, created_at AS event_at FROM appointments
				UNION ALL
				SELECT doctor_id AS user_id, created_at AS event_at FROM appointments
				UNION ALL
				SELECT sender_id AS user_id, created_at AS event_at FROM messages
			) events
			GROUP BY user_id
		)
		SELECT
			COUNT(*)::int AS total_users,
			COUNT(*) FILTER (WHERE u.role = 'patient')::int AS total_patients,
			COUNT(*) FILTER (WHERE u.role = 'doctor')::int AS total_doctors,
			COUNT(*) FILTER (WHERE u.role = 'admin')::int AS total_admins,
			COUNT(*) FILTER (
				WHERE u.created_at >= NOW() - make_interval(days => cfg.new_user_window_days)
			)::int AS new_users_in_window,
			COUNT(*) FILTER (
				WHERE COALESCE(la.last_event_at, u.created_at) < NOW() - make_interval(days => cfg.inactive_window_days)
			)::int AS inactive_users_count
		FROM users u
		CROSS JOIN cfg
		LEFT JOIN last_activity la ON la.user_id = u.id
	`).Scan(
		&kpis.TotalUsers,
		&kpis.TotalPatients,
		&kpis.TotalDoctors,
		&kpis.TotalAdmins,
		&kpis.NewUsersInWindow,
		&kpis.InactiveUsersCount,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settings": settings,
		"kpis":     kpis,
	})
}

func (h *AdminUserLifecycleHandler) UpdateSettings(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}

	var req updateLifecycleSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if req.NewUserWindowDays < 1 || req.NewUserWindowDays > 3650 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new_user_window_days must be between 1 and 3650"})
		return
	}
	if req.InactiveWindowDays < 1 || req.InactiveWindowDays > 3650 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "inactive_window_days must be between 1 and 3650"})
		return
	}

	if _, err := db.Pool.Exec(c.Request.Context(), `
		UPDATE admin_user_lifecycle_settings
		SET
			new_user_window_days = $1,
			inactive_window_days = $2,
			updated_by = $3,
			updated_at = NOW()
		WHERE id = TRUE
	`, req.NewUserWindowDays, req.InactiveWindowDays, claims.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	h.GetSettingsKpis(c)
}
