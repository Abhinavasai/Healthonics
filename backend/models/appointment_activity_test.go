package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAppointmentActivity_JSONFields(t *testing.T) {
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	apptID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	actor := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	ts := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)

	a := AppointmentActivity{
		ID:            id,
		AppointmentID: apptID,
		ActorUserID:   actor,
		ActorEmail:    "doc@example.com",
		Action:        "status_changed",
		Detail:        "approved",
		CreatedAt:     ts,
	}
	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out["action"] != "status_changed" {
		t.Fatalf("action: got %v", out["action"])
	}
	if out["detail"] != "approved" {
		t.Fatalf("detail: got %v", out["detail"])
	}
	if out["actor_email"] != "doc@example.com" {
		t.Fatalf("actor_email: got %v", out["actor_email"])
	}
}
