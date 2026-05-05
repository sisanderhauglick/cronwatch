package metrics

import (
	"testing"
	"time"
)

var t0 = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func TestRecordSeen_IncrementsCounter(t *testing.T) {
	r := New()
	r.RecordSeen("backup", t0)
	r.RecordSeen("backup", t0.Add(time.Minute))

	s, ok := r.Get("backup")
	if !ok {
		t.Fatal("expected job to exist")
	}
	if s.SeenCount != 2 {
		t.Errorf("SeenCount = %d, want 2", s.SeenCount)
	}
	if !s.LastSeen.Equal(t0.Add(time.Minute)) {
		t.Errorf("LastSeen = %v, want %v", s.LastSeen, t0.Add(time.Minute))
	}
}

func TestRecordMissed_IncrementsCounter(t *testing.T) {
	r := New()
	r.RecordMissed("backup", t0)

	s, _ := r.Get("backup")
	if s.MissedCount != 1 {
		t.Errorf("MissedCount = %d, want 1", s.MissedCount)
	}
	if !s.LastMissed.Equal(t0) {
		t.Errorf("LastMissed = %v, want %v", s.LastMissed, t0)
	}
}

func TestRecordFailed_IncrementsCounter(t *testing.T) {
	r := New()
	r.RecordFailed("deploy", t0)

	s, _ := r.Get("deploy")
	if s.FailedCount != 1 {
		t.Errorf("FailedCount = %d, want 1", s.FailedCount)
	}
}

func TestGet_MissingJob(t *testing.T) {
	r := New()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("expected ok=false for unknown job")
	}
}

func TestSnapshot_MultipleJobs(t *testing.T) {
	r := New()
	r.RecordSeen("alpha", t0)
	r.RecordMissed("beta", t0)
	r.RecordFailed("gamma", t0)

	snap := r.Snapshot()
	if len(snap) != 3 {
		t.Errorf("Snapshot len = %d, want 3", len(snap))
	}
}

func TestSnapshot_IsCopy(t *testing.T) {
	r := New()
	r.RecordSeen("job", t0)
	snap := r.Snapshot()
	snap[0].SeenCount = 999

	s, _ := r.Get("job")
	if s.SeenCount == 999 {
		t.Error("Snapshot should return a copy, not a reference")
	}
}
