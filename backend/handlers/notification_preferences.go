package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/healthonyx/backend/db"
	"github.com/jackc/pgx/v5"
)

type NotificationPreferencesHandler struct{}

func NewNotificationPreferencesHandler() *NotificationPreferencesHandler {
	return &NotificationPreferencesHandler{}
}

type notificationPrefsJSON struct {
	EmailAppointmentReminders bool `json:"email_appointment_reminders"`
	SmsAppointmentReminders   bool `json:"sms_appointment_reminders"`
}

// Get returns notification preferences for the current user (defaults if missing).
func (h *NotificationPreferencesHandler) Get(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()

	var emailOn, smsOn bool
	err := db.Pool.QueryRow(ctx, `
		SELECT email_appointment_reminders, sms_appointment_reminders
		FROM user_notification_preferences WHERE user_id = $1
	`, claims.UserID).Scan(&emailOn, &smsOn)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusOK, notificationPrefsJSON{
				EmailAppointmentReminders: true,
				SmsAppointmentReminders:   false,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusOK, notificationPrefsJSON{
		EmailAppointmentReminders: emailOn,
		SmsAppointmentReminders:   smsOn,
	})
}

// Put updates notification preferences.
func (h *NotificationPreferencesHandler) Put(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	var req notificationPrefsJSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	_, err := db.Pool.Exec(c.Request.Context(), `
		INSERT INTO user_notification_preferences (user_id, email_appointment_reminders, sms_appointment_reminders, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			email_appointment_reminders = EXCLUDED.email_appointment_reminders,
			sms_appointment_reminders = EXCLUDED.sms_appointment_reminders,
			updated_at = NOW()
	`, claims.UserID, req.EmailAppointmentReminders, req.SmsAppointmentReminders)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusOK, req)
}
