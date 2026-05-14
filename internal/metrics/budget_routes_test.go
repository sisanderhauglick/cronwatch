package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRegisterRoutes_BudgetEndpoint(t *testing.T) {
	reg := New()
	collector := NewCollector(10, DefaultRetentionPolicy(24*time.Hour, 100))
	tracker := NewAlertTracker(100)
	silence := NewSilenceManager()
	runLog := NewRunLog(100)
	sloByJob := map[string]float64{"myjob": 99.5}
	budget := NewBudgetAnalyzer(collector, time.Hour, sloByJob)

	mux := http.NewServeMux()
	RegisterRoutes(mux, reg, collector, tracker, silence, runLog, budget)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics/budget", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestRegisterRoutes_MetricsEndpoint(t *testing.T) {
	reg := New()
	reg.RecordSeen("job1")
	collector := NewCollector(10, DefaultRetentionPolicy(24*time.Hour, 100))
	tracker := NewAlertTracker(100)
	silence := NewSilenceManager()
	runLog := NewRunLog(100)
	budget := NewBudgetAnalyzer(collector, time.Hour, map[string]float64{})

	mux := http.NewServeMux()
	RegisterRoutes(mux, reg, collector, tracker, silence, runLog, budget)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
