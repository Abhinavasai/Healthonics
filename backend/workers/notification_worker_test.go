package workers

import (
	"context"
	"testing"
	"time"
)

func TestNewNotificationWorker_DefaultInterval(t *testing.T) {
	w := NewNotificationWorker(0)
	if w == nil {
		t.Fatal("expected worker, got nil")
	}
	if w.interval != time.Minute {
		t.Fatalf("expected default interval %v, got %v", time.Minute, w.interval)
	}
}

func TestNewNotificationWorker_UsesProvidedInterval(t *testing.T) {
	want := 15 * time.Second
	w := NewNotificationWorker(want)
	if w.interval != want {
		t.Fatalf("expected interval %v, got %v", want, w.interval)
	}
}

func TestNotificationWorker_RunStopsOnContextCancel(t *testing.T) {
	w := NewNotificationWorker(10 * time.Millisecond)
	// Avoid DB dependency in this unit test; worker should still stop gracefully.
	w.process = func(context.Context) (int64, error) { return 0, nil }

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)
		w.Run(ctx)
	}()

	time.Sleep(15 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// ok
	case <-time.After(200 * time.Millisecond):
		t.Fatal("worker did not stop after context cancellation")
	}
}

func TestBackoffDuration(t *testing.T) {
	if got := backoffDuration(1); got != 15*time.Second {
		t.Fatalf("attempt 1: want 15s, got %v", got)
	}
	if got := backoffDuration(2); got != 60*time.Second {
		t.Fatalf("attempt 2: want 60s, got %v", got)
	}
	if got := backoffDuration(3); got != 5*time.Minute {
		t.Fatalf("attempt 3: want 5m, got %v", got)
	}
}
