package metrics

import (
	"testing"
	"time"
)

func newTestAggregator() (*Aggregator, *Collector) {
	reg := New()
	col := NewCollector(reg, DefaultRetentionPolicy(24*time.Hour, 100))
	return NewAggregator(col), col
}

func TestSummarize_EmptyCollector(t *testing.T) {
	agg, _ := newTestAggregator()
	now := time.Now()
	result := agg.Summarize(now.Add(-time.Hour), now)
	if len(result) != 0 {
		t.Fatalf("expected empty summary, got %d entries", len(result))
	}
}

func TestSummarize_AggregatesWithinWindow(t *testing.T) {
	agg, col := newTestAggregator()
	now := time.Now()

	// Inject two snapshots inside the window.
	col.snapshots = []Snapshot{
		{
			Timestamp: now.Add(-30 * time.Minute),
			Jobs: map[string]JobStats{
				"backup": {Seen: 1, Missed: 0, Failed: 0},
			},
		},
		{
			Timestamp: now.Add(-10 * time.Minute),
			Jobs: map[string]JobStats{
				"backup": {Seen: 0, Missed: 1, Failed: 0},
			},
		},
	}

	result := agg.Summarize(now.Add(-time.Hour), now)
	if len(result) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(result))
	}
	s := result[0]
	if s.JobName != "backup" {
		t.Errorf("unexpected job name: %s", s.JobName)
	}
	if s.Seen != 1 || s.Missed != 1 {
		t.Errorf("unexpected counts: seen=%d missed=%d", s.Seen, s.Missed)
	}
	expected := float64(1) / float64(2)
	if s.SuccessRate != expected {
		t.Errorf("expected success rate %.2f, got %.2f", expected, s.SuccessRate)
	}
}

func TestSummarize_ExcludesOutsideWindow(t *testing.T) {
	agg, col := newTestAggregator()
	now := time.Now()

	col.snapshots = []Snapshot{
		{
			Timestamp: now.Add(-3 * time.Hour), // outside window
			Jobs: map[string]JobStats{
				"cleanup": {Seen: 5},
			},
		},
	}

	result := agg.Summarize(now.Add(-time.Hour), now)
	if len(result) != 0 {
		t.Fatalf("expected 0 summaries, got %d", len(result))
	}
}

func TestSummarize_SuccessRateAllFailed(t *testing.T) {
	agg, col := newTestAggregator()
	now := time.Now()

	col.snapshots = []Snapshot{
		{
			Timestamp: now.Add(-5 * time.Minute),
			Jobs: map[string]JobStats{
				"sync": {Seen: 0, Missed: 0, Failed: 3},
			},
		},
	}

	result := agg.Summarize(now.Add(-time.Hour), now)
	if len(result) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(result))
	}
	if result[0].SuccessRate != 0.0 {
		t.Errorf("expected success rate 0.0, got %.2f", result[0].SuccessRate)
	}
}
