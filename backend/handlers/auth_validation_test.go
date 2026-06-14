package handlers

import (
	"testing"
)

// ---------------------------------------------------------------------------
// validEmailFormat tests (covers the improved net/mail-based validator)
// ---------------------------------------------------------------------------

func TestValidEmailFormat_ValidEmails(t *testing.T) {
	valid := []string{
		"user@example.com",
		"user+tag@domain.co.uk",
		"firstname.lastname@company.org",
		"admin@healthonyx.io",
		"test123@sub.domain.com",
	}
	for _, e := range valid {
		if !validEmailFormat(e) {
			t.Errorf("expected %q to be valid", e)
		}
	}
}

func TestValidEmailFormat_InvalidEmails(t *testing.T) {
	invalid := []string{
		"",
		"notanemail",
		"@missinglocal.com",
		"missing@",
		"a@b",           // no TLD dot
		"user @space.com", // space in local
		"two@@at.com",
		"plaintext",
	}
	for _, e := range invalid {
		if validEmailFormat(e) {
			t.Errorf("expected %q to be invalid, but validEmailFormat returned true", e)
		}
	}
}

// ---------------------------------------------------------------------------
// validPassword tests
// ---------------------------------------------------------------------------

func TestValidPassword_TooShort(t *testing.T) {
	cases := []string{"", "a", "ab", "abc", "abcd", "abcde"}
	for _, pw := range cases {
		if validPassword(pw) {
			t.Errorf("expected password %q to be invalid (too short)", pw)
		}
	}
}

func TestValidPassword_Valid(t *testing.T) {
	cases := []string{"abcdef", "Password1!", "verylongpassword12345", "🔑🔑🔑🔑🔑🔑"}
	for _, pw := range cases {
		if !validPassword(pw) {
			t.Errorf("expected password %q to be valid", pw)
		}
	}
}

func TestValidPassword_UnicodeRunes(t *testing.T) {
	// 6 multibyte runes — should be valid (6 runes, not bytes)
	pw := "αβγδεζ"
	if !validPassword(pw) {
		t.Errorf("expected unicode 6-rune password to be valid")
	}
	// 5 runes — invalid
	pw5 := "αβγδε"
	if validPassword(pw5) {
		t.Errorf("expected unicode 5-rune password to be invalid")
	}
}

// ---------------------------------------------------------------------------
// validRole tests
// ---------------------------------------------------------------------------

func TestValidRole(t *testing.T) {
	valid := []string{"patient", "doctor", "admin"}
	for _, r := range valid {
		if !validRole(r) {
			t.Errorf("expected role %q to be valid", r)
		}
	}
	invalid := []string{"", "superuser", "Patient", "DOCTOR", "nurse", "root"}
	for _, r := range invalid {
		if validRole(r) {
			t.Errorf("expected role %q to be invalid", r)
		}
	}
}

// ---------------------------------------------------------------------------
// previewText tests (messaging)
// ---------------------------------------------------------------------------

func TestPreviewText_ShortString(t *testing.T) {
	s := "Hello world"
	out := previewText(s)
	if out != s {
		t.Errorf("expected %q, got %q", s, out)
	}
}

func TestPreviewText_ExactLimit(t *testing.T) {
	s := make([]rune, 160)
	for i := range s {
		s[i] = 'a'
	}
	out := previewText(string(s))
	if len([]rune(out)) != 160 {
		t.Errorf("expected 160 runes, got %d", len([]rune(out)))
	}
}

func TestPreviewText_TruncatesOver160Runes(t *testing.T) {
	s := make([]rune, 300)
	for i := range s {
		s[i] = 'x'
	}
	out := previewText(string(s))
	// 160 runes + "…" suffix
	if len([]rune(out)) != 161 {
		t.Errorf("expected 161 runes (160 + ellipsis), got %d", len([]rune(out)))
	}
}

func TestPreviewText_TrimsWhitespace(t *testing.T) {
	out := previewText("  hello  ")
	if out != "hello" {
		t.Errorf("expected trimmed output, got %q", out)
	}
}

func TestPreviewText_UnicodeHandled(t *testing.T) {
	// 200 CJK runes (each 3 bytes in UTF-8)
	s := make([]rune, 200)
	for i := range s {
		s[i] = '中'
	}
	out := previewText(string(s))
	runes := []rune(out)
	if len(runes) != 161 {
		t.Errorf("expected 161 runes for unicode truncation, got %d", len(runes))
	}
}

func TestPreviewText_Empty(t *testing.T) {
	if previewText("") != "" {
		t.Error("expected empty string for empty input")
	}
	if previewText("   ") != "" {
		t.Error("expected empty string for whitespace-only input")
	}
}
