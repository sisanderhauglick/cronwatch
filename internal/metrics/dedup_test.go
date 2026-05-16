package metrics

import (
	"testing"
	"time"
)

func newTestDedupManager() *DedupManager {
	return NewDedupManager(5 * time.Minute)
}

func TestDedup_FirstAlertNotDuplicate(t *testing.T) {
	d := newTestDedupManager()
	now := time.Now()
	if d.IsDuplicate("backup", "missed", now) {
		t.Fatal("expected first alert to not be a duplicate")
	}
}

func TestDedup_SecondAlertWithinWindowIsDuplicate(t *testing.T) {
	d := newTestDedupManager()
	now := time.Now()
	d.IsDuplicate("backup", "missed", now)
	if !d.IsDuplicate("backup", "missed", now.Add(1*time.Minute)) {
		t.Fatal("expected second alert within window to be duplicate")
	}
}

func TestDedup_AlertAfterWindowNotDuplicate(t *testing.T) {
	d := newTestDedupManager()
	now := time.Now()
	d.IsDuplicate("backup", "missed", now)
	if d.IsDuplicate("backup", "missed", now.Add(10*time.Minute)) {
		t.Fatal("expected alert after window to not be duplicate")
	}
}

func TestDedup_DifferentReasonNotDuplicate(t *testing.T) {
	d := newTestDedupManager()
	now := time.Now()
	d.IsDuplicate("backup", "missed", now)
	if d.IsDuplicate("backup", "failed", now.Add(1*time.Minute)) {
		t.Fatal("expected different reason to not be duplicate")
	}
}

func TestDedup_ResetAllowsNextAlert(t *testing.T) {
	d := newTestDedupManager()
	now := time.Now()
	d.IsDuplicate("backup", "missed", now)
	d.Reset("backup", "missed")
	if d.IsDuplicate("backup", "missed", now.Add(1*time.Minute)) {
		t.Fatal("expected alert after reset to not be duplicate")
	}
}

func TestDedup_PruneRemovesOldEntries(t *testing.T) {
	d := newTestDedupManager()
	now := time.Now()
	d.IsDuplicate("backup", "missed", now.Add(-10*time.Minute))
	d.Prune(now)
	if len(d.Entries()) != 0 {
		t.Fatalf("expected 0 entries after prune, got %d", len(d.Entries()))
	}
}

func TestDedup_PruneKeepsRecentEntries(t *testing.T) {
	d := newTestDedupManager()
	now := time.Now()
	d.IsDuplicate("backup", "missed", now.Add(-1*time.Minute))
	d.Prune(now)
	if len(d.Entries()) != 1 {
		t.Fatalf("expected 1 entry after prune, got %d", len(d.Entries()))
	}
}
