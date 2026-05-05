package metrics

import (
	"testing"
	"time"
)

func makeSnapshot(age time.Duration) Snapshot {
	return Snapshot{
		CollectedAt: time.Now().Add(-age),
		Jobs:        map[string]JobStats{},
	}
}

func TestDefaultRetentionPolicy(t *testing.T) {
	p := DefaultRetentionPolicy()
	if p.MaxAge != 24*time.Hour {
		t.Errorf("expected MaxAge 24h, got %v", p.MaxAge)
	}
	if p.MaxSnapshots != 100 {
		t.Errorf("expected MaxSnapshots 100, got %d", p.MaxSnapshots)
	}
}

func TestRetentionPolicy_FiltersOldSnapshots(t *testing.T) {
	p := RetentionPolicy{MaxAge: time.Hour, MaxSnapshots: 0}

	snapshots := []Snapshot{
		makeSnapshot(2 * time.Hour), // too old
		makeSnapshot(30 * time.Minute), // ok
		makeSnapshot(10 * time.Minute), // ok
	}

	result := p.apply(snapshots)
	if len(result) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(result))
	}
}

func TestRetentionPolicy_TrimsToMaxSnapshots(t *testing.T) {
	p := RetentionPolicy{MaxAge: 0, MaxSnapshots: 2}

	snapshots := []Snapshot{
		makeSnapshot(30 * time.Minute),
		makeSnapshot(20 * time.Minute),
		makeSnapshot(10 * time.Minute),
	}

	result := p.apply(snapshots)
	if len(result) != 2 {
		t.Errorf("expected 2 snapshots after trim, got %d", len(result))
	}
	// Should keep the most recent (smallest age).
	if result[0].CollectedAt.Before(snapshots[1].CollectedAt.Add(-time.Second)) {
		t.Error("expected most recent snapshots to be retained")
	}
}

func TestRetentionPolicy_EmptyInput(t *testing.T) {
	p := DefaultRetentionPolicy()
	result := p.apply(nil)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d", len(result))
	}
}

func TestRetentionPolicy_NoLimits(t *testing.T) {
	p := RetentionPolicy{MaxAge: 0, MaxSnapshots: 0}

	snapshots := []Snapshot{
		makeSnapshot(48 * time.Hour),
		makeSnapshot(72 * time.Hour),
	}

	result := p.apply(snapshots)
	if len(result) != 2 {
		t.Errorf("expected all 2 snapshots retained, got %d", len(result))
	}
}
