package handlers

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/healthonyx/backend/failures"
)

type ProviderCallbacksHandler struct {
	sendGridSecret string
	twilioSecret   string
}

func NewProviderCallbacksHandler(sendGridSecret, twilioSecret string) *ProviderCallbacksHandler {
	return &ProviderCallbacksHandler{
		sendGridSecret: strings.TrimSpace(sendGridSecret),
		twilioSecret:   strings.TrimSpace(twilioSecret),
	}
}

type sendGridEvent struct {
	Event      string            `json:"event"`
	Timestamp  int64             `json:"timestamp"`
	Email      string            `json:"email"`
	Reason     string            `json:"reason"`
	SGMessage  string            `json:"sg_message_id"`
	UniqueArgs map[string]string `json:"unique_args"`
}

func verifySendGridSignature(secret, ts, body, gotSig string) bool {
	if strings.TrimSpace(secret) == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(strings.TrimSpace(ts) + "." + body))
	wantHex := hex.EncodeToString(mac.Sum(nil))
	got := strings.ToLower(strings.TrimSpace(gotSig))
	return hmac.Equal([]byte(wantHex), []byte(got))
}

func verifyTwilioSignature(secret, body, gotSig string) bool {
	if strings.TrimSpace(secret) == "" {
		return false
	}
	mac := hmac.New(sha1.New, []byte(secret))
	_, _ = mac.Write([]byte(body))
	want := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(strings.TrimSpace(want)), []byte(strings.TrimSpace(gotSig)))
}

func sendGridFailureTaxonomy(event string) (reason string, suppress bool) {
	switch strings.ToLower(strings.TrimSpace(event)) {
	case "bounce":
		return "bounce", true
	case "dropped":
		return "dropped", true
	case "spamreport":
		return "spam_report", true
	case "blocked":
		return "blocked", true
	case "deferred":
		return "deferred", false
	default:
		return "provider_failure", false
	}
}

func twilioFailureTaxonomy(status string) (reason string, suppress bool) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "undelivered":
		return "undelivered", true
	case "failed":
		return "failed", true
	default:
		return "provider_failure", false
	}
}

