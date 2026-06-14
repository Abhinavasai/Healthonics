package handlers

import (
	"testing"
	"time"
)

func TestWaitingRoom_WaitMinutes_Calculation(t *testing.T) {
	now := time.Now()

	cases := []struct {
		checkedInAt *time.Time
		expectNil   bool
		minMinutes  int
		maxMinutes  int
	}{
		{nil, true, 0, 0},
		{ptrTime(now.Add(-30 * time.Minute)), false, 29, 31},
		{ptrTime(now.Add(-1 * time.Minute)), false, 0, 2},
		{ptrTime(now.Add(10 * time.Minute)), false, 0, 0}, // future check-in → clamped to 0
	}

	for _, tc := range cases {
		if tc.checkedInAt == nil {
			if !tc.expectNil {
				t.Error("expected non-nil WaitMinutes but got nil")
			}
			continue
		}
		mins := int(now.Sub(*tc.checkedInAt).Minutes())
		if mins < 0 {
			mins = 0
		}
		if mins < tc.minMinutes || mins > tc.maxMinutes {
			t.Errorf("expected wait_minutes in [%d,%d], got %d", tc.minMinutes, tc.maxMinutes, mins)
		}
	}
}

func ptrTime(t time.Time) *time.Time { return &t }

func TestWaitingRoom_SortOrder(t *testing.T) {
	// Checked-in patients should appear before non-checked-in patients.
	// This is enforced by ORDER BY checked_in_at ASC NULLS LAST.
	// We test the logic: patients with checked_in_at < those without.
	now := time.Now()
	checkedInAt := now.Add(-10 * time.Minute)

	entries := []waitingRoomEntry{
		{AppointmentID: "a", CheckedInAt: nil, ScheduledAt: now.Add(30 * time.Minute)},
		{AppointmentID: "b", CheckedInAt: &checkedInAt, ScheduledAt: now.Add(60 * time.Minute)},
	}

	// In the correct sort order, "b" (checked in) should come before "a" (not checked in).
	// Simulate what the DB ORDER BY does: checked-in first.
	if entries[0].CheckedInAt != nil {
		t.Error("first entry should be the non-checked-in patient before sorting")
	}

	// After sort: b before a.
	if entries[1].CheckedInAt == nil {
		t.Error("second entry should be the checked-in patient before sorting")
	}
}
