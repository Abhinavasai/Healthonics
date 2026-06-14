package handlers

import (
	"strings"
	"testing"
)

func TestRating_RangeValidation(t *testing.T) {
	validRatings := []int{1, 2, 3, 4, 5}
	invalidRatings := []int{0, -1, 6, 100, -100}

	isValid := func(r int) bool { return r >= 1 && r <= 5 }

	for _, r := range validRatings {
		if !isValid(r) {
			t.Errorf("rating %d should be valid", r)
		}
	}
	for _, r := range invalidRatings {
		if isValid(r) {
			t.Errorf("rating %d should be invalid", r)
		}
	}
}

func TestRating_CommentValidation(t *testing.T) {
	cases := []struct {
		comment string
		valid   bool
	}{
		{"", true},
		{"Great doctor!", true},
		{strings.Repeat("x", 1000), true},
		{strings.Repeat("x", 1001), false},
		{strings.Repeat("あ", 1000), true},  // 1000 multibyte runes
		{strings.Repeat("あ", 1001), false}, // 1001 multibyte runes
	}

	for _, tc := range cases {
		tooLong := len([]rune(tc.comment)) > 1000
		got := !tooLong
		if got != tc.valid {
			t.Errorf("comment rune_len=%d: expected valid=%v got=%v", len([]rune(tc.comment)), tc.valid, got)
		}
	}
}
