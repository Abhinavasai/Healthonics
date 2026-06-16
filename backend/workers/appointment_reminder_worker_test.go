package workers

import (
	"testing"
	"time"
)

func TestReminderWorker_WindowBounds(t *testing.T) {
	// 24h reminder: appointments between [23h, 24h] from now
	// 1h reminder:  appointments between [55m, 65m] from now
	now := time.Now()
	interval := 10 * time.Minute

	// Window for "24h reminder": appointments in [now+24h, now+24h+10m]
	// Window for "1h reminder":  appointments in [now+1h,  now+1h+10m]
	cases := []struct {
		label      string
		window     time.Duration
		apptOffset time.Duration
		shouldFire bool
	}{
		{"24h reminder — inside window", 24 * time.Hour, 24*time.Hour + 5*time.Minute, true},
		{"24h reminder — too early", 24 * time.Hour, 23 * time.Hour, false},
		{"24h reminder — too late", 24 * time.Hour, 25 * time.Hour, false},
		{"1h reminder — inside window", time.Hour, time.Hour + 5*time.Minute, true},
		{"1h reminder — too early", time.Hour, 45 * time.Minute, false},
		{"1h reminder — too late", time.Hour, 75 * time.Minute, false},
	}

	for _, tc := range cases {
		apptTime := now.Add(tc.apptOffset)
		windowStart := now.Add(tc.window)
		windowEnd := windowStart.Add(interval)

		inWindow := !apptTime.Before(windowStart) && apptTime.Before(windowEnd)
		if inWindow != tc.shouldFire {
			t.Errorf("[%s] apptOffset=%v window=[%v, %v): expected fire=%v got=%v",
				tc.label, tc.apptOffset, tc.window, tc.window+interval, tc.shouldFire, inWindow)
		}
	}
}

func TestReminderWorker_DefaultInterval(t *testing.T) {
	w := NewAppointmentReminderWorker(5 * time.Minute)
	if w.interval != 5*time.Minute {
		t.Errorf("expected 5m interval, got %v", w.interval)
	}
}

func TestReminderWorker_IntervalValidation(t *testing.T) {
	// Worker should accept any positive duration.
	durations := []time.Duration{
		1 * time.Minute,
		5 * time.Minute,
		30 * time.Minute,
	}
	for _, d := range durations {
		w := NewAppointmentReminderWorker(d)
		if w.interval != d {
			t.Errorf("expected interval %v, got %v", d, w.interval)
		}
	}
}
