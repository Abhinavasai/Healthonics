package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/healthonyx/backend/failures"
	"github.com/jackc/pgx/v5"
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
	Preferences   []NotificationPreference `json:"preferences"`
	ConsentSource string                   `json:"consent_source"`
	PolicyVersion string                   `json:"policy_version"`
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

func normalizeConsentSource(raw string) string {
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" {
		return "self_service_portal"
	}
	switch s {
	case "self_service_portal", "admin_console", "support_assisted", "api":
		return s
	default:
		return "unknown"
	}
}

func normalizePolicyVersion(raw string) string {
	s := strings.TrimSpace(raw)
	if s != "" {
		return s
	}
	env := strings.TrimSpace(os.Getenv("NOTIFICATION_POLICY_VERSION"))
	if env != "" {
		return env
	}
	return "v1"
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

		CREATE TABLE IF NOT EXISTS notification_consent_events (
			id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id            UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			actor_user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			actor_role         TEXT NOT NULL,
			actor_source       TEXT NOT NULL DEFAULT 'self_service_portal',
			policy_version     TEXT NOT NULL DEFAULT 'v1',
			category           TEXT NOT NULL,
			prev_enabled       BOOLEAN NOT NULL,
			prev_email_enabled BOOLEAN NOT NULL,
			prev_sms_enabled   BOOLEAN NOT NULL,
			prev_in_app_enabled BOOLEAN NOT NULL,
			new_enabled        BOOLEAN NOT NULL,
			new_email_enabled  BOOLEAN NOT NULL,
			new_sms_enabled    BOOLEAN NOT NULL,
			new_in_app_enabled BOOLEAN NOT NULL,
			created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_notification_consent_events_user ON notification_consent_events(user_id, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_notification_consent_events_actor ON notification_consent_events(actor_user_id, created_at DESC);
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
		SELECT id::text, title, body, channel, status, provider, attempts, last_error, scheduled_for::text, next_retry_at::text, sent_at::text, created_at::text
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
		Provider     string  `json:"provider"`
		Attempts     int     `json:"attempts"`
		LastError    string  `json:"last_error"`
		ScheduledFor *string `json:"scheduled_for"`
		NextRetryAt  *string `json:"next_retry_at"`
		SentAt       *string `json:"sent_at"`
		CreatedAt    string  `json:"created_at"`
	}
	var out []row
	for rows.Next() {
		var r row
		var sched sql.NullString
		var nextRetry sql.NullString
		var sentAt sql.NullString
		if err := rows.Scan(&r.ID, &r.Title, &r.Body, &r.Channel, &r.Status, &r.Provider, &r.Attempts, &r.LastError, &sched, &nextRetry, &sentAt, &r.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		if sched.Valid {
			s := sched.String
			r.ScheduledFor = &s
		}
		if nextRetry.Valid {
			s := nextRetry.String
			r.NextRetryAt = &s
		}
		if sentAt.Valid {
			s := sentAt.String
			r.SentAt = &s
		}
		out = append(out, r)
	}
	if out == nil {
		out = []row{}
	}
	c.JSON(http.StatusOK, gin.H{"notifications": out})
}

// AdminList GET /api/admin/notifications
func (h *NotificationsHandler) AdminList(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}
	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id::text, user_id::text, title, body, channel, status, provider, attempts, last_error, scheduled_for::text, next_retry_at::text, sent_at::text, created_at::text
		FROM notifications
		ORDER BY created_at DESC
		LIMIT 300
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	type row struct {
		ID           string  `json:"id"`
		UserID       string  `json:"user_id"`
		Title        string  `json:"title"`
		Body         string  `json:"body"`
		Channel      string  `json:"channel"`
		Status       string  `json:"status"`
		Provider     string  `json:"provider"`
		Attempts     int     `json:"attempts"`
		LastError    string  `json:"last_error"`
		ScheduledFor *string `json:"scheduled_for"`
		NextRetryAt  *string `json:"next_retry_at"`
		SentAt       *string `json:"sent_at"`
		CreatedAt    string  `json:"created_at"`
	}
	var out []row
	for rows.Next() {
		var r row
		var sched sql.NullString
		var nextRetry sql.NullString
		var sentAt sql.NullString
		if err := rows.Scan(&r.ID, &r.UserID, &r.Title, &r.Body, &r.Channel, &r.Status, &r.Provider, &r.Attempts, &r.LastError, &sched, &nextRetry, &sentAt, &r.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		if sched.Valid {
			s := sched.String
			r.ScheduledFor = &s
		}
		if nextRetry.Valid {
			s := nextRetry.String
			r.NextRetryAt = &s
		}
		if sentAt.Valid {
			s := sentAt.String
			r.SentAt = &s
		}
		out = append(out, r)
	}
	if out == nil {
		out = []row{}
	}
	c.JSON(http.StatusOK, gin.H{"notifications": out})
}

// AdminSummary GET /api/admin/notifications/summary
func (h *NotificationsHandler) AdminSummary(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}
	var pending, sent, failed int64
	err := db.Pool.QueryRow(c.Request.Context(), `
		SELECT
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'sent'),
			COUNT(*) FILTER (WHERE status = 'failed')
		FROM notifications
	`).Scan(&pending, &sent, &failed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"pending_count": pending,
		"sent_count":    sent,
		"failed_count":  failed,
	})
}

