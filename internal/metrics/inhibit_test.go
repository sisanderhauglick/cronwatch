package metrics

import (
	"testing"
	"time"
)

func newTestInhibitManager(rules []InhibitRule) (*InhibitManager, *Registry) {
	r := New()
	return NewInhibitManager(r, rules, 5*time.Minute), r
}

func TestInhibit_NotInhibitedByDefault(t *testing.T) {
	im, _ := newTestInhibitManager(nil)
	if im.IsInhibited("job-b", time.Now()) {
		t.Fatal("expected not inhibited with no rules")
	}
}

func TestInhibit_InhibitedWhenSourceFailed(t *testing.T) {
	rules := []InhibitRule{{SourceJob: "job-a", TargetJob: "job-b"}}
	im, r := newTestInhibitManager(rules)

	r.RecordFailed("job-a")

	if !im.IsInhibited("job-b", time.Now()) {
		t.Fatal("expected job-b to be inhibited when job-a has failures")
	}
}

func TestInhibit_InhibitedWhenSourceMissed(t *testing.T) {
	rules := []InhibitRule{{SourceJob: "job-a", TargetJob: "job-b"}}
	im, r := newTestInhibitManager(rules)

	r.RecordMissed("job-a")

	if !im.IsInhibited("job-b", time.Now()) {
		t.Fatal("expected job-b to be inhibited when job-a has missed runs")
	}
}

func TestInhibit_NotInhibitedWhenSourceHealthy(t *testing.T) {
	rules := []InhibitRule{{SourceJob: "job-a", TargetJob: "job-b"}}
	im, r := newTestInhibitManager(rules)

	r.RecordSeen("job-a")

	if im.IsInhibited("job-b", time.Now()) {
		t.Fatal("expected job-b not inhibited when job-a is healthy")
	}
}

func TestInhibit_NotInhibitedWhenWindowExpired(t *testing.T) {
	rules := []InhibitRule{{SourceJob: "job-a", TargetJob: "job-b"}}
	im, r := newTestInhibitManager(rules)

	r.RecordFailed("job-a")

	// Evaluate far in the future — outside the 5-minute window.
	future := time.Now().Add(10 * time.Minute)
	if im.IsInhibited("job-b", future) {
		t.Fatal("expected job-b not inhibited when source data is stale")
	}
}

func TestInhibit_AddRuleAtRuntime(t *testing.T) {
	im, r := newTestInhibitManager(nil)
	r.RecordFailed("job-a")

	// Not inhibited before rule is added.
	if im.IsInhibited("job-b", time.Now()) {
		t.Fatal("should not be inhibited before rule added")
	}

	im.AddRule(InhibitRule{SourceJob: "job-a", TargetJob: "job-b"})

	if !im.IsInhibited("job-b", time.Now()) {
		t.Fatal("expected inhibition after rule added")
	}
}

func TestInhibit_RulesSnapshot(t *testing.T) {
	rules := []InhibitRule{
		{SourceJob: "a", TargetJob: "b"},
		{SourceJob: "c", TargetJob: "d"},
	}
	im, _ := newTestInhibitManager(rules)
	snap := im.Rules()
	if len(snap) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(snap))
	}
}
