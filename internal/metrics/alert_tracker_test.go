package metrics

import (
	"testing"
	"time"
)

func makeEvent(job, reason string) AlertEvent {
	return AlertEvent{
		JobName:   job,
		Reason:    reason,
		SentAt:    time.Now(),
		Notifiers: []string{"log"},
	}
}

func TestAlertTracker_RecordAndRecent(t *testing.T) {
	tr := NewAlertTracker(10)
	tr.Record(makeEvent("jobA", "missed"))
	tr.Record(makeEvent("jobB", "failed"))

	events := tr.Recent(5)
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	// newest first
	if events[0].JobName != "jobB" {
		t.Errorf("expected newest event first, got %s", events[0].JobName)
	}
}

func TestAlertTracker_EvictsOldest(t *testing.T) {
	tr := NewAlertTracker(3)
	for i := 0; i < 5; i++ {
		tr.Record(makeEvent("job", "missed"))
	}
	if len(tr.events) != 3 {
		t.Errorf("expected 3 retained events, got %d", len(tr.events))
	}
}

func TestAlertTracker_RecentCapped(t *testing.T) {
	tr := NewAlertTracker(10)
	tr.Record(makeEvent("jobA", "missed"))
	tr.Record(makeEvent("jobB", "missed"))

	events := tr.Recent(1)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].JobName != "jobB" {
		t.Errorf("expected most recent job, got %s", events[0].JobName)
	}
}

func TestAlertTracker_CountByJob(t *testing.T) {
	tr := NewAlertTracker(20)
	tr.Record(makeEvent("jobA", "missed"))
	tr.Record(makeEvent("jobA", "failed"))
	tr.Record(makeEvent("jobB", "missed"))

	counts := tr.CountByJob()
	if counts["jobA"] != 2 {
		t.Errorf("expected 2 alerts for jobA, got %d", counts["jobA"])
	}
	if counts["jobB"] != 1 {
		t.Errorf("expected 1 alert for jobB, got %d", counts["jobB"])
	}
}

func TestAlertTracker_DefaultSentAt(t *testing.T) {
	tr := NewAlertTracker(10)
	before := time.Now()
	tr.Record(AlertEvent{JobName: "jobA", Reason: "missed"})
	after := time.Now()

	events := tr.Recent(1)
	if events[0].SentAt.Before(before) || events[0].SentAt.After(after) {
		t.Error("SentAt not set to approximately now")
	}
}

func TestAlertTracker_EmptyRecent(t *testing.T) {
	tr := NewAlertTracker(10)
	if got := tr.Recent(5); len(got) != 0 {
		t.Errorf("expected empty slice, got %d events", len(got))
	}
}
