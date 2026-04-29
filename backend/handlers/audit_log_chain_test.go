package handlers

import (
	"testing"
	"time"
)

func TestVerifyImmutableAuditChain_Valid(t *testing.T) {
	t0 := time.Date(2026, 4, 28, 23, 0, 0, 0, time.UTC)
	e1 := immutableAuditEvent{
		Action:     "admin_user_created",
		EntityType: "user",
		EntityID:   "u1",
		Detail:     "created user",
		EventTime:  t0,
		PrevHash:   "",
	}
	e1.EventHash = computeImmutableAuditHash(e1.PrevHash, e1.Action, e1.EntityType, e1.EntityID, e1.Detail, e1.EventTime)

	e2 := immutableAuditEvent{
		Action:     "admin_user_deactivated",
		EntityType: "user",
		EntityID:   "u2",
		Detail:     "deactivated user",
		EventTime:  t0.Add(1 * time.Minute),
		PrevHash:   e1.EventHash,
	}
	e2.EventHash = computeImmutableAuditHash(e2.PrevHash, e2.Action, e2.EntityType, e2.EntityID, e2.Detail, e2.EventTime)

	if err := verifyImmutableAuditChain([]immutableAuditEvent{e1, e2}); err != nil {
		t.Fatalf("expected valid chain, got error: %v", err)
	}
}

func TestVerifyImmutableAuditChain_DetectsTamper(t *testing.T) {
	t0 := time.Date(2026, 4, 28, 23, 0, 0, 0, time.UTC)
	e1 := immutableAuditEvent{
		Action:     "prescription_updated",
		EntityType: "prescription",
		EntityID:   "rx1",
		Detail:     "updated dosage",
		EventTime:  t0,
		PrevHash:   "",
	}
	e1.EventHash = computeImmutableAuditHash(e1.PrevHash, e1.Action, e1.EntityType, e1.EntityID, e1.Detail, e1.EventTime)

	e2 := immutableAuditEvent{
		Action:     "prescription_revoked",
		EntityType: "prescription",
		EntityID:   "rx1",
		Detail:     "revoked",
		EventTime:  t0.Add(2 * time.Minute),
		PrevHash:   e1.EventHash,
	}
	e2.EventHash = computeImmutableAuditHash(e2.PrevHash, e2.Action, e2.EntityType, e2.EntityID, e2.Detail, e2.EventTime)

	// Tamper detail after checksum creation.
	e2.Detail = "tampered detail"
	if err := verifyImmutableAuditChain([]immutableAuditEvent{e1, e2}); err == nil {
		t.Fatal("expected tamper detection error")
	}
}

func TestVerifyImmutableAuditChain_DetectsPrevHashBreak(t *testing.T) {
	t0 := time.Date(2026, 4, 28, 23, 0, 0, 0, time.UTC)
	e1 := immutableAuditEvent{
		Action:     "appointment_status_changed",
		EntityType: "appointment",
		EntityID:   "a1",
		Detail:     "approved",
		EventTime:  t0,
		PrevHash:   "",
	}
	e1.EventHash = computeImmutableAuditHash(e1.PrevHash, e1.Action, e1.EntityType, e1.EntityID, e1.Detail, e1.EventTime)

	e2 := immutableAuditEvent{
		Action:     "appointment_status_changed",
		EntityType: "appointment",
		EntityID:   "a1",
		Detail:     "completed",
		EventTime:  t0.Add(3 * time.Minute),
		PrevHash:   "broken_prev_hash",
	}
	e2.EventHash = computeImmutableAuditHash(e2.PrevHash, e2.Action, e2.EntityType, e2.EntityID, e2.Detail, e2.EventTime)

	if err := verifyImmutableAuditChain([]immutableAuditEvent{e1, e2}); err == nil {
		t.Fatal("expected prev_hash break detection")
	}
}
