package metrics

import (
	"testing"
	"time"
)

func makeSnap(job string, missed, failed int, lastSeen time.Time) Snapshot {
	return Snapshot{
		JobName:  job,
		Missed:   missed,
		Failed:   failed,
		LastSeen: lastSeen,
	}
}

func TestEvaluate_OKWhenHealthy(t *testing.T) {
	e := NewHealthEvaluator()
	now := time.Now()
	snap := makeSnap("backup", 0, 0, now.Add(-1*time.Hour))
	h := e.Evaluate(snap, now)
	if h.Status != HealthOK {
		t.Fatalf("expected ok, got %s", h.Status)
	}
}

func TestEvaluate_DownWhenMissed(t *testing.T) {
	e := NewHealthEvaluator()
	now := time.Now()
	snap := makeSnap("backup", 2, 0, now.Add(-1*time.Hour))
	h := e.Evaluate(snap, now)
	if h.Status != HealthDown {
		t.Fatalf("expected down, got %s", h.Status)
	}
}

func TestEvaluate_DegradedWhenFailed(t *testing.T) {
	e := NewHealthEvaluator()
	now := time.Now()
	snap := makeSnap("backup", 0, 1, now.Add(-1*time.Hour))
	h := e.Evaluate(snap, now)
	if h.Status != HealthDegraded {
		t.Fatalf("expected degraded, got %s", h.Status)
	}
}

func TestEvaluate_DegradedWhenStale(t *testing.T) {
	e := NewHealthEvaluator()
	now := time.Now()
	snap := makeSnap("backup", 0, 0, now.Add(-48*time.Hour))
	h := e.Evaluate(snap, now)
	if h.Status != HealthDegraded {
		t.Fatalf("expected degraded, got %s", h.Status)
	}
}

func TestEvaluate_UnknownWhenNoData(t *testing.T) {
	e := NewHealthEvaluator()
	now := time.Now()
	snap := makeSnap("backup", 0, 0, time.Time{})
	h := e.Evaluate(snap, now)
	if h.Status != HealthUnknown {
		t.Fatalf("expected unknown, got %s", h.Status)
	}
}

func TestEvaluate_LastSeenPopulated(t *testing.T) {
	e := NewHealthEvaluator()
	now := time.Now()
	ts := now.Add(-30 * time.Minute)
	snap := makeSnap("sync", 0, 0, ts)
	h := e.Evaluate(snap, now)
	if h.LastSeen == nil {
		t.Fatal("expected LastSeen to be set")
	}
}
