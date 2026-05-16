package metrics

import (
	"testing"
	"time"
)

func newTestStaleDetector(threshold time.Duration) (*StaleDetector, *Collector) {
	c := NewCollector(DefaultRetentionPolicy())
	return NewStaleDetector(c, threshold), c
}

func TestStaleDetector_EmptyCollector(t *testing.T) {
	sd, _ := newTestStaleDetector(5 * time.Minute)
	entries := sd.Detect(time.Now())
	if len(entries) != 0 {
		t.Fatalf("expected no entries, got %d", len(entries))
	}
}

func TestStaleDetector_FreshJobNotStale(t *testing.T) {
	sd, c := newTestStaleDetector(5 * time.Minute)
	now := time.Now()

	snap := Snapshot{Job: "backup", Timestamp: now.Add(-1 * time.Minute), Seen: 1}
	c.Collect(snap)

	entries := sd.Detect(now)
	if len(entries) != 0 {
		t.Fatalf("expected no stale entries, got %d", len(entries))
	}
}

func TestStaleDetector_StaleJobDetected(t *testing.T) {
	sd, c := newTestStaleDetector(5 * time.Minute)
	now := time.Now()

	snap := Snapshot{Job: "cleanup", Timestamp: now.Add(-10 * time.Minute), Seen: 3}
	c.Collect(snap)

	entries := sd.Detect(now)
	if len(entries) != 1 {
		t.Fatalf("expected 1 stale entry, got %d", len(entries))
	}
	if entries[0].Job != "cleanup" {
		t.Errorf("expected job 'cleanup', got %q", entries[0].Job)
	}
	if entries[0].Staleness <= 5*time.Minute {
		t.Errorf("expected staleness > 5m, got %v", entries[0].Staleness)
	}
}

func TestStaleDetector_MultipleJobs(t *testing.T) {
	sd, c := newTestStaleDetector(5 * time.Minute)
	now := time.Now()

	c.Collect(Snapshot{Job: "fresh", Timestamp: now.Add(-2 * time.Minute), Seen: 1})
	c.Collect(Snapshot{Job: "stale1", Timestamp: now.Add(-6 * time.Minute), Seen: 1})
	c.Collect(Snapshot{Job: "stale2", Timestamp: now.Add(-20 * time.Minute), Seen: 1})

	entries := sd.Detect(now)
	if len(entries) != 2 {
		t.Fatalf("expected 2 stale entries, got %d", len(entries))
	}
}

func TestStaleDetector_SetThreshold(t *testing.T) {
	sd, c := newTestStaleDetector(5 * time.Minute)
	now := time.Now()

	c.Collect(Snapshot{Job: "marginal", Timestamp: now.Add(-3 * time.Minute), Seen: 1})

	// Not stale at 5m threshold
	if len(sd.Detect(now)) != 0 {
		t.Fatal("expected no stale entries at 5m threshold")
	}

	// Lower threshold — now it should be stale
	sd.SetThreshold(2 * time.Minute)
	if len(sd.Detect(now)) != 1 {
		t.Fatal("expected 1 stale entry after lowering threshold")
	}
}
