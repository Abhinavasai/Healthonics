package workers

import (
	"context"
	"log"
	"time"

	"github.com/healthonyx/backend/db"
)

// NotificationWorker marks due pending notifications as sent.
// This is an in-app scheduler foundation; channel adapters (email/sms) can
// plug into this worker in later PRs without changing scheduling semantics.
type NotificationWorker struct {
	interval time.Duration
	now      func() time.Time
	process  func(context.Context) (int64, error)
}

func NewNotificationWorker(interval time.Duration) *NotificationWorker {
	if interval <= 0 {
		interval = time.Minute
	}
	return &NotificationWorker{
		interval: interval,
		now:      time.Now,
		process:  nil,
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

// ProcessDue marks notifications as sent when they are pending and due now.
func (w *NotificationWorker) ProcessDue(ctx context.Context) (int64, error) {
	res, err := db.Pool.Exec(ctx, `
		UPDATE notifications
		SET status = 'sent'
		WHERE status = 'pending'
		  AND (scheduled_for IS NULL OR scheduled_for <= $1)
	`, w.now().UTC())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

func (w *NotificationWorker) processDue(ctx context.Context) (int64, error) {
	if w.process != nil {
		return w.process(ctx)
	}
	return w.ProcessDue(ctx)
}

