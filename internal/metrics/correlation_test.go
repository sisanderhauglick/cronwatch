package metrics

import (
	"testing"
	"time"
)

func newTestCorrelationAnalyzer(snaps []Snapshot) *CorrelationAnalyzer {
	c := &Collector{max: 100}
	for _, s := range snaps {
		c.snapshots = append(c.snapshots, s)
	}
	return NewCorrelationAnalyzer(c, 24*time.Hour)
}

func TestCorrelation_EmptyCollector(t *testing.T) {
	ca := newTestCorrelationAnalyzer(nil)
	results := ca.Analyze(time.Now())
	if len(results) != 0 {
		t.Fatalf("expected no results, got %d", len(results))
	}
}

func TestCorrelation_PerfectPositive(t *testing.T) {
	now := time.Now()
	snaps := []Snapshot{
		{Timestamp: now.Add(-3 * time.Hour), Jobs: map[string]JobStats{
			"alpha": {Failed: 1}, "beta": {Failed: 1},
		}},
		{Timestamp: now.Add(-2 * time.Hour), Jobs: map[string]JobStats{
			"alpha": {Seen: 1}, "beta": {Seen: 1},
		}},
		{Timestamp: now.Add(-1 * time.Hour), Jobs: map[string]JobStats{
			"alpha": {Failed: 1}, "beta": {Failed: 1},
		}},
	}
	ca := newTestCorrelationAnalyzer(snaps)
	results := ca.Analyze(now)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Correlation != 1.0 {
		t.Errorf("expected correlation 1.0, got %f", results[0].Correlation)
	}
}

func TestCorrelation_NoCorrelation(t *testing.T) {
	now := time.Now()
	snaps := []Snapshot{
		{Timestamp: now.Add(-3 * time.Hour), Jobs: map[string]JobStats{
			"alpha": {Failed: 1}, "beta": {Seen: 1},
		}},
		{Timestamp: now.Add(-2 * time.Hour), Jobs: map[string]JobStats{
			"alpha": {Seen: 1}, "beta": {Failed: 1},
		}},
		{Timestamp: now.Add(-1 * time.Hour), Jobs: map[string]JobStats{
			"alpha": {Failed: 1}, "beta": {Seen: 1},
		}},
	}
	ca := newTestCorrelationAnalyzer(snaps)
	results := ca.Analyze(now)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Correlation >= 0 {
		t.Errorf("expected negative correlation, got %f", results[0].Correlation)
	}
}

func TestCorrelation_ExcludesOutsideWindow(t *testing.T) {
	now := time.Now()
	snaps := []Snapshot{
		{Timestamp: now.Add(-48 * time.Hour), Jobs: map[string]JobStats{
			"alpha": {Failed: 1}, "beta": {Failed: 1},
		}},
	}
	ca := newTestCorrelationAnalyzer(snaps)
	results := ca.Analyze(now)
	if len(results) != 0 {
		t.Fatalf("expected no results for old snapshots, got %d", len(results))
	}
}

func TestCorrelation_SingleJobNoResults(t *testing.T) {
	now := time.Now()
	snaps := []Snapshot{
		{Timestamp: now.Add(-1 * time.Hour), Jobs: map[string]JobStats{
			"only": {Failed: 1},
		}},
	}
	ca := newTestCorrelationAnalyzer(snaps)
	results := ca.Analyze(now)
	if len(results) != 0 {
		t.Fatalf("expected no pair results for single job, got %d", len(results))
	}
}
