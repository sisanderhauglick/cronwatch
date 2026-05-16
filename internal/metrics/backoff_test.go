package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestBackoffManager() *BackoffManager {
	return NewBackoffManager(1*time.Second, 60*time.Second)
}

func TestBackoff_FirstAlwaysAllowed(t *testing.T) {
	b := newTestBackoffManager()
	now := time.Now()
	if !b.Allow("job1", now) {
		t.Fatal("expected first attempt to be allowed")
	}
}

func TestBackoff_SuppressedAfterFailure(t *testing.T) {
	b := newTestBackoffManager()
	now := time.Now()
	b.RecordFailure("job1", "timeout", now)
	if b.Allow("job1", now) {
		t.Fatal("expected attempt to be suppressed immediately after failure")
	}
}

func TestBackoff_AllowedAfterWait(t *testing.T) {
	b := newTestBackoffManager()
	now := time.Now()
	b.RecordFailure("job1", "timeout", now)
	future := now.Add(2 * time.Second)
	if !b.Allow("job1", future) {
		t.Fatal("expected attempt to be allowed after backoff window")
	}
}

func TestBackoff_ExponentialGrowth(t *testing.T) {
	b := newTestBackoffManager()
	now := time.Now()
	b.RecordFailure("job1", "err", now)
	entries := b.All()
	first := entries[0].NextRetry

	b.RecordFailure("job1", "err", now)
	entries = b.All()
	second := entries[0].NextRetry

	if !second.After(first) {
		t.Errorf("expected second retry to be later than first: first=%v second=%v", first, second)
	}
}

func TestBackoff_ResetAllowsImmediately(t *testing.T) {
	b := newTestBackoffManager()
	now := time.Now()
	b.RecordFailure("job1", "err", now)
	b.Reset("job1")
	if !b.Allow("job1", now) {
		t.Fatal("expected job to be allowed after reset")
	}
}

func TestBackoffHandler_ContentType(t *testing.T) {
	b := newTestBackoffManager()
	h := BackoffHandler(b)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/backoff", nil))
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("unexpected content-type: %s", ct)
	}
}

func TestBackoffHandler_EmptyReturnsEmptySlice(t *testing.T) {
	b := newTestBackoffManager()
	h := BackoffHandler(b)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/backoff", nil))
	var out []BackoffEntry
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(out))
	}
}

func TestBackoffHandler_DeleteResetsEntry(t *testing.T) {
	b := newTestBackoffManager()
	b.RecordFailure("job1", "err", time.Now())
	h := BackoffHandler(b)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/backoff?job=job1", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
	if !b.Allow("job1", time.Now()) {
		t.Error("expected job to be allowed after DELETE")
	}
}
