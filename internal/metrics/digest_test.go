package metrics

import (
	"testing"
	"time"
)

func newTestDigestAnalyzer(t *testing.T) (*DigestAnalyzer, *Collector, *HealthEvaluator) {
	t.Helper()
	c := NewCollector(DefaultRetentionPolicy(100, 24*time.Hour))
	reg := New()
	h := NewHealthEvaluator(c, reg, 2*time.Hour)
	d := NewDigestAnalyzer(c, h, time.Hour)
	return d, c, h
}

func TestDigest_EmptyCollector(t *testing.T) {
	d, _, _ := newTestDigestAnalyzer(t)
	report := d.Summarize()
	if len(report.Entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(report.Entries))
	}
	if report.Window != time.Hour {
		t.Fatalf("expected window 1h, got %v", report.Window)
	}
}

func TestDigest_AggregatesWithinWindow(t *testing.T) {
	d, c, _ := newTestDigestAnalyzer(t)
	now := time.Now()
	d.now = func() time.Time { return now }

	snap := Snapshot{
		Timestamp: now.Add(-30 * time.Minute),
		Jobs: map[string]JobStats{
			"backup": {Seen: 3, Failed: 1, Missed: 0, TotalLatencyMs: 300},
		},
	}
	c.Collect(snap)

	report := d.Summarize()
	if len(report.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(report.Entries))
	}
	e := report.Entries[0]
	if e.JobName != "backup" {
		t.Errorf("unexpected job name: %s", e.JobName)
	}
	if e.TotalRuns != 3 {
		t.Errorf("expected 3 runs, got %d", e.TotalRuns)
	}
	if e.Failures != 1 {
		t.Errorf("expected 1 failure, got %d", e.Failures)
	}
}

func TestDigest_ExcludesSnapshotsOutsideWindow(t *testing.T) {
	d, c, _ := newTestDigestAnalyzer(t)
	now := time.Now()
	d.now = func() time.Time { return now }

	old := Snapshot{
		Timestamp: now.Add(-2 * time.Hour),
		Jobs: map[string]JobStats{
			"cleanup": {Seen: 5, Failed: 2},
		},
	}
	c.Collect(old)

	report := d.Summarize()
	if len(report.Entries) != 0 {
		t.Fatalf("expected 0 entries for out-of-window snapshot, got %d", len(report.Entries))
	}
}

func TestDigest_SuccessRateComputed(t *testing.T) {
	d, c, _ := newTestDigestAnalyzer(t)
	now := time.Now()
	d.now = func() time.Time { return now }

	snap := Snapshot{
		Timestamp: now.Add(-10 * time.Minute),
		Jobs: map[string]JobStats{
			"sync": {Seen: 4, Failed: 1, Missed: 1, TotalLatencyMs: 400},
		},
	}
	c.Collect(snap)

	report := d.Summarize()
	if len(report.Entries) == 0 {
		t.Fatal("expected entries")
	}
	e := report.Entries[0]
	// success = (4-1-1)/4 = 0.5
	if e.SuccessRate < 0.49 || e.SuccessRate > 0.51 {
		t.Errorf("expected success rate ~0.5, got %f", e.SuccessRate)
	}
}

func TestDigest_EntriesSortedByName(t *testing.T) {
	d, c, _ := newTestDigestAnalyzer(t)
	now := time.Now()
	d.now = func() time.Time { return now }

	snap := Snapshot{
		Timestamp: now.Add(-5 * time.Minute),
		Jobs: map[string]JobStats{
			"zebra": {Seen: 1},
			"alpha": {Seen: 2},
			"mango": {Seen: 3},
		},
	}
	c.Collect(snap)

	report := d.Summarize()
	if len(report.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(report.Entries))
	}
	if report.Entries[0].JobName != "alpha" || report.Entries[2].JobName != "zebra" {
		t.Errorf("entries not sorted: %v", report.Entries)
	}
}
