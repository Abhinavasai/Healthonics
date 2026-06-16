package handlers

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// extractUrgency tests
// ---------------------------------------------------------------------------

func TestExtractUrgency_Emergency(t *testing.T) {
	cases := []string{
		"Urgency: EMERGENCY. Go to ER immediately.",
		"This is an emergency situation requiring immediate care.",
		"Emergency — call 911 now.",
	}
	for _, s := range cases {
		if got := extractUrgency(s); got != "emergency" {
			t.Errorf("input=%q: expected emergency, got %q", s, got)
		}
	}
}

func TestExtractUrgency_Urgent(t *testing.T) {
	cases := []string{
		"Urgency: Urgent. See a doctor within 24 hours.",
		"This requires urgent medical attention.",
	}
	for _, s := range cases {
		if got := extractUrgency(s); got != "urgent" {
			t.Errorf("input=%q: expected urgent, got %q", s, got)
		}
	}
}

func TestExtractUrgency_SelfCare(t *testing.T) {
	cases := []string{
		"Urgency: Self-care. Rest and monitor at home.",
		"selfcare is appropriate here.",
	}
	for _, s := range cases {
		if got := extractUrgency(s); got != "selfcare" {
			t.Errorf("input=%q: expected selfcare, got %q", s, got)
		}
	}
}

func TestExtractUrgency_Routine_Default(t *testing.T) {
	cases := []string{
		"Schedule an appointment with your doctor.",
		"",
		"Monitor your symptoms over the next few days.",
	}
	for _, s := range cases {
		if got := extractUrgency(s); got != "routine" {
			t.Errorf("input=%q: expected routine (default), got %q", s, got)
		}
	}
}

func TestExtractUrgency_CaseInsensitive(t *testing.T) {
	if got := extractUrgency("EMERGENCY SITUATION"); got != "emergency" {
		t.Errorf("expected emergency for uppercase, got %q", got)
	}
	if got := extractUrgency("URGENT care needed"); got != "urgent" {
		t.Errorf("expected urgent for uppercase, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// symptomFallback tests
// ---------------------------------------------------------------------------

func TestSymptomFallback_ChestPain_Emergency(t *testing.T) {
	reply := symptomFallback("I have chest pain")
	lower := strings.ToLower(reply)
	if !strings.Contains(lower, "emergency") {
		t.Errorf("expected EMERGENCY for chest pain, got: %q", reply)
	}
}

func TestSymptomFallback_BreathingDifficulty_Emergency(t *testing.T) {
	reply := symptomFallback("I can't breathe properly")
	lower := strings.ToLower(reply)
	if !strings.Contains(lower, "emergency") {
		t.Errorf("expected EMERGENCY for breathing difficulty, got: %q", reply)
	}
}

func TestSymptomFallback_FeverWithRash_Emergency(t *testing.T) {
	reply := symptomFallback("I have a fever and a rash")
	lower := strings.ToLower(reply)
	if !strings.Contains(lower, "emergency") {
		t.Errorf("expected EMERGENCY for fever+rash, got: %q", reply)
	}
}

func TestSymptomFallback_FeverWithStiffNeck_Emergency(t *testing.T) {
	reply := symptomFallback("high fever and stiff neck")
	lower := strings.ToLower(reply)
	if !strings.Contains(lower, "emergency") {
		t.Errorf("expected EMERGENCY for fever+stiff neck, got: %q", reply)
	}
}

func TestSymptomFallback_Fever_Urgent(t *testing.T) {
	reply := symptomFallback("I have a fever")
	if !strings.Contains(strings.ToLower(reply), "urgent") {
		t.Errorf("expected urgent for fever, got: %q", reply)
	}
}

func TestSymptomFallback_Headache_Routine(t *testing.T) {
	reply := symptomFallback("headache for two days")
	lower := strings.ToLower(reply)
	if !strings.Contains(lower, "routine") && !strings.Contains(lower, "headache") {
		t.Errorf("expected routine headache reply, got: %q", reply)
	}
}

func TestSymptomFallback_Generic_Routine(t *testing.T) {
	reply := symptomFallback("feeling a bit tired")
	lower := strings.ToLower(reply)
	if !strings.Contains(lower, "routine") {
		t.Errorf("expected routine (generic), got: %q", reply)
	}
}

func TestSymptomFallback_Empty_Routine(t *testing.T) {
	reply := symptomFallback("")
	if reply == "" {
		t.Error("expected non-empty fallback for empty input")
	}
}

// ---------------------------------------------------------------------------
// NewAIHealthHandler constructor tests
// ---------------------------------------------------------------------------

func TestNewAIHealthHandler_DefaultsApplied(t *testing.T) {
	h := NewAIHealthHandler("", "", "", "")
	if h.azureAPIVersion == "" {
		t.Error("expected default api version to be set")
	}
	if h.azureModel == "" {
		t.Error("expected default model to be set")
	}
}

func TestNewAIHealthHandler_CustomValues(t *testing.T) {
	h := NewAIHealthHandler("https://myendpoint.com/", "mykey", "2024-01-01", "gpt-4")
	if h.azureEndpoint != "https://myendpoint.com" {
		t.Errorf("expected trailing slash trimmed, got %q", h.azureEndpoint)
	}
	if h.azureAPIKey != "mykey" {
		t.Errorf("expected api key preserved, got %q", h.azureAPIKey)
	}
	if h.azureModel != "gpt-4" {
		t.Errorf("expected custom model, got %q", h.azureModel)
	}
}
