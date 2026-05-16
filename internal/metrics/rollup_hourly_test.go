package metrics

import (
	"testing"
	"time"
)

func TestHourlyRollup_ReturnsNBuckets(t *testing.T) {
	a, c := newTestRollupAnalyzer()
	now := time.Now().Truncate(time.Hour)

	// Place one snapshot per hour for the past 3 hours
	for i := 1; i <= 3; i++ {
		ts := now.Add(-time.Duration(i)*time.Hour + 30*time.Minute)
		c.Collect(Snapshot{Job: "sync", Timestamp: ts, Seen: 1})
	}

	buckets := a.HourlyRollup(3, now.Add(time.Hour))
	if len(buckets) != 3 {
		t.Fatalf("expected 3 buckets, got %d", len(buckets))
	}
	for _, b := range buckets {
		if b.Job != "sync" {
			t.Errorf("unexpected job %q", b.Job)
		}
		if b.Window != "1h" {
			t.Errorf("unexpected window %q", b.Window)
		}
		if b.Runs != 1 {
			t.Errorf("expected runs=1, got %d", b.Runs)
		}
	}
}

func TestHourlyRollup_EmptyWhenNoData(t *testing.T) {
	a, _ := newTestRollupAnalyzer()
	now := time.Now()
	buckets := a.HourlyRollup(6, now)
	if len(buckets) != 0 {
		t.Errorf("expected 0 buckets, got %d", len(buckets))
	}
}
