package failures

import (
	"errors"
	"strings"
	"testing"
)

func TestFailureEncodeSchema(t *testing.T) {
	f := Failure{
		Code:            "NTF_PROVIDER_SEND_FAILED",
		Severity:        SeverityWarning,
		RemediationHint: "retry",
		Message:         "provider timeout",
		Subsystem:       "provider",
	}
	encoded := f.Encode()
	required := []string{"code=", "severity=", "subsystem=", "remediation=", "message="}
	for _, k := range required {
		if !strings.Contains(encoded, k) {
			t.Fatalf("expected encoded failure to contain %s: %s", k, encoded)
		}
	}
}

func TestClassifyProviderSend(t *testing.T) {
	f := ClassifyProviderSend("sendgrid", errors.New("sendgrid: api key is required"))
	if f.Code != "NTF_PROVIDER_CONFIG_MISSING" {
		t.Fatalf("unexpected code: %s", f.Code)
	}
}

func TestIncidentDrillReplayMappings(t *testing.T) {
	cases := []struct {
		provider string
		event    string
		wantCode string
	}{
		{"sendgrid", "bounce", "NTF_CALLBACK_HARD_FAILURE"},
		{"sendgrid", "deferred", "NTF_CALLBACK_TRANSIENT_FAILURE"},
		{"twilio", "delivered", "NTF_CALLBACK_SUCCESS"},
	}
	for _, c := range cases {
		got := ClassifyCallback(c.provider, c.event, "drill")
		if got.Code != c.wantCode {
			t.Fatalf("provider=%s event=%s: got %s want %s", c.provider, c.event, got.Code, c.wantCode)
		}
	}
}
