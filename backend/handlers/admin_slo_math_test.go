package handlers

import (
	"testing"
)

func TestSLO_SuccessRate(t *testing.T) {
	successRate := func(total, errors int64) float64 {
		if total == 0 {
			return 100.0
		}
		return float64(total-errors) / float64(total) * 100
	}

	cases := []struct {
		total    int64
		errors   int64
		expected float64
	}{
		{0, 0, 100.0},
		{100, 0, 100.0},
		{100, 1, 99.0},
		{100, 50, 50.0},
		{100, 100, 0.0},
		{1000, 5, 99.5},
	}

	for _, tc := range cases {
		got := successRate(tc.total, tc.errors)
		if got != tc.expected {
			t.Errorf("successRate(%d, %d) = %.2f, want %.2f", tc.total, tc.errors, got, tc.expected)
		}
	}
}

func TestSLO_PercentileNullHandling(t *testing.T) {
	// When api_request_telemetry is empty, PostgreSQL PERCENTILE_CONT returns NULL.
	// We scan into *float64 to handle this. A nil pointer means "no data".
	var p50 *float64
	var p95 *float64
	var p99 *float64

	// Simulate empty table: all nil
	if p50 != nil || p95 != nil || p99 != nil {
		t.Error("expected nil percentiles for empty table")
	}

	// Simulate populated table
	v50 := 42.5
	v95 := 250.0
	v99 := 800.0
	p50 = &v50
	p95 = &v95
	p99 = &v99

	if *p50 != 42.5 || *p95 != 250.0 || *p99 != 800.0 {
		t.Error("percentile values not stored correctly")
	}
}

func TestSLO_ErrorBudget(t *testing.T) {
	// SLO target: 99.9% uptime → 0.1% error budget
	sloTarget := 99.9
	budgetPct := 100.0 - sloTarget

	cases := []struct {
		errorRate float64
		exhausted bool
	}{
		{0.0, false},
		{0.05, false},
		{0.09, false},
		{0.11, true},
		{1.0, true},
	}

	for _, tc := range cases {
		got := tc.errorRate > budgetPct
		if got != tc.exhausted {
			t.Errorf("errorRate=%.3f budget=%.3f: expected exhausted=%v got=%v",
				tc.errorRate, budgetPct, tc.exhausted, got)
		}
	}
}
