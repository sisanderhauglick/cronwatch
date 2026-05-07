package metrics

import (
	"testing"
	"time"
)

func newTestAnomalyDetector() (*AnomalyDetector, *Collector, *LatencyTracker) {
	c := NewCollector(DefaultRetentionPolicy(100, 24*time.Hour))
	lt := NewLatencyTracker(time.Hour)
	det := NewAnomalyDetector(c, lt, 30*time.Minute, 2.0)
	return det, c, lt
}

func TestAnomalyDetector_EmptyCollector(t *testing.T) {
	det, _, _ := newTestAnomalyDetector()
	reports := det.Detect(time.Now())
	if len(reports) != 0 {
		t.Fatalf("expected no reports, got %d", len(reports))
	}
}

func TestAnomalyDetector_DetectsFailureBurst(t *testing.T) {
	det, c, _ := newTestAnomalyDetector()
	now := time.Now()

	// Record 3 failed snapshots within the window
	for i := 0; i < 3; i++ {
		snap := Snapshot{
			Job:         "backup",
			Failed:      1,
			CollectedAt: now.Add(-time.Duration(i) * 5 * time.Minute),
		}
		c.store(snap)
	}

	reports := det.Detect(now)
	if len(reports) == 0 {
		t.Fatal("expected failure_burst report")
	}
	found := false
	for _, r := range reports {
		if r.Job == "backup" && r.Kind == "failure_burst" {
			found = true
			if r.Score < 0.5 {
				t.Errorf("expected score >= 0.5, got %f", r.Score)
			}
		}
	}
	if !found {
		t.Error("failure_burst report not found for job 'backup'")
	}
}

func TestAnomalyDetector_NoReportWhenHealthy(t *testing.T) {
	det, c, _ := newTestAnomalyDetector()
	now := time.Now()

	for i := 0; i < 4; i++ {
		snap := Snapshot{
			Job:         "sync",
			Seen:        1,
			Failed:      0,
			CollectedAt: now.Add(-time.Duration(i) * 5 * time.Minute),
		}
		c.store(snap)
	}

	reports := det.Detect(now)
	for _, r := range reports {
		if r.Job == "sync" && r.Kind == "failure_burst" {
			t.Error("unexpected failure_burst for healthy job")
		}
	}
}

func TestAnomalyDetector_DefaultSpikeRatio(t *testing.T) {
	c := NewCollector(DefaultRetentionPolicy(100, 24*time.Hour))
	lt := NewLatencyTracker(time.Hour)
	det := NewAnomalyDetector(c, lt, 30*time.Minute, 0) // 0 should default to 2.0
	if det.spikeRatio != 2.0 {
		t.Errorf("expected default spikeRatio 2.0, got %f", det.spikeRatio)
	}
}

func TestAnomalyDetector_ExcludesOutsideWindow(t *testing.T) {
	det, c, _ := newTestAnomalyDetector()
	now := time.Now()

	// All snapshots are older than the 30-minute window
	for i := 1; i <= 3; i++ {
		snap := Snapshot{
			Job:         "old-job",
			Failed:      1,
			CollectedAt: now.Add(-time.Duration(i) * time.Hour),
		}
		c.store(snap)
	}

	reports := det.Detect(now)
	for _, r := range reports {
		if r.Job == "old-job" {
			t.Error("should not report on snapshots outside window")
		}
	}
}
