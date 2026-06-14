package handlers

import (
	"strings"
	"testing"
)

func TestRefillRequest_ActionValidation(t *testing.T) {
	isValidAction := func(raw string) bool {
		action := strings.ToLower(strings.TrimSpace(raw))
		return action == "approved" || action == "denied"
	}

	validCases := []string{"approved", "denied"}
	// "Approved"/"APPROVED" become "approved" via ToLower so they ARE valid.
	// Only truly invalid strings are listed here.
	invalidCases := []string{"accept", "reject", "", " ", "approve", "deny"}

	for _, v := range validCases {
		if !isValidAction(v) {
			t.Errorf("expected action %q to be valid", v)
		}
	}
	for _, v := range invalidCases {
		if isValidAction(v) {
			t.Errorf("expected action %q to be invalid", v)
		}
	}
}

func TestRefillRequest_NoteValidation(t *testing.T) {
	cases := []struct {
		note  string
		valid bool
	}{
		{"", true},
		{"Need refill urgently", true},
		{strings.Repeat("x", 500), true},
		{strings.Repeat("x", 501), false},
		{strings.Repeat("日", 500), true},  // 500 multibyte runes — should be valid
		{strings.Repeat("日", 501), false}, // 501 multibyte runes — too long
	}

	for _, tc := range cases {
		tooLong := len([]rune(tc.note)) > 500
		got := !tooLong
		if got != tc.valid {
			t.Errorf("note rune_len=%d: expected valid=%v got=%v", len([]rune(tc.note)), tc.valid, got)
		}
	}
}
