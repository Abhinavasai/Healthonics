package workers

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewDocumentSummaryWorker_DefaultInterval(t *testing.T) {
	w := NewDocumentSummaryWorker(0)
	if w == nil {
		t.Fatal("worker should not be nil")
	}
	if w.interval != 2*time.Second {
		t.Fatalf("expected default interval 2s, got %v", w.interval)
	}
}

func TestNewDocumentSummaryWorker_UsesProvidedInterval(t *testing.T) {
	want := 5 * time.Second
	w := NewDocumentSummaryWorker(want)
	if w.interval != want {
		t.Fatalf("expected interval %v, got %v", want, w.interval)
	}
}

func TestDocumentSummaryWorker_ProcessPendingUsesInjectedProcess(t *testing.T) {
	w := NewDocumentSummaryWorker(50 * time.Millisecond)
	calls := 0
	w.process = func(context.Context) (int64, error) {
		calls++
		return 3, nil
	}
	got, err := w.processPending(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 3 {
		t.Fatalf("expected processed 3, got %d", got)
	}
	if calls != 1 {
		t.Fatalf("expected one injected call, got %d", calls)
	}
}

func TestDocumentSummaryWorker_RunHonorsContextCancel(t *testing.T) {
	w := NewDocumentSummaryWorker(10 * time.Millisecond)
	calls := 0
	w.process = func(context.Context) (int64, error) {
		calls++
		return 0, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(done)
	}()
	time.Sleep(25 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("worker did not stop after context cancel")
	}
	if calls == 0 {
		t.Fatal("expected at least one process call")
	}
}

func TestDocumentSummaryWorker_ProcessPendingPropagatesError(t *testing.T) {
	w := NewDocumentSummaryWorker(10 * time.Millisecond)
	wantErr := errors.New("boom")
	w.process = func(context.Context) (int64, error) { return 0, wantErr }
	_, err := w.processPending(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected propagated error, got %v", err)
	}
}
