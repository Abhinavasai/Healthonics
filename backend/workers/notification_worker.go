package workers

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/healthonyx/backend/db"
	"github.com/healthonyx/backend/failures"
	"github.com/healthonyx/backend/providers"
	"github.com/jackc/pgx/v5"
)

type NotificationProviderSet struct {
	email providers.EmailProvider
	sms   providers.SMSProvider
}

func NewNotificationProviderSet(email providers.EmailProvider, sms providers.SMSProvider) NotificationProviderSet {
	return NotificationProviderSet{
		email: email,
		sms:   sms,
	}
}

type NotificationWorker struct {
	interval   time.Duration
	maxRetries int
	now        func() time.Time
	process    func(context.Context) (int64, error)
	providers  NotificationProviderSet
}

type NotificationWorkerConfig struct {
	Interval   time.Duration
	MaxRetries int
	Providers  NotificationProviderSet
}

type queuedNotification struct {
	ID       string
	To       string
	Title    string
	Body     string
	Channel  string
	Provider string
	Attempts int
}

func NewNotificationWorker(interval time.Duration) *NotificationWorker {
	return NewNotificationWorkerWithConfig(NotificationWorkerConfig{
		Interval:   interval,
		MaxRetries: 3,
	})
}

func NewNotificationWorkerWithConfig(cfg NotificationWorkerConfig) *NotificationWorker {
	interval := cfg.Interval
	if interval <= 0 {
		interval = time.Minute
	}
	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	return &NotificationWorker{
		interval:   interval,
		maxRetries: maxRetries,
		now:        time.Now,
		process:    nil,
		providers:  cfg.Providers,
	}
}

// Run starts a ticker loop and continues until context cancellation.
func (w *NotificationWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Run once on startup to catch existing due jobs.
	if _, err := w.processDue(ctx); err != nil {
		log.Printf("notification worker initial process error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := w.processDue(ctx); err != nil {
				log.Printf("notification worker process error: %v", err)
			}
		}
	}
}

