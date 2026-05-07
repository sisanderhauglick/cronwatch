package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestAnomalyHandler() (*AnomalyDetector, http.HandlerFunc) {
	c := NewCollector(DefaultRetentionPolicy(100, 24*time.Hour))
	lt := NewLatencyTracker(time.Hour)
	det := NewAnomalyDetector(c, lt, 30*time.Minute, 2.0)
	return det, AnomalyHandler(det)
}

func TestAnomalyHandler_ContentType(t *testing.T) {
	_, h := newTestAnomalyHandler()
	rr := httptest.NewRecorder()
	h(rr, httptest.NewRequest(http.MethodGet, "/metrics/anomalies", nil))

	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestAnomalyHandler_EmptyReturnsEmptySlice(t *testing.T) {
	_, h := newTestAnomalyHandler()
	rr := httptest.NewRecorder()
	h(rr, httptest.NewRequest(http.MethodGet, "/metrics/anomalies", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var result []anomalyResponse
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d items", len(result))
	}
}

func TestAnomalyHandler_ReturnsFailureBurst(t *testing.T) {
	det, h := newTestAnomalyHandler()
	now := time.Now()

	for i := 0; i < 3; i++ {
		det.collector.store(Snapshot{
			Job:         "nightly",
			Failed:      1,
			CollectedAt: now.Add(-time.Duration(i) * 5 * time.Minute),
		})
	}

	rr := httptest.NewRecorder()
	h(rr, httptest.NewRequest(http.MethodGet, "/metrics/anomalies", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var result []anomalyResponse
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one anomaly report")
	}
	if result[0].Job != "nightly" {
		t.Errorf("expected job 'nightly', got '%s'", result[0].Job)
	}
	if result[0].Kind != "failure_burst" {
		t.Errorf("expected kind 'failure_burst', got '%s'", result[0].Kind)
	}
	if result[0].Score <= 0 {
		t.Errorf("expected positive score, got %f", result[0].Score)
	}
}
