package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "state.json")
}

func TestNew_CreatesEmptyStore(t *testing.T) {
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, ok := s.Get("nonexistent")
	if ok {
		t.Error("expected no state for unknown job")
	}
}

func TestSetAndGet(t *testing.T) {
	path := tempPath(t)
	s, _ := New(path)

	now := time.Now().UTC().Truncate(time.Second)
	st := JobState{LastSeen: now, LastStatus: "ok"}
	if err := s.Set("backup", st); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected state to exist")
	}
	if !got.LastSeen.Equal(now) {
		t.Errorf("LastSeen: got %v, want %v", got.LastSeen, now)
	}
	if got.LastStatus != "ok" {
		t.Errorf("LastStatus: got %q, want %q", got.LastStatus, "ok")
	}
}

func TestPersistence(t *testing.T) {
	path := tempPath(t)
	s1, _ := New(path)
	now := time.Now().UTC().Truncate(time.Second)
	_ = s1.Set("job1", JobState{LastSeen: now, LastStatus: "failed"})

	s2, err := New(path)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	got, ok := s2.Get("job1")
	if !ok {
		t.Fatal("state not persisted")
	}
	if got.LastStatus != "failed" {
		t.Errorf("expected failed, got %q", got.LastStatus)
	}
}

func TestNew_IgnoresMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "no_such.json")
	_, err := New(path)
	if err != nil {
		t.Errorf("expected no error for missing file, got: %v", err)
	}
}

func TestFlush_AtomicWrite(t *testing.T) {
	path := tempPath(t)
	s, _ := New(path)
	_ = s.Set("x", JobState{LastStatus: "ok"})

	// tmp file should not linger
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Error("tmp file should not exist after flush")
	}
}
