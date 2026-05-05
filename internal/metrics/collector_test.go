package metrics

import (
	"testing"
	"time"
)

func newTestCollector(window time.Duration) (*Collector, *Registry) {
	reg := New()
	return NewCollector(reg, window), reg
}

func TestCollector_LatestEmpty(t *testing.T) {
	c, _ := newTestCollector(time.Minute)
	_, ok := c.Latest()
	if ok {
		t.Fatal("expected no snapshot on empty collector")
	}
}

func TestCollector_CollectStoresSnapshot(t *testing.T) {
	c, reg := newTestCollector(time.Minute)
	reg.RecordSeen("job1")

	c.Collect()

	snap, ok := c.Latest()
	if !ok {
		t.Fatal("expected a snapshot after Collect")
	}
	if _, found := snap.Snapshot["job1"]; !found {
		t.Error("expected job1 in snapshot")
	}
}

func TestCollector_AllReturnsAll(t *testing.T) {
	c, reg := newTestCollector(time.Minute)
	reg.RecordSeen("job1")

	c.Collect()
	c.Collect()

	all := c.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(all))
	}
}

func TestCollector_PrunesOldSnapshots(t *testing.T) {
	c, reg := newTestCollector(50 * time.Millisecond)
	reg.RecordSeen("job1")

	// Inject an old snapshot directly.
	old := TimedSnapshot{
		At:       time.Now().Add(-200 * time.Millisecond),
		Snapshot: map[string]JobStats{"job1": {}},
	}
	c.mu.Lock()
	c.snapshots = append(c.snapshots, old)
	c.mu.Unlock()

	// Collect triggers prune.
	c.Collect()

	all := c.All()
	for _, s := range all {
		if s.At.Equal(old.At) {
			t.Error("old snapshot should have been pruned")
		}
	}
}

func TestCollector_LatestReflectsNewest(t *testing.T) {
	c, reg := newTestCollector(time.Minute)
	reg.RecordSeen("job1")
	c.Collect()

	reg.RecordMissed("job2")
	c.Collect()

	snap, ok := c.Latest()
	if !ok {
		t.Fatal("expected snapshot")
	}
	if _, found := snap.Snapshot["job2"]; !found {
		t.Error("latest snapshot should contain job2")
	}
}
