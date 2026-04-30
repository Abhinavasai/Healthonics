package handlers

import "testing"

func TestValidateHighRiskGuardrail(t *testing.T) {
	if err := validateHighRiskGuardrail("important clinical reason", "CONFIRM"); err != nil {
		t.Fatalf("expected valid payload, got %v", err)
	}
	if err := validateHighRiskGuardrail("x", "CONFIRM"); err != nil {
		t.Fatalf("expected single-char reason to pass, got %v", err)
	}
	if err := validateHighRiskGuardrail("   ", "CONFIRM"); err == nil {
		t.Fatal("expected empty reason to fail")
	}
	if err := validateHighRiskGuardrail("valid enough reason", "NO"); err == nil {
		t.Fatal("expected invalid confirm to fail")
	}
}

func TestAppendAuditReason(t *testing.T) {
	d := appendAuditReason("deactivated user demo@healthonyx.com", "duplicate account request")
	if d == "deactivated user demo@healthonyx.com" {
		t.Fatal("expected reason to be appended")
	}
}
