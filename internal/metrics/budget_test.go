package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestBudgetAnalyzer(slo map[string]float64) (*BudgetAnalyzer, *Collector) {
	c := NewCollector(10, DefaultRetentionPolicy(24*time.Hour, 100))
	return NewBudgetAnalyzer(c, time.Hour, slo), c
}

func TestBudget_EmptyCollector(t *testing.T) {
	a, _ := newTestBudgetAnalyzer(map[string]float64{"backup": 99.9})
	results := a.Analyze(time.Now())
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].AllowedErrors != 0 {
		t.Errorf("expected 0 allowed errors with no runs, got %d", results[0].AllowedErrors)
	}
}

func TestBudget_WithinBudget(t *testing.T) {
	a, c := newTestBudgetAnalyzer(map[string]float64{"backup": 90.0})
	now := time.Now()
	c.Collect(Snapshot{
		Time: now.Add(-10 * time.Minute),
		Jobs: map[string]JobStats{
			"backup": {Seen: 10, Failed: 0, Missed: 0},
		},
	})
	results := a.Analyze(now)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.AllowedErrors != 1 {
		t.Errorf("expected 1 allowed error (10%% of 10), got %d", r.AllowedErrors)
	}
	if r.Exhausted {
		t.Error("budget should not be exhausted")
	}
}

func TestBudget_ExhaustedBudget(t *testing.T) {
	a, c := newTestBudgetAnalyzer(map[string]float64{"backup": 99.0})
	now := time.Now()
	c.Collect(Snapshot{
		Time: now.Add(-5 * time.Minute),
		Jobs: map[string]JobStats{
			"backup": {Seen: 10, Failed: 3, Missed: 0},
		},
	})
	results := a.Analyze(now)
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	r := results[0]
	if !r.Exhausted {
		t.Errorf("expected budget exhausted: allowed=%d actual=%d", r.AllowedErrors, r.ActualErrors)
	}
}

func TestBudget_ExcludesOldSnapshots(t *testing.T) {
	a, c := newTestBudgetAnalyzer(map[string]float64{"job": 95.0})
	now := time.Now()
	c.Collect(Snapshot{
		Time: now.Add(-2 * time.Hour), // outside 1h window
		Jobs: map[string]JobStats{
			"job": {Seen: 100, Failed: 50},
		},
	})
	results := a.Analyze(now)
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	if results[0].ActualErrors != 0 {
		t.Errorf("expected 0 actual errors (old snapshot excluded), got %d", results[0].ActualErrors)
	}
}

func TestBudgetHandler_ContentTypeAndBody(t *testing.T) {
	a, c := newTestBudgetAnalyzer(map[string]float64{"myjob": 99.0})
	now := time.Now()
	c.Collect(Snapshot{
		Time: now.Add(-1 * time.Minute),
		Jobs: map[string]JobStats{"myjob": {Seen: 5, Failed: 1}},
	})
	h := BudgetHandler(a)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/budget", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
	var out []ErrorBudget
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if out[0].Job != "myjob" {
		t.Errorf("unexpected job name: %s", out[0].Job)
	}
}
