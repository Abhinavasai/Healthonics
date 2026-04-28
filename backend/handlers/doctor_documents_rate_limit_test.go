package handlers

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAllowSummarizeForDoctor_RateLimit(t *testing.T) {
	doctorID := uuid.New()

	summarizeRateMu.Lock()
	summarizeRateByDoctor = map[uuid.UUID]summarizeRateBucket{}
	summarizeRateMu.Unlock()

	if !allowSummarizeForDoctor(doctorID, 2) {
		t.Fatalf("first call should be allowed")
	}
	if !allowSummarizeForDoctor(doctorID, 2) {
		t.Fatalf("second call should be allowed")
	}
	if allowSummarizeForDoctor(doctorID, 2) {
		t.Fatalf("third call should be blocked by rate limit")
	}
}

func TestAllowSummarizeForDoctor_WindowReset(t *testing.T) {
	doctorID := uuid.New()

	summarizeRateMu.Lock()
	summarizeRateByDoctor = map[uuid.UUID]summarizeRateBucket{
		doctorID: {
			windowStart: time.Now().Add(-2 * time.Minute),
			count:       50,
		},
	}
	summarizeRateMu.Unlock()

	if !allowSummarizeForDoctor(doctorID, 1) {
		t.Fatalf("window should reset after one minute")
	}
	if allowSummarizeForDoctor(doctorID, 1) {
		t.Fatalf("second call in new window should be blocked for limit=1")
	}
}
