package handlers

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// buildSlotReason tests
// ---------------------------------------------------------------------------

func TestBuildSlotReason_HighPrefLowNoshow_Morning(t *testing.T) {
	reason := buildSlotReason(0.8, 0.05, 9) // morning, high pref, low noshow
	if !strings.Contains(strings.ToLower(reason), "morning") {
		t.Errorf("expected 'morning' in reason, got: %q", reason)
	}
	if !strings.Contains(strings.ToLower(reason), "preferred") {
		t.Errorf("expected 'preferred' keyword in high-match reason, got: %q", reason)
	}
}

func TestBuildSlotReason_HighPrefLowNoshow_Afternoon(t *testing.T) {
	reason := buildSlotReason(0.8, 0.05, 14)
	if !strings.Contains(strings.ToLower(reason), "afternoon") {
		t.Errorf("expected 'afternoon' in reason, got: %q", reason)
	}
}

func TestBuildSlotReason_HighPrefLowNoshow_Evening(t *testing.T) {
	reason := buildSlotReason(0.8, 0.05, 18)
	if !strings.Contains(strings.ToLower(reason), "evening") {
		t.Errorf("expected 'evening' in reason, got: %q", reason)
	}
}

func TestBuildSlotReason_MedPref(t *testing.T) {
	reason := buildSlotReason(0.5, 0.3, 10)
	if reason == "" {
		t.Error("expected non-empty reason for medium pref score")
	}
	if !strings.Contains(strings.ToLower(reason), "morning") {
		t.Errorf("expected 'morning' in reason for hour=10, got: %q", reason)
	}
}

func TestBuildSlotReason_LowPrefLowNoshow(t *testing.T) {
	reason := buildSlotReason(0.1, 0.02, 9)
	if !strings.Contains(strings.ToLower(reason), "reliability") {
		t.Errorf("expected 'reliability' keyword for low pref, low noshow, got: %q", reason)
	}
}

func TestBuildSlotReason_LowPrefHighNoshow(t *testing.T) {
	reason := buildSlotReason(0.1, 0.5, 9)
	if reason == "" {
		t.Error("expected non-empty fallback reason")
	}
}

func TestBuildSlotReason_ZeroPrefZeroNoshow(t *testing.T) {
	// New patient: no history, no noshow — should get reliability path.
	reason := buildSlotReason(0, 0, 8)
	if reason == "" {
		t.Error("expected non-empty reason for zero pref, zero noshow")
	}
}

// ---------------------------------------------------------------------------
// Score calculation property tests
// ---------------------------------------------------------------------------

func TestSlotScore_PerfectPrefAndReliability(t *testing.T) {
	prefScore := 1.0
	noshowPenalty := 0.0
	score := prefScore*0.6 + (1-noshowPenalty)*0.4
	if score != 1.0 {
		t.Errorf("expected perfect score 1.0, got %f", score)
	}
}

func TestSlotScore_NoPrefFullNoshow(t *testing.T) {
	prefScore := 0.0
	noshowPenalty := 1.0
	score := prefScore*0.6 + (1-noshowPenalty)*0.4
	if score != 0.0 {
		t.Errorf("expected worst score 0.0, got %f", score)
	}
}

func TestSlotScore_NewPatientNoHistory(t *testing.T) {
	// No pref history → prefScore=0; score is purely reliability-based.
	prefScore := 0.0
	noshowPenalty := 0.1
	score := prefScore*0.6 + (1-noshowPenalty)*0.4
	expected := 0.9 * 0.4
	diff := score - expected
	if diff < -1e-9 || diff > 1e-9 {
		t.Errorf("expected %f, got %f", expected, score)
	}
}

func TestSlotScore_AlwaysInRange(t *testing.T) {
	// For any valid pref (0-1) and noshow (0-1), score must be in [0,1].
	cases := [][2]float64{{0, 0}, {0, 1}, {1, 0}, {1, 1}, {0.5, 0.5}, {0.7, 0.2}}
	for _, c := range cases {
		score := c[0]*0.6 + (1-c[1])*0.4
		if score < 0 || score > 1 {
			t.Errorf("pref=%f noshow=%f: score %f out of range [0,1]", c[0], c[1], score)
		}
	}
}
