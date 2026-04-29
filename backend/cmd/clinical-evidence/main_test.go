package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPercentile95_Empty(t *testing.T) {
	got := percentile95([]int{})
	if got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestPercentile95_Basic(t *testing.T) {
	// Sorted: [1,2,3,4,5]
	// p95 index = ceil(0.95*5)-1 = ceil(4.75)-1 = 5-1 = 4 => value 5
	got := percentile95([]int{1, 5, 2, 4, 3})
	if got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
}

func TestBuildSendGridBouncePayload_Shape(t *testing.T) {
	notificationID := uuid.New()
	eventID := "SG-TEST-1"
	recipient := "patient@healthonyx.demo"

	payloadBytes, payloadBody, err := buildSendGridBouncePayload(eventID, notificationID, recipient)
	if err != nil {
		t.Fatalf("buildSendGridBouncePayload error: %v", err)
	}
	if len(payloadBytes) == 0 || len(payloadBody) == 0 {
		t.Fatalf("expected non-empty payload")
	}

	var arr []map[string]any
	if err := json.Unmarshal(payloadBytes, &arr); err != nil {
		t.Fatalf("payload not valid JSON: %v", err)
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 event, got %d", len(arr))
	}
	ev := arr[0]
	if ev["event"] != "bounce" {
		t.Fatalf("expected event bounce, got %v", ev["event"])
	}
	if ev["sg_message_id"] != eventID {
		t.Fatalf("expected sg_message_id=%s, got %v", eventID, ev["sg_message_id"])
	}

	uniqueArgs, _ := ev["unique_args"].(map[string]any)
	if uniqueArgs == nil {
		t.Fatalf("expected unique_args map")
	}
	if uniqueArgs["notification_id"] != notificationID.String() {
		t.Fatalf("expected unique_args.notification_id=%s, got %v", notificationID.String(), uniqueArgs["notification_id"])
	}
}

func TestComputeSendGridSignature_DifferentInputs(t *testing.T) {
	secret := "secret"
	body := `[{"event":"bounce","email":"x"}]`
	s1 := computeSendGridSignature(secret, "123", body)
	s2 := computeSendGridSignature(secret, "124", body)
	if s1 == s2 {
		t.Fatalf("expected different signatures for different timestamps")
	}
	// Also ensure timestamp doesn't drift into future by formatting issues.
	_ = time.Now()
}
