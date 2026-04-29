package failures

import (
	"fmt"
	"strings"
)

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

type Failure struct {
	Code            string
	Severity        Severity
	RemediationHint string
	Message         string
	Subsystem       string
}

func (f Failure) Encode() string {
	msg := strings.TrimSpace(strings.ReplaceAll(f.Message, "|", "/"))
	return fmt.Sprintf(
		"code=%s|severity=%s|subsystem=%s|remediation=%s|message=%s",
		strings.TrimSpace(f.Code),
		strings.TrimSpace(string(f.Severity)),
		strings.TrimSpace(f.Subsystem),
		strings.TrimSpace(strings.ReplaceAll(f.RemediationHint, "|", "/")),
		msg,
	)
}

func ClassifyProviderSend(provider string, err error) Failure {
	p := strings.ToLower(strings.TrimSpace(provider))
	msg := errorMessage(err)
	switch {
	case strings.Contains(msg, "not configured"),
		strings.Contains(msg, "api key is required"),
		strings.Contains(msg, "account credentials are required"),
		strings.Contains(msg, "from address is required"),
		strings.Contains(msg, "from number is required"):
		return Failure{
			Code:            "NTF_PROVIDER_CONFIG_MISSING",
			Severity:        SeverityCritical,
			RemediationHint: "configure provider credentials in environment",
			Message:         msg,
			Subsystem:       "provider",
		}
	case strings.Contains(msg, "recipient is required"), strings.Contains(msg, "body is required"):
		return Failure{
			Code:            "NTF_PROVIDER_PAYLOAD_INVALID",
			Severity:        SeverityWarning,
			RemediationHint: "validate recipient and message payload before queueing",
			Message:         msg,
			Subsystem:       "provider",
		}
	default:
		code := "NTF_PROVIDER_SEND_FAILED"
		if p == "twilio" {
			code = "NTF_TWILIO_SEND_FAILED"
		} else if p == "sendgrid" {
			code = "NTF_SENDGRID_SEND_FAILED"
		}
		return Failure{
			Code:            code,
			Severity:        SeverityWarning,
			RemediationHint: "retry delivery and inspect provider callback logs",
			Message:         msg,
			Subsystem:       "provider",
		}
	}
}

func ClassifySuppression(provider, reason string) Failure {
	return Failure{
		Code:            "NTF_RECIPIENT_SUPPRESSED",
		Severity:        SeverityInfo,
		RemediationHint: "clear suppression only after user consent or provider delist",
		Message:         strings.TrimSpace(provider + ":" + reason),
		Subsystem:       "worker",
	}
}

func ClassifyCallback(provider, event, detail string) Failure {
	evt := strings.ToLower(strings.TrimSpace(event))
	code := "NTF_CALLBACK_PROVIDER_FAILURE"
	sev := SeverityWarning
	hint := "review callback payload and reconcile notification state"
	switch evt {
	case "bounce", "blocked", "dropped", "spamreport", "undelivered", "failed":
		code = "NTF_CALLBACK_HARD_FAILURE"
		sev = SeverityWarning
		hint = "suppress recipient and notify support if recurrent"
	case "deferred":
		code = "NTF_CALLBACK_TRANSIENT_FAILURE"
		sev = SeverityInfo
		hint = "keep retry policy active and monitor budget burn"
	case "delivered", "processed", "sent":
		code = "NTF_CALLBACK_SUCCESS"
		sev = SeverityInfo
		hint = "no remediation required"
	}
	return Failure{
		Code:            code,
		Severity:        sev,
		RemediationHint: hint,
		Message:         strings.TrimSpace(provider + ":" + evt + ":" + detail),
		Subsystem:       "callback",
	}
}

func ClassifyAPI(code, message string) Failure {
	return Failure{
		Code:            strings.TrimSpace(code),
		Severity:        SeverityWarning,
		RemediationHint: "check request contract and authentication context",
		Message:         strings.TrimSpace(message),
		Subsystem:       "api",
	}
}

func errorMessage(err error) string {
	if err == nil {
		return "unknown error"
	}
	return strings.TrimSpace(err.Error())
}
