package state

import (
	"log"
	"io"
	"path/filepath"
	"testing"
	"time"
)

func newTestUpdater(t *testing.T) *Updater {
	t.Helper()
	path := filepath.Join(t.TempDir(), "state.json")
	store, err := New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	logger := log.New(io.Discard, "", 0)
	return NewUpdater(store, 2*time.Minute, logger)
}

func TestRecordSeen_SetsOkStatus(t *testing.T) {
	u := newTestUpdater(t)
	now := time.Now().UTC()
	if err := u.RecordSeen("backup", now); err != nil {
		t.Fatalf("RecordSeen: %v", err)
	}
	st, ok := u.store.Get("backup")
	if !ok {
		t.Fatal("state not found")
	}
	if st.LastStatus != "ok" {
		t.Errorf("expected ok, got %q", st.LastStatus)
	}
}

func TestRecordFailed_SetsFailedStatus(t *testing.T) {
	u := newTestUpdater(t)
	now := time.Now().UTC()
	if err := u.RecordFailed("cleanup", now); err != nil {
		t.Fatalf("RecordFailed: %v", err)
	}
	st, _ := u.store.Get("cleanup")
	if st.LastStatus != "failed" {
		t.Errorf("expected failed, got %q", st.LastStatus)
	}
}

func TestCheckMissed_NoneWhenRecentlySeen(t *testing.T) {
	u := newTestUpdater(t)
	now := time.Now().UTC()
	_ = u.RecordSeen("heartbeat", now.Add(-30*time.Second))

	events, err := u.CheckMissed("heartbeat", "* * * * *", now)
	if err != nil {
		t.Fatalf("CheckMissed: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 missed events, got %d", len(events))
	}
}

func TestCheckMissed_DetectsMissed(t *testing.T) {
	u := newTestUpdater(t)
	// last seen 3 hours ago; hourly job should have 2 missed runs
	lastSeen := time.Now().UTC().Truncate(time.Hour).Add(-3 * time.Hour)
	_ = u.RecordSeen("report", lastSeen)

	now := lastSeen.Add(3 * time.Hour)
	events, err := u.CheckMissed("report", "0 * * * *", now)
	if err != nil {
		t.Fatalf("CheckMissed: %v", err)
	}
	if len(events) < 2 {
		t.Errorf("expected at least 2 missed events, got %d", len(events))
	}
	for _, e := range events {
		if e.JobName != "report" {
			t.Errorf("unexpected job name %q", e.JobName)
		}
	}
}
