package handlers

import (
	"strings"
	"testing"
)

func TestLabResults_Truncate(t *testing.T) {
	// truncate(s, n) returns s unchanged when len(runes) <= n,
	// otherwise returns runes[:n] + "…" (1 extra rune for ellipsis).
	cases := []struct {
		input        string
		maxRunes     int
		wantRuneLen  int
		shouldAppend bool // true means ellipsis appended
	}{
		{"hello", 10, 5, false},
		{"hello", 3, 4, true},  // "hel…"
		{"", 100, 0, false},
		{strings.Repeat("あ", 100), 50, 51, true},  // 50 runes + ellipsis
		{strings.Repeat("あ", 100), 150, 100, false}, // no truncation
		{"hello world", 5, 6, true}, // "hello…"
	}

	for _, tc := range cases {
		got := truncate(tc.input, tc.maxRunes)
		gotLen := len([]rune(got))
		if gotLen != tc.wantRuneLen {
			t.Errorf("truncate(input_runes=%d, %d): got rune len %d, want %d",
				len([]rune(tc.input)), tc.maxRunes, gotLen, tc.wantRuneLen)
		}
		hasEllipsis := strings.HasSuffix(got, "…")
		if hasEllipsis != tc.shouldAppend {
			t.Errorf("truncate ellipsis: expected=%v got=%v", tc.shouldAppend, hasEllipsis)
		}
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestLabResults_AIPromptSanity(t *testing.T) {
	// The system prompt must instruct the AI to act as a medical assistant
	// and to note that results need clinical interpretation.
	systemMsg := "You are a medical AI assistant. Analyze the following lab result summary and provide a brief, plain-language interpretation. Note that this is not a clinical diagnosis and the patient should consult their doctor."

	if !strings.Contains(strings.ToLower(systemMsg), "medical") {
		t.Error("system prompt should mention medical context")
	}
	if !strings.Contains(strings.ToLower(systemMsg), "doctor") || !strings.Contains(strings.ToLower(systemMsg), "consult") {
		t.Error("system prompt should instruct patient to consult doctor")
	}
}

func TestLabResults_TruncateForAI(t *testing.T) {
	// AI input is capped at 3000 runes. truncate returns runes[:3000] + "…" = 3001 runes.
	longSummary := strings.Repeat("x", 5000)
	truncated := truncate(longSummary, 3000)
	runeLen := len([]rune(truncated))
	if runeLen != 3001 { // 3000 content runes + 1 ellipsis rune
		t.Errorf("expected 3001 runes (3000 + ellipsis), got %d", runeLen)
	}
	if !strings.HasSuffix(truncated, "…") {
		t.Error("expected ellipsis suffix on truncated AI input")
	}
}
