package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestUptimeTracker(t *testing.T) (*UptimeTracker, *Collector) {
	t.Helper()
	c := NewCollector(DefaultRetentionPolicy(24*time.Hour, 100))
	return NewUptimeTracker(c), c
}

func TestUptimeTracker_NoDataFullUptime(t *testing.T) {
	tracker, _ := newTestUptimeTracker(t)
	now := time.Now()
	tracker.RecordSeen("job-a", now)

	results := tracker.Compute(time.Hour)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].UptimePct != 100.0 {
		t.Errorf("expected 100%% uptime, got %.2f", results[0].UptimePct)
	}
}

func TestUptimeTracker_ComputesPartialUptime(t *testing.T) {
	tracker, c := newTestUptimeTracker(t)
	now := time.Now()
	tracker.RecordSeen("job-b", now.Add(-30*time.Minute))

	snap := Snapshot{
		Timestamp: now.Add(-10 * time.Minute),
		Stats: map[string]JobStats{
			"job-b": {SeenCount: 3, MissedCount: 1},
		},
	}
	c.store(snap)

	results := tracker.Compute(time.Hour)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	got := results[0].UptimePct
	want := 75.0
	if got != want {
		t.Errorf("expected %.2f%% uptime, got %.2f", want, got)
	}
}

func TestUptimeTracker_ExcludesOldSnapshots(t *testing.T) {
	tracker, c := newTestUptimeTracker(t)
	now := time.Now()
	tracker.RecordSeen("job-c", now.Add(-2*time.Hour))

	oldSnap := Snapshot{
		Timestamp: now.Add(-90 * time.Minute),
		Stats: map[string]JobStats{
			"job-c": {SeenCount: 0, MissedCount: 5},
		},
	}
	c.store(oldSnap)

	results := tracker.Compute(time.Hour)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].UptimePct != 100.0 {
		t.Errorf("old snapshots should be excluded, got %.2f", results[0].UptimePct)
	}
}

func TestUptimeHandler_ContentTypeAndBody(t *testing.T) {
	tracker, _ := newTestUptimeTracker(t)
	now := time.Now()
	tracker.RecordSeen("job-d", now)

	handler := UptimeHandler(tracker, time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/uptime", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}

	var out []uptimeResponse
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(out) != 1 || out[0].Job != "job-d" {
		t.Errorf("unexpected response: %+v", out)
	}
}
