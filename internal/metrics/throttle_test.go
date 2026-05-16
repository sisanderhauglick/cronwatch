package metrics

import (
	"testing"
	"time"
)

func newTestThrottleManager(cooldown time.Duration) *ThrottleManager {
	tm := NewThrottleManager(cooldown)
	fixed := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	tm.now = func() time.Time { return fixed }
	return tm
}

func TestThrottle_FirstAlwaysAllowed(t *testing.T) {
	tm := newTestThrottleManager(5 * time.Minute)
	if !tm.Allow("job-a") {
		t.Fatal("expected first alert to be allowed")
	}
}

func TestThrottle_SuppressedWithinCooldown(t *testing.T) {
	tm := newTestThrottleManager(5 * time.Minute)
	tm.Allow("job-a")
	if tm.Allow("job-a") {
		t.Fatal("expected second alert within cooldown to be suppressed")
	}
	e, ok := tm.Stats("job-a")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Suppressed != 1 {
		t.Fatalf("expected suppressed=1, got %d", e.Suppressed)
	}
}

func TestThrottle_AllowedAfterCooldown(t *testing.T) {
	tm := newTestThrottleManager(5 * time.Minute)
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	tm.now = func() time.Time { return base }
	tm.Allow("job-a")

	tm.now = func() time.Time { return base.Add(6 * time.Minute) }
	if !tm.Allow("job-a") {
		t.Fatal("expected alert to be allowed after cooldown")
	}
}

func TestThrottle_ResetClearsEntry(t *testing.T) {
	tm := newTestThrottleManager(5 * time.Minute)
	tm.Allow("job-a")
	tm.Reset("job-a")
	_, ok := tm.Stats("job-a")
	if ok {
		t.Fatal("expected entry to be removed after reset")
	}
}

func TestThrottle_AllReturnsAllEntries(t *testing.T) {
	tm := newTestThrottleManager(5 * time.Minute)
	tm.Allow("job-a")
	tm.Allow("job-b")
	all := tm.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestThrottle_SuppressedCountAccumulates(t *testing.T) {
	tm := newTestThrottleManager(10 * time.Minute)
	tm.Allow("job-x") // allowed
	tm.Allow("job-x") // suppressed 1
	tm.Allow("job-x") // suppressed 2
	e, _ := tm.Stats("job-x")
	if e.Suppressed != 2 {
		t.Fatalf("expected suppressed=2, got %d", e.Suppressed)
	}
}