// RetryFailed POST /api/admin/notifications/:id/retry
func (h *NotificationsHandler) RetryFailed(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}
	id, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		f := failures.ClassifyAPI("NTF_RETRY_ID_INVALID", "invalid notification id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification id", "failure": f})
		return
	}
	var prevStatus string
	err = db.Pool.QueryRow(c.Request.Context(), `
		UPDATE notifications
		SET status = 'pending',
		    scheduled_for = NOW(),
		    attempts = 0,
		    next_retry_at = NULL,
		    last_error = '',
		    provider = ''
		WHERE id = $1
		  AND status = 'failed'
		RETURNING status
	`, id).Scan(&prevStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			f := failures.ClassifyAPI("NTF_RETRY_TARGET_MISSING", "failed notification not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "Failed notification not found", "failure": f})
			return
		}
		f := failures.ClassifyAPI("NTF_RETRY_UPDATE_FAILED", "internal error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error", "failure": f})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":     id.String(),
		"status": "pending",
	})
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
	consentSource := normalizeConsentSource(body.ConsentSource)
	policyVersion := normalizePolicyVersion(body.PolicyVersion)

	currentRows, err := db.Pool.Query(c.Request.Context(), `
		SELECT category, enabled, email_enabled, sms_enabled, in_app_enabled
		FROM notification_preferences
		WHERE user_id = $1
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	current := map[string]NotificationPreference{}
	for currentRows.Next() {
		var row NotificationPreference
		if err := currentRows.Scan(&row.Category, &row.Enabled, &row.EmailEnabled, &row.SmsEnabled, &row.InAppEnabled); err != nil {
			currentRows.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		current[row.Category] = row
	}
	currentRows.Close()

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

		prev, ok := current[pref.Category]
		if !ok {
			prev = NotificationPreference{
				Category:     pref.Category,
				Enabled:      true,
				EmailEnabled: true,
				SmsEnabled:   false,
				InAppEnabled: true,
			}
		}
		changed := prev.Enabled != pref.Enabled ||
			prev.EmailEnabled != pref.EmailEnabled ||
			prev.SmsEnabled != pref.SmsEnabled ||
			prev.InAppEnabled != pref.InAppEnabled
		if changed {
			_, err = db.Pool.Exec(c.Request.Context(), `
				INSERT INTO notification_consent_events (
					user_id, actor_user_id, actor_role, actor_source, policy_version, category,
					prev_enabled, prev_email_enabled, prev_sms_enabled, prev_in_app_enabled,
					new_enabled, new_email_enabled, new_sms_enabled, new_in_app_enabled
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			`,
				claims.UserID, claims.UserID, claims.Role, consentSource, policyVersion, pref.Category,
				prev.Enabled, prev.EmailEnabled, prev.SmsEnabled, prev.InAppEnabled,
				pref.Enabled, pref.EmailEnabled, pref.SmsEnabled, pref.InAppEnabled,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"updated": len(body.Preferences), "policy_version": policyVersion})
}

// AdminConsentHistory GET /api/admin/notifications/consent-history
func (h *NotificationsHandler) AdminConsentHistory(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}
	if err := h.ensurePreferencesSchema(c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id::text, user_id::text, actor_user_id::text, actor_role, actor_source, policy_version,
		       category, prev_enabled, prev_email_enabled, prev_sms_enabled, prev_in_app_enabled,
		       new_enabled, new_email_enabled, new_sms_enabled, new_in_app_enabled, created_at::text
		FROM notification_consent_events
		ORDER BY created_at DESC
		LIMIT 300
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	type row struct {
		ID            string `json:"id"`
		UserID        string `json:"user_id"`
		ActorUserID   string `json:"actor_user_id"`
		ActorRole     string `json:"actor_role"`
		ActorSource   string `json:"actor_source"`
		PolicyVersion string `json:"policy_version"`
		Category      string `json:"category"`
		PrevEnabled   bool   `json:"prev_enabled"`
		PrevEmail     bool   `json:"prev_email_enabled"`
		PrevSMS       bool   `json:"prev_sms_enabled"`
		PrevInApp     bool   `json:"prev_in_app_enabled"`
		NewEnabled    bool   `json:"new_enabled"`
		NewEmail      bool   `json:"new_email_enabled"`
		NewSMS        bool   `json:"new_sms_enabled"`
		NewInApp      bool   `json:"new_in_app_enabled"`
		CreatedAt     string `json:"created_at"`
	}
	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(
			&r.ID, &r.UserID, &r.ActorUserID, &r.ActorRole, &r.ActorSource, &r.PolicyVersion,
			&r.Category, &r.PrevEnabled, &r.PrevEmail, &r.PrevSMS, &r.PrevInApp,
			&r.NewEnabled, &r.NewEmail, &r.NewSMS, &r.NewInApp, &r.CreatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		out = append(out, r)
	}
	if out == nil {
		out = []row{}
	}
	c.JSON(http.StatusOK, gin.H{"events": out})
}
