package metrics

import (
	"testing"
	"time"
)

func newTestCooldownManager() *CooldownManager {
	cm := NewCooldownManager()
	cm.now = func() time.Time { return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC) }
	return cm
}

func TestCooldown_NotActiveByDefault(t *testing.T) {
	cm := newTestCooldownManager()
	if cm.InCooldown("backup") {
		t.Fatal("expected not in cooldown")
	}
}

func TestCooldown_ActiveAfterActivate(t *testing.T) {
	cm := newTestCooldownManager()
	cm.SetCooldown("backup", 5*time.Minute)
	cm.Activate("backup")
	if !cm.InCooldown("backup") {
		t.Fatal("expected in cooldown")
	}
}

func TestCooldown_ExpiresAfterDuration(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	cm := NewCooldownManager()
	cm.now = func() time.Time { return base }
	cm.SetCooldown("backup", 5*time.Minute)
	cm.Activate("backup")

	// Advance time past cooldown
	cm.now = func() time.Time { return base.Add(6 * time.Minute) }
	if cm.InCooldown("backup") {
		t.Fatal("expected cooldown to have expired")
	}
}

func TestCooldown_ResetClearsCooldown(t *testing.T) {
	cm := newTestCooldownManager()
	cm.SetCooldown("backup", 10*time.Minute)
	cm.Activate("backup")
	cm.Reset("backup")
	if cm.InCooldown("backup") {
		t.Fatal("expected cooldown cleared after reset")
	}
}

func TestCooldown_AllReturnsEntries(t *testing.T) {
	cm := newTestCooldownManager()
	cm.SetCooldown("jobA", time.Minute)
	cm.SetCooldown("jobB", 2*time.Minute)
	cm.Activate("jobA")

	entries := cm.All()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestCooldown_ActivateWithoutSetCooldown(t *testing.T) {
	cm := newTestCooldownManager()
	// Activate without setting a duration — cooldown is zero, so InCooldown returns false immediately
	cm.Activate("orphan")
	if cm.InCooldown("orphan") {
		t.Fatal("expected not in cooldown when duration is zero")
	}
}
