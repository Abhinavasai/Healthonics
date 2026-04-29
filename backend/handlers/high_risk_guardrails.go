package handlers

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func validateHighRiskGuardrail(reason, confirm string) error {
	r := strings.TrimSpace(reason)
	if utf8.RuneCountInString(r) < 8 {
		return fmt.Errorf("reason must be at least 8 characters")
	}
	if strings.ToUpper(strings.TrimSpace(confirm)) != "CONFIRM" {
		return fmt.Errorf("confirm must equal CONFIRM")
	}
	return nil
}

func appendAuditReason(detail, reason string) string {
	r := strings.TrimSpace(reason)
	if r == "" {
		return strings.TrimSpace(detail)
	}
	return fmt.Sprintf("%s | reason=%s", strings.TrimSpace(detail), r)
}
