package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestThrottleHandler() (*ThrottleManager, *ThrottleHandler) {
	tm := newTestThrottleManager(5 * time.Minute)
	return tm, NewThrottleHandler(tm)
}

func TestThrottleHandler_ContentType(t *testing.T) {
	_, h := newTestThrottleHandler()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics/throttle", nil))
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}

func TestThrottleHandler_EmptyReturnsEmptySlice(t *testing.T) {
	_, h := newTestThrottleHandler()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics/throttle", nil))
	var out []ThrottleEntry
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty slice, got %d items", len(out))
	}
}

func TestThrottleHandler_GetSingleEntry(t *testing.T) {
	tm, h := newTestThrottleHandler()
	tm.Allow("job-a")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics/throttle/job-a", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var e ThrottleEntry
	if err := json.NewDecoder(rr.Body).Decode(&e); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if e.JobName != "job-a" {
		t.Fatalf("expected job-a, got %s", e.JobName)
	}
}

func TestThrottleHandler_GetMissingReturns404(t *testing.T) {
	_, h := newTestThrottleHandler()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics/throttle/unknown", nil))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestThrottleHandler_DeleteResetsEntry(t *testing.T) {
	tm, h := newTestThrottleHandler()
	tm.Allow("job-b")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodDelete, "/metrics/throttle/job-b", nil))
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	_, ok := tm.Stats("job-b")
	if ok {
		t.Fatal("expected entry to be deleted")
	}
}
