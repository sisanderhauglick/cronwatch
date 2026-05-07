package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestLatencyTracker() *LatencyTracker {
	return NewLatencyTracker(1 * time.Hour)
}

func TestLatencyTracker_NoDataReturnsZero(t *testing.T) {
	tr := newTestLatencyTracker()
	p50, p95, p99 := tr.Stats("backup")
	if p50 != 0 || p95 != 0 || p99 != 0 {
		t.Fatalf("expected zeros, got %v %v %v", p50, p95, p99)
	}
}

func TestLatencyTracker_SingleSample(t *testing.T) {
	tr := newTestLatencyTracker()
	tr.Record("backup", 200*time.Millisecond)
	p50, p95, p99 := tr.Stats("backup")
	if p50 != 200*time.Millisecond {
		t.Fatalf("expected p50=200ms, got %v", p50)
	}
	if p95 != 200*time.Millisecond || p99 != 200*time.Millisecond {
		t.Fatalf("expected p95/p99=200ms, got %v %v", p95, p99)
	}
}

func TestLatencyTracker_PercentilesOrdered(t *testing.T) {
	tr := newTestLatencyTracker()
	for i := 1; i <= 100; i++ {
		tr.Record("sync", time.Duration(i)*time.Millisecond)
	}
	p50, p95, p99 := tr.Stats("sync")
	if p50 > p95 || p95 > p99 {
		t.Fatalf("percentiles out of order: p50=%v p95=%v p99=%v", p50, p95, p99)
	}
}

func TestLatencyTracker_PrunesOldRecords(t *testing.T) {
	tr := NewLatencyTracker(10 * time.Millisecond)
	tr.Record("job", 50*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	// recording a new sample triggers pruning
	tr.Record("job", 1*time.Millisecond)
	_, _, p99 := tr.Stats("job")
	if p99 >= 50*time.Millisecond {
		t.Fatalf("expected old record pruned, p99=%v", p99)
	}
}

func TestLatencyHandler_ContentTypeAndBody(t *testing.T) {
	tr := newTestLatencyTracker()
	tr.Record("cleanup", 100*time.Millisecond)
	tr.Record("cleanup", 200*time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/metrics/latency", nil)
	rec := httptest.NewRecorder()
	LatencyHandler(tr)(rec, req)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}

	var results []latencyResponse
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Job != "cleanup" {
		t.Fatalf("expected job=cleanup, got %s", results[0].Job)
	}
	if results[0].P50 <= 0 {
		t.Fatalf("expected positive p50, got %v", results[0].P50)
	}
}
