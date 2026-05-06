package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestSummaryHandler(t *testing.T) http.Handler {
	t.Helper()
	reg := New()
	col := NewCollector(reg, DefaultRetentionPolicy())
	agg := NewAggregator(col)
	return SummaryHandler(agg, 5*time.Minute)
}

func TestSummaryHandler_EmptyReturnsEmptySlice(t *testing.T) {
	h := newTestSummaryHandler(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics/summary", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var result []interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d items", len(result))
	}
}

func TestSummaryHandler_ContentType(t *testing.T) {
	h := newTestSummaryHandler(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics/summary", nil)
	h.ServeHTTP(rr, req)

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

func TestSummaryHandler_WithData(t *testing.T) {
	reg := New()
	reg.RecordSeen("backup")
	reg.RecordSeen("backup")
	reg.RecordFailed("backup")

	col := NewCollector(reg, DefaultRetentionPolicy())
	col.Collect()

	agg := NewAggregator(col)
	h := SummaryHandler(agg, 5*time.Minute)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics/summary", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var summaries []JobSummary
	if err := json.Unmarshal(rr.Body.Bytes(), &summaries); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].Job != "backup" {
		t.Errorf("expected job 'backup', got %q", summaries[0].Job)
	}
	if summaries[0].TotalRuns != 2 {
		t.Errorf("expected TotalRuns=2, got %d", summaries[0].TotalRuns)
	}
}
