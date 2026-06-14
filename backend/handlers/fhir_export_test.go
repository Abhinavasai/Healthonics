package handlers

import "testing"

// ---------------------------------------------------------------------------
// appointmentStatusToFHIR tests
// ---------------------------------------------------------------------------

func TestAppointmentStatusToFHIR_Approved(t *testing.T) {
	if got := appointmentStatusToFHIR("approved"); got != "planned" {
		t.Errorf("expected 'planned', got %q", got)
	}
}

func TestAppointmentStatusToFHIR_Completed(t *testing.T) {
	if got := appointmentStatusToFHIR("completed"); got != "finished" {
		t.Errorf("expected 'finished', got %q", got)
	}
}

func TestAppointmentStatusToFHIR_Cancelled(t *testing.T) {
	if got := appointmentStatusToFHIR("cancelled"); got != "cancelled" {
		t.Errorf("expected 'cancelled', got %q", got)
	}
}

func TestAppointmentStatusToFHIR_Rejected(t *testing.T) {
	if got := appointmentStatusToFHIR("rejected"); got != "cancelled" {
		t.Errorf("expected 'cancelled' for rejected, got %q", got)
	}
}

func TestAppointmentStatusToFHIR_Unknown(t *testing.T) {
	for _, s := range []string{"pending", "", "scheduled", "anything"} {
		if got := appointmentStatusToFHIR(s); got != "unknown" {
			t.Errorf("status=%q: expected 'unknown', got %q", s, got)
		}
	}
}

// ---------------------------------------------------------------------------
// prescriptionStatusToFHIR tests
// ---------------------------------------------------------------------------

func TestPrescriptionStatusToFHIR_Active(t *testing.T) {
	if got := prescriptionStatusToFHIR("active"); got != "active" {
		t.Errorf("expected 'active', got %q", got)
	}
}

func TestPrescriptionStatusToFHIR_Completed(t *testing.T) {
	if got := prescriptionStatusToFHIR("completed"); got != "completed" {
		t.Errorf("expected 'completed', got %q", got)
	}
}

func TestPrescriptionStatusToFHIR_Revoked(t *testing.T) {
	if got := prescriptionStatusToFHIR("revoked"); got != "stopped" {
		t.Errorf("expected 'stopped' for revoked, got %q", got)
	}
}

func TestPrescriptionStatusToFHIR_Unknown(t *testing.T) {
	for _, s := range []string{"", "draft", "pending", "anything"} {
		if got := prescriptionStatusToFHIR(s); got != "unknown" {
			t.Errorf("status=%q: expected 'unknown', got %q", s, got)
		}
	}
}
