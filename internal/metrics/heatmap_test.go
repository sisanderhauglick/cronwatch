package metrics

import (
	"testing"
	"time"
)

func newTestHeatmapAnalyzer() (*HeatmapAnalyzer, *Collector) {
	c := NewCollector(DefaultRetentionPolicy())
	a := NewHeatmapAnalyzer(c, 7*24*time.Hour)
	return a, c
}

func TestHeatmap_EmptyCollector(t *testing.T) {
	a, _ := newTestHeatmapAnalyzer()
	now := time.Now()
	buckets := a.Analyze("job1", now)
	if len(buckets) != 0 {
		t.Fatalf("expected 0 buckets, got %d", len(buckets))
	}
}

func TestHeatmap_CountsRunsInBucket(t *testing.T) {
	a, c := newTestHeatmapAnalyzer()
	now := time.Date(2024, 6, 10, 14, 0, 0, 0, time.UTC) // Monday 14:00

	for i := 0; i < 3; i++ {
		snap := Snapshot{Timestamp: now.Add(time.Duration(i) * time.Minute), Seen: 1}
		c.Store("job1", snap)
	}

	buckets := a.Analyze("job1", now.Add(time.Hour))
	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(buckets))
	}
	b := buckets[0]
	if b.Total != 3 {
		t.Errorf("expected Total=3, got %d", b.Total)
	}
	if b.Hour != 14 {
		t.Errorf("expected Hour=14, got %d", b.Hour)
	}
	if b.DayOfWeek != time.Monday {
		t.Errorf("expected Monday, got %v", b.DayOfWeek)
	}
}

func TestHeatmap_RecordsFailedAndMissed(t *testing.T) {
	a, c := newTestHeatmapAnalyzer()
	now := time.Date(2024, 6, 11, 9, 0, 0, 0, time.UTC) // Tuesday 09:00

	c.Store("job1", Snapshot{Timestamp: now, Seen: 1, Failed: 1})
	c.Store("job1", Snapshot{Timestamp: now.Add(time.Minute), Seen: 1, Missed: 1})
	c.Store("job1", Snapshot{Timestamp: now.Add(2 * time.Minute), Seen: 1})

	buckets := a.Analyze("job1", now.Add(time.Hour))
	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(buckets))
	}
	b := buckets[0]
	if b.Failed != 1 {
		t.Errorf("expected Failed=1, got %d", b.Failed)
	}
	if b.Missed != 1 {
		t.Errorf("expected Missed=1, got %d", b.Missed)
	}
	if b.FailRate == 0 {
		t.Errorf("expected non-zero FailRate")
	}
}

func TestHeatmap_ExcludesSnapshotsOutsideWindow(t *testing.T) {
	a, c := newTestHeatmapAnalyzer()
	now := time.Now()
	old := now.Add(-8 * 24 * time.Hour) // older than 7-day window

	c.Store("job1", Snapshot{Timestamp: old, Seen: 1, Failed: 1})
	c.Store("job1", Snapshot{Timestamp: now, Seen: 1})

	buckets := a.Analyze("job1", now)
	for _, b := range buckets {
		if b.Failed > 0 {
			t.Errorf("old failed snapshot should be excluded")
		}
	}
}
