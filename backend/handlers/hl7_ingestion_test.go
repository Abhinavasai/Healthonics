package handlers

import "testing"

func TestParseHL7LabMessage_Conformance(t *testing.T) {
	raw := "MSH|^~\\&|LAB|HOSP|HEALTHONYX|APP|202604290100||ORU^R01|MSG-123|P|2.5\r" +
		"PID|1||patient-1^^^HEALTHONYX||Test^Patient\r" +
		"OBR|1||ORDER-9|GLUCOSE^Lab Panel\r" +
		"OBX|1|NM|GLU^Glucose||98|mg/dL|70-110|N|||F\r" +
		"OBX|2|NM|HGB^Hemoglobin||13.8|g/dL|12-16|N|||F\r"
	msg, err := parseHL7LabMessage(raw)
	if err != nil {
		t.Fatalf("expected valid parse, got %v", err)
	}
	if msg.MessageControlID != "MSG-123" {
		t.Fatalf("unexpected message control id: %q", msg.MessageControlID)
	}
	if msg.PatientIdentifier != "patient-1" {
		t.Fatalf("unexpected patient id: %q", msg.PatientIdentifier)
	}
	if len(msg.Observations) != 2 {
		t.Fatalf("expected 2 observations, got %d", len(msg.Observations))
	}
}

func TestParseHL7LabMessage_MalformedResilience(t *testing.T) {
	raw := "PID|1||patient-1^^^HEALTHONYX||Test^Patient\rOBX|1|NM|GLU^Glucose||98|mg/dL"
	_, err := parseHL7LabMessage(raw)
	if err == nil {
		t.Fatal("expected malformed message error")
	}
}

func TestParseHL7LabMessage_ReconciliationPayload(t *testing.T) {
	raw := "MSH|^~\\&|LAB|HOSP|HEALTHONYX|APP|202604290100||ORU^R01|MSG-REC-1|P|2.5\n" +
		"PID|1||550e8400-e29b-41d4-a716-446655440000^^^HEALTHONYX||Test^Patient\n" +
		"OBX|1|ST|COVID^PCR||negative|||N|||F\n"
	msg, err := parseHL7LabMessage(raw)
	if err != nil {
		t.Fatalf("expected parse success, got %v", err)
	}
	if msg.PatientIdentifier != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("expected reconciliable patient identifier, got %q", msg.PatientIdentifier)
	}
	if msg.Observations[0].Code != "COVID" {
		t.Fatalf("unexpected observation code: %q", msg.Observations[0].Code)
	}
}
