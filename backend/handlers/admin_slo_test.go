package handlers

import "testing"

func TestSLOThresholdAlerts(t *testing.T) {
	overview := adminSLOOverview{
		Notifications: adminSLOMetric{ErrorBudgetUsed: 120, P95LatencyMS: 2200},
		AI:            adminSLOMetric{ErrorBudgetUsed: 55, P95LatencyMS: 900},
		API:           adminSLOMetric{ErrorBudgetUsed: 10, P95LatencyMS: 2500},
	}
	alerts := buildSLOAlerts(overview)
	if len(alerts) < 4 {
		t.Fatalf("expected multiple alerts, got %d", len(alerts))
	}
}

func TestSLOMetricCalculations(t *testing.T) {
	if got := successPercent(100, 2); got < 97.9 || got > 98.1 {
		t.Fatalf("unexpected success percent: %f", got)
	}
	if got := errorBudgetUsed(98, 99); got < 199.9 || got > 200.1 {
		t.Fatalf("unexpected error budget used: %f", got)
	}
}
