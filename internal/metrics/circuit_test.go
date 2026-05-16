package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestCircuitBreaker() *CircuitBreaker {
	cb := NewCircuitBreaker(3, 5*time.Minute)
	return cb
}

func TestCircuit_ClosedByDefault(t *testing.T) {
	cb := newTestCircuitBreaker()
	if cb.IsOpen("job1") {
		t.Fatal("expected circuit to be closed by default")
	}
}

func TestCircuit_OpensAfterThreshold(t *testing.T) {
	cb := newTestCircuitBreaker()
	cb.RecordFailure("job1")
	cb.RecordFailure("job1")
	if cb.IsOpen("job1") {
		t.Fatal("should not be open before threshold")
	}
	cb.RecordFailure("job1")
	if !cb.IsOpen("job1") {
		t.Fatal("expected circuit to be open after threshold")
	}
}

func TestCircuit_ClosedAfterSuccess(t *testing.T) {
	cb := newTestCircuitBreaker()
	for i := 0; i < 3; i++ {
		cb.RecordFailure("job1")
	}
	cb.RecordSuccess("job1")
	if cb.IsOpen("job1") {
		t.Fatal("expected circuit to be closed after success")
	}
}

func TestCircuit_HalfOpenAfterReset(t *testing.T) {
	cb := NewCircuitBreaker(1, 10*time.Millisecond)
	past := time.Now().Add(-1 * time.Second)
	cb.RecordFailure("job1")
	// manually backdate the opened time
	cb.mu.Lock()
	cb.entries["job1"].OpenedAt = past
	cb.mu.Unlock()
	// IsOpen should transition to half_open
	if cb.IsOpen("job1") {
		t.Fatal("expected circuit to be half-open (not open) after reset window")
	}
	cb.mu.Lock()
	state := cb.entries["job1"].State
	cb.mu.Unlock()
	if state != CircuitHalfOpen {
		t.Fatalf("expected half_open, got %s", state)
	}
}

func TestCircuit_AllReturnsEntries(t *testing.T) {
	cb := newTestCircuitBreaker()
	cb.RecordFailure("jobA")
	cb.RecordFailure("jobB")
	entries := cb.All()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestCircuitHandler_GetContentType(t *testing.T) {
	cb := newTestCircuitBreaker()
	h := CircuitHandler(cb)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics/circuits", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}

func TestCircuitHandler_GetReturnsEntries(t *testing.T) {
	cb := newTestCircuitBreaker()
	for i := 0; i < 3; i++ {
		cb.RecordFailure("jobX")
	}
	h := CircuitHandler(cb)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics/circuits", nil))
	var out []CircuitEntry
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(out) != 1 || out[0].Job != "jobX" || out[0].State != CircuitOpen {
		t.Fatalf("unexpected entries: %+v", out)
	}
}

func TestCircuitHandler_DeleteResetsCircuit(t *testing.T) {
	cb := newTestCircuitBreaker()
	for i := 0; i < 3; i++ {
		cb.RecordFailure("jobY")
	}
	h := CircuitHandler(cb)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/metrics/circuits?job=jobY", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if cb.IsOpen("jobY") {
		t.Fatal("expected circuit to be closed after DELETE")
	}
}