func (w *NotificationWorker) ProcessDue(ctx context.Context) (int64, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT n.id::text, n.user_id::text, n.title, n.body, n.channel, n.provider, n.attempts
		FROM notifications n
		WHERE n.status = 'pending'
		  AND (n.scheduled_for IS NULL OR n.scheduled_for <= $1)
		  AND (n.next_retry_at IS NULL OR n.next_retry_at <= $1)
		ORDER BY COALESCE(n.next_retry_at, n.scheduled_for, n.created_at) ASC
		LIMIT 100
	`, w.now().UTC())
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	processed := int64(0)
	for rows.Next() {
		var n queuedNotification
		if err := rows.Scan(&n.ID, &n.To, &n.Title, &n.Body, &n.Channel, &n.Provider, &n.Attempts); err != nil {
			return processed, err
		}
		if err := w.processSingle(ctx, n); err != nil {
			return processed, err
		}
		processed++
	}
	return processed, rows.Err()
}

func (w *NotificationWorker) processDue(ctx context.Context) (int64, error) {
	if w.process != nil {
		return w.process(ctx)
	}
	return w.ProcessDue(ctx)
}

func (w *NotificationWorker) processSingle(ctx context.Context, n queuedNotification) error {
	if suppressed, reason, err := w.isSuppressed(ctx, n); err != nil {
		return err
	} else if suppressed {
		f := failures.ClassifySuppression(n.Provider, reason)
		nextAttempt := n.Attempts + 1
		if _, err := db.Pool.Exec(ctx, `
			INSERT INTO notification_delivery_attempts(notification_id, channel, provider, attempt_no, success, error_message, latency_ms)
			VALUES($1, $2, $3, $4, FALSE, $5, 0)
		`, n.ID, n.Channel, n.Provider, nextAttempt, f.Encode()); err != nil {
			log.Printf("notification worker: failed to insert delivery attempt for %s: %v", n.ID, err)
		}
		if _, err := db.Pool.Exec(ctx, `
			INSERT INTO notification_dead_letters(notification_id, channel, provider, final_error, attempt_count)
			VALUES($1, $2, $3, $4, $5)
			ON CONFLICT(notification_id) DO UPDATE SET
				provider = EXCLUDED.provider,
				final_error = EXCLUDED.final_error,
				attempt_count = EXCLUDED.attempt_count,
				created_at = NOW()
		`, n.ID, n.Channel, n.Provider, f.Encode(), nextAttempt); err != nil {
			log.Printf("notification worker: failed to insert dead letter for %s: %v", n.ID, err)
		}
		_, updateErr := db.Pool.Exec(ctx, `
			UPDATE notifications
			SET status = 'failed',
			    attempts = $2,
			    last_error = $3,
			    next_retry_at = NULL
			WHERE id = $1
		`, n.ID, nextAttempt, f.Encode())
		return updateErr
	}

	startedAt := time.Now()
	mode, err := w.deliver(ctx, n)
	latencyMS := int(time.Since(startedAt).Milliseconds())
	if latencyMS < 0 {
		latencyMS = 0
	}
	nextAttempt := n.Attempts + 1
	classified := failures.ClassifyProviderSend(mode, err)
	if _, execErr := db.Pool.Exec(ctx, `
		INSERT INTO notification_delivery_attempts(notification_id, channel, provider, attempt_no, success, error_message, latency_ms)
		VALUES($1, $2, $3, $4, $5, $6, $7)
	`, n.ID, n.Channel, mode, nextAttempt, err == nil, errorStringOrClassified(err, classified), latencyMS); execErr != nil {
		log.Printf("notification worker: failed to insert delivery attempt for %s: %v", n.ID, execErr)
	}

	if err == nil {
		_, updateErr := db.Pool.Exec(ctx, `
			UPDATE notifications
			SET status = 'sent',
			    sent_at = $2,
			    attempts = $3,
			    provider = $4,
			    next_retry_at = NULL,
			    last_error = ''
			WHERE id = $1
		`, n.ID, w.now().UTC(), nextAttempt, mode)
		return updateErr
	}

	if nextAttempt >= w.maxRetries {
		if _, deadErr := db.Pool.Exec(ctx, `
			INSERT INTO notification_dead_letters(notification_id, channel, provider, final_error, attempt_count)
			VALUES($1, $2, $3, $4, $5)
			ON CONFLICT(notification_id) DO UPDATE SET
				provider = EXCLUDED.provider,
				final_error = EXCLUDED.final_error,
				attempt_count = EXCLUDED.attempt_count,
				created_at = NOW()
		`, n.ID, n.Channel, mode, classified.Encode(), nextAttempt); deadErr != nil {
			return deadErr
		}
		_, updateErr := db.Pool.Exec(ctx, `
			UPDATE notifications
			SET status = 'failed',
			    attempts = $2,
			    provider = $3,
			    last_error = $4,
			    next_retry_at = NULL
			WHERE id = $1
		`, n.ID, nextAttempt, mode, classified.Encode())
		return updateErr
	}

	nextRetry := w.now().UTC().Add(backoffDuration(nextAttempt))
	_, updateErr := db.Pool.Exec(ctx, `
		UPDATE notifications
		SET status = 'pending',
		    attempts = $2,
		    provider = $3,
		    last_error = $4,
		    next_retry_at = $5
		WHERE id = $1
	`, n.ID, nextAttempt, mode, classified.Encode(), nextRetry)
	return updateErr
}

func (w *NotificationWorker) isSuppressed(ctx context.Context, n queuedNotification) (bool, string, error) {
	channel := strings.ToLower(strings.TrimSpace(n.Channel))
	recipient := strings.TrimSpace(strings.ToLower(n.To))
	if recipient == "" || (channel != "email" && channel != "sms") {
		return false, "", nil
	}
	provider := "sendgrid"
	if channel == "sms" {
		provider = "twilio"
	}
	var reason sql.NullString
	err := db.Pool.QueryRow(ctx, `
		SELECT reason_code
		FROM notification_suppressions
		WHERE provider = $1 AND recipient = $2 AND active = TRUE
		LIMIT 1
	`, provider, recipient).Scan(&reason)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, "", nil
		}
		return false, "", err
	}
	return true, reason.String, nil
}

func (w *NotificationWorker) deliver(ctx context.Context, n queuedNotification) (string, error) {
	ch := strings.ToLower(strings.TrimSpace(n.Channel))
	switch ch {
	case "", "in_app":
		return "in_app", nil
	case "email":
		if w.providers.email == nil {
			return "email", errors.New("email provider not configured")
		}
		err := w.providers.email.Send(ctx, providers.DeliveryRequest{
			To:      strings.TrimSpace(n.To),
			Subject: strings.TrimSpace(n.Title),
			Body:    n.Body,
		})
		return "sendgrid", err
	case "sms":
		if w.providers.sms == nil {
			return "sms", errors.New("sms provider not configured")
		}
		err := w.providers.sms.Send(ctx, providers.DeliveryRequest{
			To:   strings.TrimSpace(n.To),
			Body: n.Body,
		})
		return "twilio", err
	default:
		return ch, errors.New("unsupported notification channel")
	}
}

func backoffDuration(attemptNo int) time.Duration {
	if attemptNo <= 1 {
		return 15 * time.Second
	}
	if attemptNo == 2 {
		return 60 * time.Second
	}
	return 5 * time.Minute
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func errorStringOrClassified(err error, classified failures.Failure) string {
	if err == nil {
		return ""
	}
	return classified.Encode()
}