func (h *ProviderCallbacksHandler) SendGridWebhook(c *gin.Context) {
	bodyBytes, err := c.GetRawData()
	if err != nil {
		f := failures.ClassifyAPI("NTF_CALLBACK_PAYLOAD_INVALID", "invalid payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "failure": f})
		return
	}
	body := string(bodyBytes)
	ts := c.GetHeader("X-Twilio-Email-Event-Webhook-Timestamp")
	sig := c.GetHeader("X-Twilio-Email-Event-Webhook-Signature")
	valid := verifySendGridSignature(h.sendGridSecret, ts, body, sig)
	if !valid {
		f := failures.ClassifyAPI("NTF_CALLBACK_SIGNATURE_INVALID", "invalid webhook signature")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid webhook signature", "failure": f})
		return
	}

	var events []sendGridEvent
	if err := json.Unmarshal(bodyBytes, &events); err != nil {
		f := failures.ClassifyAPI("NTF_CALLBACK_PAYLOAD_INVALID", "invalid payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "failure": f})
		return
	}
	processed := 0
	duplicates := 0
	for _, e := range events {
		eventID := strings.TrimSpace(e.SGMessage)
		if eventID == "" {
			continue
		}
		var notifID *uuid.UUID
		if e.UniqueArgs != nil {
			if rawID := strings.TrimSpace(e.UniqueArgs["notification_id"]); rawID != "" {
				if id, parseErr := uuid.Parse(rawID); parseErr == nil {
					notifID = &id
				}
			}
		}
		payloadJSON, _ := json.Marshal(e)
		cmd, insErr := db.Pool.Exec(c.Request.Context(), `
			INSERT INTO provider_callback_events(provider, event_id, event_type, notification_id, recipient, signature_valid, payload_json, occurred_at)
			VALUES ('sendgrid', $1, $2, $3, $4, TRUE, $5, to_timestamp($6))
			ON CONFLICT(provider, event_id) DO NOTHING
		`, eventID, strings.ToLower(strings.TrimSpace(e.Event)), nullableUUID(notifID), strings.TrimSpace(strings.ToLower(e.Email)), payloadJSON, e.Timestamp)
		if insErr != nil {
			f := failures.ClassifyAPI("NTF_CALLBACK_PERSIST_FAILED", "internal error")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error", "failure": f})
			return
		}
		if cmd.RowsAffected() == 0 {
			duplicates++
			continue
		}
		processed++
		h.reconcileSendGridEvent(c, e, notifID)
	}
	c.JSON(http.StatusOK, gin.H{"processed": processed, "duplicates": duplicates})
}

func (h *ProviderCallbacksHandler) reconcileSendGridEvent(c *gin.Context, e sendGridEvent, notifID *uuid.UUID) {
	event := strings.ToLower(strings.TrimSpace(e.Event))
	if notifID != nil {
		switch event {
		case "delivered", "processed":
			_, _ = db.Pool.Exec(c.Request.Context(), `
				UPDATE notifications
				SET status = 'sent', provider = 'sendgrid', sent_at = NOW(), last_error = ''
				WHERE id = $1
			`, *notifID)
		default:
			_, suppress := sendGridFailureTaxonomy(event)
			classified := failures.ClassifyCallback("sendgrid", event, strings.TrimSpace(e.Reason))
			_, _ = db.Pool.Exec(c.Request.Context(), `
				UPDATE notifications
				SET status = 'failed', provider = 'sendgrid', last_error = $2, next_retry_at = NULL
				WHERE id = $1
			`, *notifID, classified.Encode())
			if suppress {
				_, _ = db.Pool.Exec(c.Request.Context(), `
					INSERT INTO notification_dead_letters(notification_id, channel, provider, final_error, attempt_count)
					VALUES($1, 'email', 'sendgrid', $2, 1)
					ON CONFLICT(notification_id) DO UPDATE SET final_error = EXCLUDED.final_error, provider = EXCLUDED.provider, attempt_count = GREATEST(notification_dead_letters.attempt_count, EXCLUDED.attempt_count), created_at = NOW()
				`, *notifID, classified.Encode())
			}
		}
	}
	reason, suppress := sendGridFailureTaxonomy(event)
	if suppress && strings.TrimSpace(e.Email) != "" {
		_, _ = db.Pool.Exec(c.Request.Context(), `
			INSERT INTO notification_suppressions(provider, recipient, reason_code, reason_detail, active, last_event_at, updated_at)
			VALUES('sendgrid', $1, $2, $3, TRUE, NOW(), NOW())
			ON CONFLICT(provider, recipient)
			DO UPDATE SET reason_code = EXCLUDED.reason_code, reason_detail = EXCLUDED.reason_detail, active = TRUE, last_event_at = NOW(), updated_at = NOW()
		`, strings.TrimSpace(strings.ToLower(e.Email)), reason, strings.TrimSpace(e.Reason))
	}
}

func (h *ProviderCallbacksHandler) TwilioWebhook(c *gin.Context) {
	bodyBytes, err := c.GetRawData()
	if err != nil {
		f := failures.ClassifyAPI("NTF_CALLBACK_PAYLOAD_INVALID", "invalid payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "failure": f})
		return
	}
	body := string(bodyBytes)
	sig := c.GetHeader("X-Twilio-Signature")
	if !verifyTwilioSignature(h.twilioSecret, body, sig) {
		f := failures.ClassifyAPI("NTF_CALLBACK_SIGNATURE_INVALID", "invalid webhook signature")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid webhook signature", "failure": f})
		return
	}
	messageSID := strings.TrimSpace(c.PostForm("MessageSid"))
	if messageSID == "" {
		f := failures.ClassifyAPI("NTF_CALLBACK_PAYLOAD_INVALID", "missing MessageSid")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing MessageSid", "failure": f})
		return
	}
	status := strings.ToLower(strings.TrimSpace(c.PostForm("MessageStatus")))
	to := strings.TrimSpace(c.PostForm("To"))
	errMsg := strings.TrimSpace(c.PostForm("ErrorMessage"))
	notificationIDRaw := strings.TrimSpace(c.PostForm("NotificationId"))
	var notifID *uuid.UUID
	if notificationIDRaw != "" {
		if id, parseErr := uuid.Parse(notificationIDRaw); parseErr == nil {
			notifID = &id
		}
	}
	payloadJSON, _ := json.Marshal(map[string]string{
		"message_sid":     messageSID,
		"message_status":  status,
		"to":              to,
		"error_message":   errMsg,
		"notification_id": notificationIDRaw,
	})
	cmd, insErr := db.Pool.Exec(c.Request.Context(), `
		INSERT INTO provider_callback_events(provider, event_id, event_type, notification_id, recipient, signature_valid, payload_json, occurred_at)
		VALUES('twilio', $1, $2, $3, $4, TRUE, $5, NOW())
		ON CONFLICT(provider, event_id) DO NOTHING
	`, messageSID, status, nullableUUID(notifID), to, payloadJSON)
	if insErr != nil {
		f := failures.ClassifyAPI("NTF_CALLBACK_PERSIST_FAILED", "internal error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error", "failure": f})
		return
	}
	if cmd.RowsAffected() == 0 {
		c.JSON(http.StatusOK, gin.H{"processed": 0, "duplicates": 1})
		return
	}
	h.reconcileTwilioEvent(c, status, to, errMsg, notifID)
	c.JSON(http.StatusOK, gin.H{"processed": 1, "duplicates": 0})
}

func (h *ProviderCallbacksHandler) reconcileTwilioEvent(c *gin.Context, status, to, errMsg string, notifID *uuid.UUID) {
	if notifID != nil {
		switch status {
		case "delivered", "sent":
			_, _ = db.Pool.Exec(c.Request.Context(), `
				UPDATE notifications
				SET status = 'sent', provider = 'twilio', sent_at = NOW(), last_error = ''
				WHERE id = $1
			`, *notifID)
		default:
			_, suppress := twilioFailureTaxonomy(status)
			classified := failures.ClassifyCallback("twilio", status, errMsg)
			_, _ = db.Pool.Exec(c.Request.Context(), `
				UPDATE notifications
				SET status = 'failed', provider = 'twilio', last_error = $2, next_retry_at = NULL
				WHERE id = $1
			`, *notifID, classified.Encode())
			if suppress {
				_, _ = db.Pool.Exec(c.Request.Context(), `
					INSERT INTO notification_dead_letters(notification_id, channel, provider, final_error, attempt_count)
					VALUES($1, 'sms', 'twilio', $2, 1)
					ON CONFLICT(notification_id) DO UPDATE SET final_error = EXCLUDED.final_error, provider = EXCLUDED.provider, attempt_count = GREATEST(notification_dead_letters.attempt_count, EXCLUDED.attempt_count), created_at = NOW()
				`, *notifID, classified.Encode())
			}
		}
	}
	reason, suppress := twilioFailureTaxonomy(status)
	if suppress && strings.TrimSpace(to) != "" {
		_, _ = db.Pool.Exec(c.Request.Context(), `
			INSERT INTO notification_suppressions(provider, recipient, reason_code, reason_detail, active, last_event_at, updated_at)
			VALUES('twilio', $1, $2, $3, TRUE, NOW(), NOW())
			ON CONFLICT(provider, recipient)
			DO UPDATE SET reason_code = EXCLUDED.reason_code, reason_detail = EXCLUDED.reason_detail, active = TRUE, last_event_at = NOW(), updated_at = NOW()
		`, strings.TrimSpace(to), reason, strings.TrimSpace(errMsg))
	}
}

func nullableUUID(id *uuid.UUID) any {
	if id == nil {
		return sql.NullString{}
	}
	return *id
}
