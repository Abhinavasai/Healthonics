package workers

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// itoa10 tests
// ---------------------------------------------------------------------------

func TestItoa10_Zero(t *testing.T) {
	if got := itoa10(0); got != "0" {
		t.Errorf("expected '0', got %q", got)
	}
}

func TestItoa10_Positive(t *testing.T) {
	cases := map[int]string{
		1:    "1",
		9:    "9",
		10:   "10",
		42:   "42",
		100:  "100",
		9999: "9999",
	}
	for n, want := range cases {
		if got := itoa10(n); got != want {
			t.Errorf("itoa10(%d): expected %q, got %q", n, want, got)
		}
	}
}

func TestItoa10_Negative(t *testing.T) {
	cases := map[int]string{
		-1:   "-1",
		-42:  "-42",
		-100: "-100",
	}
	for n, want := range cases {
		if got := itoa10(n); got != want {
			t.Errorf("itoa10(%d): expected %q, got %q", n, want, got)
		}
	}
}

// ---------------------------------------------------------------------------
// ftoa tests
// ---------------------------------------------------------------------------

func TestFtoa_ZeroDecimals(t *testing.T) {
	if got := ftoa(3.9, 0); got != "3" {
		t.Errorf("expected '3', got %q", got)
	}
}

func TestFtoa_OneDecimal(t *testing.T) {
	if got := ftoa(5.5, 1); got != "5.5" {
		t.Errorf("expected '5.5', got %q", got)
	}
	if got := ftoa(0.0, 1); got != "0.0" {
		t.Errorf("expected '0.0', got %q", got)
	}
	if got := ftoa(12.3, 1); got != "12.3" {
		t.Errorf("expected '12.3', got %q", got)
	}
}

// ---------------------------------------------------------------------------
// buildAnomalyMessage tests
// ---------------------------------------------------------------------------

func TestBuildAnomalyMessage_NormalUUID(t *testing.T) {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	msg := buildAnomalyMessage(userID, 75, 10.0)
	if !strings.Contains(msg, "550e8400") {
		t.Errorf("expected first 8 chars of UUID in message, got: %q", msg)
	}
	if !strings.Contains(msg, "75") {
		t.Errorf("expected count 75 in message, got: %q", msg)
	}
	if !strings.Contains(msg, "10.0") {
		t.Errorf("expected hourly rate in message, got: %q", msg)
	}
}

func TestBuildAnomalyMessage_ShortUserID_NoPanic(t *testing.T) {
	// Should not panic even if userID is shorter than 8 chars.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("buildAnomalyMessage panicked on short userID: %v", r)
		}
	}()
	msg := buildAnomalyMessage("abc", 10, 2.0)
	if !strings.Contains(msg, "abc") {
		t.Errorf("expected 'abc' in message, got: %q", msg)
	}
}

func TestBuildAnomalyMessage_EmptyUserID_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("buildAnomalyMessage panicked on empty userID: %v", r)
		}
	}()
	_ = buildAnomalyMessage("", 5, 1.5)
}

func TestBuildAnomalyMessage_ZeroRate(t *testing.T) {
	msg := buildAnomalyMessage("abc12345def", 50, 0.0)
	if !strings.Contains(msg, "0.0") {
		t.Errorf("expected '0.0' hourly rate, got: %q", msg)
	}
}

func TestBuildAnomalyMessage_ContainsAuditKeyword(t *testing.T) {
	msg := buildAnomalyMessage("abc12345", 100, 5.0)
	if !strings.Contains(strings.ToLower(msg), "audit") {
		t.Errorf("expected 'audit' keyword in message, got: %q", msg)
	}
}
