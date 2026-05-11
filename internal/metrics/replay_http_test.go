package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestReplayHandler() (*ReplayAnalyzer, *AlertTracker, http.HandlerFunc) {
	c := NewCollector(DefaultRetentionPolicy())
	t := NewAlertTracker(64)
	ra := NewReplayAnalyzer(c, t)
	return ra, t, ReplayHandler(ra)
}

func TestReplayHandler_ContentType(t *testing.T) {
	_, _, h := newTestReplayHandler()
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/replay", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestReplayHandler_EmptyReturnsEmptySlice(t *testing.T) {
	_, _, h := newTestReplayHandler()
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/replay", nil))

	var result []ReplayEntry
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(result))
	}
}

func TestReplayHandler_FiltersByJob(t *testing.T) {
	_, tracker, h := newTestReplayHandler()
	now := time.Now()
	tracker.Record(AlertEvent{JobName: "alpha", Kind: "missed", FiredAt: now.Add(-10 * time.Minute), Message: "m"})
	tracker.Record(AlertEvent{JobName: "beta", Kind: "failed", FiredAt: now.Add(-5 * time.Minute), Message: "f"})

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/replay?job=alpha", nil))

	var result []ReplayEntry
	json.NewDecoder(rec.Body).Decode(&result) //nolint:errcheck
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if result[0].JobName != "alpha" {
		t.Errorf("expected alpha, got %s", result[0].JobName)
	}
}

func TestReplayHandler_AllJobsWhenNoFilter(t *testing.T) {
	_, tracker, h := newTestReplayHandler()
	now := time.Now()
	tracker.Record(AlertEvent{JobName: "alpha", Kind: "missed", FiredAt: now.Add(-10 * time.Minute), Message: "m"})
	tracker.Record(AlertEvent{JobName: "beta", Kind: "failed", FiredAt: now.Add(-5 * time.Minute), Message: "f"})

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/replay", nil))

	var result []ReplayEntry
	json.NewDecoder(rec.Body).Decode(&result) //nolint:errcheck
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
}
