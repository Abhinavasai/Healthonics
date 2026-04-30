package handlers

import (
	"fmt"
	"strings"
)

func validateHighRiskGuardrail(reason, confirm string) error {
	r := strings.TrimSpace(reason)
	if r == "" {
		return fmt.Errorf("reason is required")
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
