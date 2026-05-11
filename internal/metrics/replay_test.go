package metrics

import (
	"testing"
	"time"
)

func newTestReplayAnalyzer() (*ReplayAnalyzer, *AlertTracker) {
	c := NewCollector(DefaultRetentionPolicy())
	t := NewAlertTracker(64)
	return NewReplayAnalyzer(c, t), t
}

func TestReplay_EmptyTracker(t *testing.T) {
	ra, _ := newTestReplayAnalyzer()
	now := time.Now()
	entries := ra.Replay("job1", now.Add(-time.Hour), now)
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestReplay_FiltersByJobName(t *testing.T) {
	ra, tracker := newTestReplayAnalyzer()
	now := time.Now()
	tracker.Record(AlertEvent{JobName: "job1", Kind: "missed", FiredAt: now.Add(-10 * time.Minute), Message: "missed"})
	tracker.Record(AlertEvent{JobName: "job2", Kind: "failed", FiredAt: now.Add(-5 * time.Minute), Message: "failed"})

	entries := ra.Replay("job1", now.Add(-time.Hour), now)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].JobName != "job1" {
		t.Errorf("expected job1, got %s", entries[0].JobName)
	}
}

func TestReplay_FiltersByTimeRange(t *testing.T) {
	ra, tracker := newTestReplayAnalyzer()
	now := time.Now()
	tracker.Record(AlertEvent{JobName: "job1", Kind: "missed", FiredAt: now.Add(-2 * time.Hour), Message: "old"})
	tracker.Record(AlertEvent{JobName: "job1", Kind: "missed", FiredAt: now.Add(-10 * time.Minute), Message: "recent"})

	entries := ra.Replay("job1", now.Add(-time.Hour), now)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Message != "recent" {
		t.Errorf("expected recent, got %s", entries[0].Message)
	}
}

func TestReplay_SortedByTime(t *testing.T) {
	ra, tracker := newTestReplayAnalyzer()
	now := time.Now()
	tracker.Record(AlertEvent{JobName: "job1", Kind: "missed", FiredAt: now.Add(-5 * time.Minute), Message: "second"})
	tracker.Record(AlertEvent{JobName: "job1", Kind: "failed", FiredAt: now.Add(-20 * time.Minute), Message: "first"})

	entries := ra.Replay("job1", now.Add(-time.Hour), now)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if !entries[0].OccuredAt.Before(entries[1].OccuredAt) {
		t.Error("entries not sorted ascending by time")
	}
}

func TestReplayAll_ReturnsAllJobs(t *testing.T) {
	ra, tracker := newTestReplayAnalyzer()
	now := time.Now()
	tracker.Record(AlertEvent{JobName: "job1", Kind: "missed", FiredAt: now.Add(-10 * time.Minute), Message: "m1"})
	tracker.Record(AlertEvent{JobName: "job2", Kind: "failed", FiredAt: now.Add(-8 * time.Minute), Message: "m2"})
	tracker.Record(AlertEvent{JobName: "job3", Kind: "missed", FiredAt: now.Add(-6 * time.Minute), Message: "m3"})

	entries := ra.ReplayAll(now.Add(-time.Hour), now)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
}
