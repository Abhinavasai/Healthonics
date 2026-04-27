package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/healthonyx/backend/db"
)

type NotificationsHandler struct{}

func NewNotificationsHandler() *NotificationsHandler {
	return &NotificationsHandler{}
}

type NotificationPreference struct {
	Category     string `json:"category"`
	Enabled      bool   `json:"enabled"`
	EmailEnabled bool   `json:"email_enabled"`
	SmsEnabled   bool   `json:"sms_enabled"`
	InAppEnabled bool   `json:"in_app_enabled"`
}

type upsertNotificationPreferencesBody struct {
	Preferences []NotificationPreference `json:"preferences"`
}

var defaultNotificationPreferences = []NotificationPreference{
	{Category: "appointment_reminders", Enabled: true, EmailEnabled: true, SmsEnabled: false, InAppEnabled: true},
	{Category: "medication_reminders", Enabled: true, EmailEnabled: true, SmsEnabled: false, InAppEnabled: true},
	{Category: "lab_result_alerts", Enabled: true, EmailEnabled: true, SmsEnabled: false, InAppEnabled: true},
	{Category: "announcements", Enabled: true, EmailEnabled: true, SmsEnabled: false, InAppEnabled: true},
}

func validNotificationCategory(category string) bool {
	switch category {
	case "appointment_reminders", "medication_reminders", "lab_result_alerts", "announcements":
		return true
	default:
		return false
	}
}

func normalizeNotificationPreference(input NotificationPreference) (NotificationPreference, error) {
	category := strings.TrimSpace(strings.ToLower(input.Category))
	if !validNotificationCategory(category) {
		return NotificationPreference{}, errors.New("invalid category")
	}
	return NotificationPreference{
		Category:     category,
		Enabled:      input.Enabled,
		EmailEnabled: input.EmailEnabled,
		SmsEnabled:   input.SmsEnabled,
		InAppEnabled: input.InAppEnabled,
	}, nil
}

func (h *NotificationsHandler) ensurePreferencesSchema(c *gin.Context) error {
	_, err := db.Pool.Exec(c.Request.Context(), `
		CREATE TABLE IF NOT EXISTS notification_preferences (
			user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			category       TEXT NOT NULL,
			enabled        BOOLEAN NOT NULL DEFAULT TRUE,
			email_enabled  BOOLEAN NOT NULL DEFAULT TRUE,
			sms_enabled    BOOLEAN NOT NULL DEFAULT FALSE,
			in_app_enabled BOOLEAN NOT NULL DEFAULT TRUE,
			updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (user_id, category),
			CONSTRAINT notification_preferences_category_check CHECK (
				category IN ('appointment_reminders', 'medication_reminders', 'lab_result_alerts', 'announcements')
			)
		)
	`)
	return err
}

func (h *NotificationsHandler) seedDefaultsIfMissing(c *gin.Context, userID string) error {
	var count int
	if err := db.Pool.QueryRow(c.Request.Context(),
		`SELECT COUNT(*) FROM notification_preferences WHERE user_id = $1`, userID,
	).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	for _, pref := range defaultNotificationPreferences {
		_, err := db.Pool.Exec(c.Request.Context(), `
			INSERT INTO notification_preferences (user_id, category, enabled, email_enabled, sms_enabled, in_app_enabled)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (user_id, category) DO NOTHING
		`, userID, pref.Category, pref.Enabled, pref.EmailEnabled, pref.SmsEnabled, pref.InAppEnabled)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *NotificationsHandler) ListMine(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}

	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id::text, title, body, channel, status, scheduled_for::text, created_at::text
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 100
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	type row struct {
		ID           string  `json:"id"`
		Title        string  `json:"title"`
		Body         string  `json:"body"`
		Channel      string  `json:"channel"`
		Status       string  `json:"status"`
		ScheduledFor *string `json:"scheduled_for"`
		CreatedAt    string  `json:"created_at"`
	}
	var out []row
	for rows.Next() {
		var r row
		var sched sql.NullString
		if err := rows.Scan(&r.ID, &r.Title, &r.Body, &r.Channel, &r.Status, &sched, &r.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		if sched.Valid {
			s := sched.String
			r.ScheduledFor = &s
		}
		out = append(out, r)
	}
	if out == nil {
		out = []row{}
	}
	c.JSON(http.StatusOK, gin.H{"notifications": out})
}

// ListPreferences GET /api/notifications/preferences
func (h *NotificationsHandler) ListPreferences(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if err := h.ensurePreferencesSchema(c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if err := h.seedDefaultsIfMissing(c, claims.UserID.String()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT category, enabled, email_enabled, sms_enabled, in_app_enabled
		FROM notification_preferences
		WHERE user_id = $1
		ORDER BY category ASC
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	out := make([]NotificationPreference, 0, 8)
	for rows.Next() {
		var row NotificationPreference
		if err := rows.Scan(&row.Category, &row.Enabled, &row.EmailEnabled, &row.SmsEnabled, &row.InAppEnabled); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		out = append(out, row)
	}
	if out == nil {
		out = []NotificationPreference{}
	}
	c.JSON(http.StatusOK, gin.H{"preferences": out})
}

// UpsertPreferences PUT /api/notifications/preferences
func (h *NotificationsHandler) UpsertPreferences(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if err := h.ensurePreferencesSchema(c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	var body upsertNotificationPreferencesBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if len(body.Preferences) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "preferences are required"})
		return
	}

	for _, raw := range body.Preferences {
		pref, err := normalizeNotificationPreference(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category in preferences"})
			return
		}
		_, err = db.Pool.Exec(c.Request.Context(), `
			INSERT INTO notification_preferences (user_id, category, enabled, email_enabled, sms_enabled, in_app_enabled, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW())
			ON CONFLICT (user_id, category)
			DO UPDATE SET
				enabled = EXCLUDED.enabled,
				email_enabled = EXCLUDED.email_enabled,
				sms_enabled = EXCLUDED.sms_enabled,
				in_app_enabled = EXCLUDED.in_app_enabled,
				updated_at = NOW()
		`, claims.UserID, pref.Category, pref.Enabled, pref.EmailEnabled, pref.SmsEnabled, pref.InAppEnabled)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"updated": len(body.Preferences)})
}
