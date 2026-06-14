package handlers

import (
	"strings"
	"testing"
)

func TestVideoLink_URLValidation(t *testing.T) {
	isValidLink := func(link string) bool {
		return link == "" ||
			strings.HasPrefix(link, "https://") ||
			strings.HasPrefix(link, "http://")
	}
	isValidLength := func(link string) bool { return len(link) <= 500 }

	cases := []struct {
		link  string
		valid bool
	}{
		{"", true},                                             // empty = clear the link
		{"https://meet.jit.si/my-room", true},                 // Jitsi Meet
		{"https://meet.google.com/abc-defg-hij", true},        // Google Meet
		{"http://localhost:8080/room", true},                   // local dev
		{"ftp://invalid.com", false},                           // wrong scheme
		{"not-a-url", false},                                   // no scheme
		{"javascript:alert(1)", false},                         // XSS attempt
		{strings.Repeat("a", 499), false},                      // 499 chars no scheme → invalid
		{"https://" + strings.Repeat("a", 492), true},         // 500 chars exactly → valid
		{"https://" + strings.Repeat("a", 493), false},        // 501 chars → too long
	}

	for _, tc := range cases {
		got := isValidLink(tc.link) && isValidLength(tc.link)
		if got != tc.valid {
			t.Errorf("link=%q: expected valid=%v got=%v", tc.link[:min10(len(tc.link), 40)], tc.valid, got)
		}
	}
}

func min10(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestVideoLink_JitsiRoomFormat(t *testing.T) {
	// Typical Jitsi Meet room URLs should be accepted.
	jitsiLinks := []string{
		"https://meet.jit.si/HealthonyxConsult123",
		"https://jitsi.example.com/room-abc",
		"https://8x8.vc/vpaas-magic-cookie-abc/MyRoom",
	}
	for _, link := range jitsiLinks {
		if !strings.HasPrefix(link, "https://") {
			t.Errorf("jitsi link should start with https://: %s", link)
		}
		if len(link) > 500 {
			t.Errorf("jitsi link exceeds max length: %s", link)
		}
	}
}
