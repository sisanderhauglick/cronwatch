package metrics

import (
	"testing"
	"time"
)

func newTestDependencyAnalyzer(t *testing.T) (*DependencyAnalyzer, *Collector) {
	t.Helper()
	c := NewCollector(100, DefaultRetentionPolicy(24*time.Hour, 1000))
	return NewDependencyAnalyzer(c, time.Hour, 0.6), c
}

func TestDependency_EmptyCollector(t *testing.T) {
	da, _ := newTestDependencyAnalyzer(t)
	edges := da.Analyze(time.Now())
	if len(edges) != 0 {
		t.Fatalf("expected no edges, got %d", len(edges))
	}
}

func TestDependency_DetectsCorrelatedFailures(t *testing.T) {
	da, c := newTestDependencyAnalyzer(t)
	now := time.Now().Truncate(time.Minute)

	// jobA and jobB always fail together across 6 distinct minutes
	for i := 0; i < 6; i++ {
		ts := now.Add(time.Duration(i) * time.Minute)
		c.Collect(Snapshot{
			Time: ts,
			Jobs: map[string]JobMetrics{
				"jobA": {Failed: 1},
				"jobB": {Failed: 1},
			},
		})
	}

	edges := da.Analyze(now.Add(10 * time.Minute))
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(edges))
	}
	if edges[0].From != "jobA" || edges[0].To != "jobB" {
		t.Errorf("unexpected edge: %+v", edges[0])
	}
	if edges[0].Correlation < 0.6 {
		t.Errorf("correlation too low: %f", edges[0].Correlation)
	}
}

func TestDependency_NoEdgeWhenUncorrelated(t *testing.T) {
	da, c := newTestDependencyAnalyzer(t)
	now := time.Now().Truncate(time.Minute)

	// jobA fails on even minutes, jobB on odd — anti-correlated / uncorrelated
	for i := 0; i < 8; i++ {
		ts := now.Add(time.Duration(i) * time.Minute)
		snap := Snapshot{Time: ts, Jobs: map[string]JobMetrics{}}
		if i%2 == 0 {
			snap.Jobs["jobA"] = JobMetrics{Failed: 1}
			snap.Jobs["jobB"] = JobMetrics{Seen: 1}
		} else {
			snap.Jobs["jobA"] = JobMetrics{Seen: 1}
			snap.Jobs["jobB"] = JobMetrics{Failed: 1}
		}
		c.Collect(snap)
	}

	edges := da.Analyze(now.Add(10 * time.Minute))
	if len(edges) != 0 {
		t.Fatalf("expected no edges for uncorrelated jobs, got %d", len(edges))
	}
}

func TestDependency_ExcludesSnapshotsOutsideWindow(t *testing.T) {
	da, c := newTestDependencyAnalyzer(t)
	now := time.Now().Truncate(time.Minute)

	// all snapshots are 2 hours old — outside the 1h window
	for i := 0; i < 6; i++ {
		ts := now.Add(-2*time.Hour + time.Duration(i)*time.Minute)
		c.Collect(Snapshot{
			Time: ts,
			Jobs: map[string]JobMetrics{
				"jobA": {Failed: 1},
				"jobB": {Failed: 1},
			},
		})
	}

	edges := da.Analyze(now)
	if len(edges) != 0 {
		t.Fatalf("expected no edges outside window, got %d", len(edges))
	}
}

func TestDependency_DefaultThreshold(t *testing.T) {
	c := NewCollector(100, DefaultRetentionPolicy(24*time.Hour, 1000))
	da := NewDependencyAnalyzer(c, time.Hour, 0) // 0 should default to 0.6
	if da.threshold != 0.6 {
		t.Errorf("expected default threshold 0.6, got %f", da.threshold)
	}
}
