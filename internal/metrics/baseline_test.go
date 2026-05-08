package metrics

import (
	"math"
	"testing"
	"time"
)

func newTestBaselineAnalyzer(window time.Duration) (*BaselineAnalyzer, *Collector) {
	c := NewCollector(DefaultRetentionPolicy(100, 7*24*time.Hour))
	return NewBaselineAnalyzer(c, window), c
}

func TestBaseline_EmptyCollector(t *testing.T) {
	analyzer, _ := newTestBaselineAnalyzer(24 * time.Hour)
	results := analyzer.Analyze(time.Now())
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestBaseline_SingleHourSingleJob(t *testing.T) {
	analyzer, c := newTestBaselineAnalyzer(24 * time.Hour)
	now := time.Now()

	snap := Snapshot{
		Timestamp: now.Add(-30 * time.Minute),
		Jobs: map[string]JobMetrics{
			"backup": {SeenCount: 3},
		},
	}
	c.Collect(snap)

	results := analyzer.Analyze(now)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Job != "backup" {
		t.Errorf("unexpected job name: %s", results[0].Job)
	}
	if results[0].AvgRunsPerHr != 3 {
		t.Errorf("expected avg 3, got %f", results[0].AvgRunsPerHr)
	}
	if results[0].SampleCount != 1 {
		t.Errorf("expected sample count 1, got %d", results[0].SampleCount)
	}
}

func TestBaseline_ExcludesSnapshotsOutsideWindow(t *testing.T) {
	analyzer, c := newTestBaselineAnalyzer(1 * time.Hour)
	now := time.Now()

	old := Snapshot{
		Timestamp: now.Add(-3 * time.Hour),
		Jobs:      map[string]JobMetrics{"sync": {SeenCount: 10}},
	}
	recent := Snapshot{
		Timestamp: now.Add(-10 * time.Minute),
		Jobs:      map[string]JobMetrics{"sync": {SeenCount: 2}},
	}
	c.Collect(old)
	c.Collect(recent)

	results := analyzer.Analyze(now)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].AvgRunsPerHr != 2 {
		t.Errorf("expected avg 2 (only recent), got %f", results[0].AvgRunsPerHr)
	}
}

func TestBaseline_StdDevMultipleHours(t *testing.T) {
	analyzer, c := newTestBaselineAnalyzer(48 * time.Hour)
	now := time.Now()

	// Two snapshots in different hours with different counts.
	hour0 := now.Truncate(time.Hour).Add(-2 * time.Hour)
	hour1 := now.Truncate(time.Hour).Add(-1 * time.Hour)

	c.Collect(Snapshot{Timestamp: hour0.Add(5 * time.Minute), Jobs: map[string]JobMetrics{"report": {SeenCount: 4}}})
	c.Collect(Snapshot{Timestamp: hour1.Add(5 * time.Minute), Jobs: map[string]JobMetrics{"report": {SeenCount: 8}}})

	results := analyzer.Analyze(now)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.AvgRunsPerHr != 6.0 {
		t.Errorf("expected avg 6, got %f", r.AvgRunsPerHr)
	}
	expectedStdDev := math.Sqrt(8.0) // sample std dev of [4,8]
	if math.Abs(r.StdDev-expectedStdDev) > 0.0001 {
		t.Errorf("expected std dev ~%f, got %f", expectedStdDev, r.StdDev)
	}
}
