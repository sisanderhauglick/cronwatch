package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestSLAEvaluator(t *testing.T) (*UptimeTracker, *SLAEvaluator) {
	t.Helper()
	col := NewCollector(DefaultRetentionPolicy())
	tracker := NewUptimeTracker(col)
	eval := NewSLAEvaluator(tracker, 24*time.Hour)
	return tracker, eval
}

func TestSLAEvaluator_NoDataFullUptime(t *testing.T) {
	_, eval := newTestSLAEvaluator(t)
	targets := []SLATarget{{JobName: "backup", TargetPct: 99.0}}
	results := eval.Evaluate(targets)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Breaching {
		t.Error("expected no breach when no data (full uptime assumed)")
	}
}

func TestSLAEvaluator_DetectsBreach(t *testing.T) {
	col := NewCollector(DefaultRetentionPolicy())
	tracker := NewUptimeTracker(col)
	now := time.Now()

	// Inject snapshots: 1 missed out of 2 => 50% uptime
	col.Store(Snapshot{Job: "nightly", SeenAt: now.Add(-2 * time.Hour), Missed: 1})
	col.Store(Snapshot{Job: "nightly", SeenAt: now.Add(-1 * time.Hour), Missed: 0})

	eval := NewSLAEvaluator(tracker, 24*time.Hour)
	results := eval.Evaluate([]SLATarget{{JobName: "nightly", TargetPct: 99.0}})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Breaching {
		t.Errorf("expected breach, actual=%.2f target=%.2f", results[0].ActualPct, results[0].TargetPct)
	}
	if results[0].MarginPct >= 0 {
		t.Errorf("expected negative margin, got %.2f", results[0].MarginPct)
	}
}

func TestSLAEvaluator_MultipleTargets(t *testing.T) {
	_, eval := newTestSLAEvaluator(t)
	targets := []SLATarget{
		{JobName: "jobA", TargetPct: 95.0},
		{JobName: "jobB", TargetPct: 99.9},
	}
	results := eval.Evaluate(targets)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestSLAHandler_ContentTypeAndBody(t *testing.T) {
	col := NewCollector(DefaultRetentionPolicy())
	tracker := NewUptimeTracker(col)
	targets := []SLATarget{{JobName: "sync", TargetPct: 99.0}}
	h := NewSLAHandler(tracker, 24*time.Hour, targets)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics/sla", nil)
	h.ServeHTTP(rec, req)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
	var results []SLAResult
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}
