package handlers

import (
	"testing"
)

// ---------------------------------------------------------------------------
// adherenceLabel tests
// ---------------------------------------------------------------------------

func TestAdherenceLabel_Good(t *testing.T) {
	cases := []int{80, 90, 100}
	for _, s := range cases {
		if got := adherenceLabel(s); got != "Good" {
			t.Errorf("score %d: expected Good, got %q", s, got)
		}
	}
}

func TestAdherenceLabel_Fair(t *testing.T) {
	cases := []int{60, 70, 79}
	for _, s := range cases {
		if got := adherenceLabel(s); got != "Fair" {
			t.Errorf("score %d: expected Fair, got %q", s, got)
		}
	}
}

func TestAdherenceLabel_Poor(t *testing.T) {
	cases := []int{40, 50, 59}
	for _, s := range cases {
		if got := adherenceLabel(s); got != "Poor" {
			t.Errorf("score %d: expected Poor, got %q", s, got)
		}
	}
}

func TestAdherenceLabel_VeryPoor(t *testing.T) {
	cases := []int{0, 1, 39}
	for _, s := range cases {
		if got := adherenceLabel(s); got != "Very Poor" {
			t.Errorf("score %d: expected Very Poor, got %q", s, got)
		}
	}
}

func TestAdherenceLabel_Boundaries(t *testing.T) {
	if got := adherenceLabel(80); got != "Good" {
		t.Errorf("expected Good at boundary 80, got %q", got)
	}
	if got := adherenceLabel(60); got != "Fair" {
		t.Errorf("expected Fair at boundary 60, got %q", got)
	}
	if got := adherenceLabel(40); got != "Poor" {
		t.Errorf("expected Poor at boundary 40, got %q", got)
	}
}
