package metrics

import (
	"testing"
	"time"
)

func newTestWatchdogManager() *WatchdogManager {
	w := NewWatchdogManager()
	w.now = func() time.Time { return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC) }
	return w
}

func TestWatchdog_NoExpiredWhenNoBeat(t *testing.T) {
	w := newTestWatchdogManager()
	w.Register("job1", 5*time.Minute)
	// LastBeat is zero — should not appear in Expired
	if got := w.Expired(); len(got) != 0 {
		t.Fatalf("expected 0 expired, got %d", len(got))
	}
}

func TestWatchdog_NotExpiredWhenFresh(t *testing.T) {
	w := newTestWatchdogManager()
	w.Register("job1", 5*time.Minute)
	w.Beat("job1")
	// beat happened at now; timeout not elapsed
	if got := w.Expired(); len(got) != 0 {
		t.Fatalf("expected 0 expired, got %d", len(got))
	}
}

func TestWatchdog_ExpiredAfterTimeout(t *testing.T) {
	w := newTestWatchdogManager()
	w.Register("job1", 5*time.Minute)
	w.Beat("job1")
	// advance clock past timeout
	w.now = func() time.Time { return time.Date(2024, 1, 1, 12, 10, 0, 0, time.UTC) }
	got := w.Expired()
	if len(got) != 1 {
		t.Fatalf("expected 1 expired, got %d", len(got))
	}
	if got[0].Job != "job1" {
		t.Errorf("unexpected job name: %s", got[0].Job)
	}
}

func TestWatchdog_TriggeredOnlyOnce(t *testing.T) {
	w := newTestWatchdogManager()
	w.Register("job1", 5*time.Minute)
	w.Beat("job1")
	w.now = func() time.Time { return time.Date(2024, 1, 1, 12, 10, 0, 0, time.UTC) }
	w.Expired() // first call marks triggered
	if got := w.Expired(); len(got) != 0 {
		t.Fatalf("expected 0 on second call, got %d", len(got))
	}
}

func TestWatchdog_BeatResetsTriggered(t *testing.T) {
	w := newTestWatchdogManager()
	w.Register("job1", 5*time.Minute)
	w.Beat("job1")
	w.now = func() time.Time { return time.Date(2024, 1, 1, 12, 10, 0, 0, time.UTC) }
	w.Expired()
	w.Beat("job1") // reset
	// still past old beat but new beat is at 12:10; advance 10 more min
	w.now = func() time.Time { return time.Date(2024, 1, 1, 12, 20, 0, 0, time.UTC) }
	got := w.Expired()
	if len(got) != 1 {
		t.Fatalf("expected 1 expired after reset, got %d", len(got))
	}
}

func TestWatchdog_AllReturnsRegistered(t *testing.T) {
	w := newTestWatchdogManager()
	w.Register("jobA", time.Minute)
	w.Register("jobB", time.Minute)
	if got := w.All(); len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
}
