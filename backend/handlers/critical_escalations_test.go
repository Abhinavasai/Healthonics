package handlers

import (
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestHasCriticalSignal_RuleEngine(t *testing.T) {
	if !hasCriticalSignal("Critical potassium value detected", "ready", "") {
		t.Fatal("expected keyword trigger")
	}
	if !hasCriticalSignal("", "failed", "timeout") {
		t.Fatal("expected failed status trigger")
	}
	if hasCriticalSignal("stable and normal report", "ready", "") {
		t.Fatal("did not expect trigger for benign summary")
	}
}

func TestNextEscalation_SlaTimers(t *testing.T) {
	now := time.Date(2026, 4, 29, 2, 0, 0, 0, time.UTC)

	tier2, next2 := nextEscalation(1, now)
	if tier2 != 2 {
		t.Fatalf("expected tier 2, got %d", tier2)
	}
	if next2.Sub(now) != 15*time.Minute {
		t.Fatalf("expected 15m SLA for tier2, got %v", next2.Sub(now))
	}

	tier3, next3 := nextEscalation(2, now)
	if tier3 != 3 {
		t.Fatalf("expected tier 3, got %d", tier3)
	}
	if next3.Sub(now) != 30*time.Minute {
		t.Fatalf("expected 30m SLA for tier3, got %v", next3.Sub(now))
	}

	tier3b, next3b := nextEscalation(3, now)
	if tier3b != 3 {
		t.Fatalf("tier should remain capped at 3, got %d", tier3b)
	}
	if next3b.Sub(now) != 30*time.Minute {
		t.Fatalf("expected 30m follow-up SLA at tier3, got %v", next3b.Sub(now))
	}
}

func TestIsPgUniqueViolation(t *testing.T) {
	if !isPgUniqueViolation(&pgconn.PgError{Code: "23505"}) {
		t.Fatal("expected true for postgres unique violation")
	}
	if isPgUniqueViolation(&pgconn.PgError{Code: "23503"}) {
		t.Fatal("expected false for non-unique postgres error")
	}
	if isPgUniqueViolation(fmt.Errorf("generic error")) {
		t.Fatal("expected false for non-postgres error")
	}
}
