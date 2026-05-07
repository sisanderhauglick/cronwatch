package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestHealthHandler() (http.HandlerFunc, *Registry) {
	reg := New()
	eval := NewHealthEvaluator()
	return HealthHandler(reg, eval), reg
}

func TestHealthHandler_EmptyRegistry(t *testing.T) {
	h, _ := newTestHealthHandler()
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	var resp HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Overall != HealthUnknown {
		t.Errorf("expected unknown overall, got %s", resp.Overall)
	}
}

func TestHealthHandler_ContentType(t *testing.T) {
	h, _ := newTestHealthHandler()
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("unexpected content-type: %s", ct)
	}
}

func TestHealthHandler_OKJob(t *testing.T) {
	h, reg := newTestHealthHandler()
	reg.RecordSeen("nightly")
	// Manually set LastSeen to recent time via RecordSeen
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	var resp HealthResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(resp.Jobs))
	}
	if resp.Jobs[0].Status != HealthOK {
		t.Errorf("expected ok, got %s", resp.Jobs[0].Status)
	}
	if resp.Overall != HealthOK {
		t.Errorf("expected overall ok, got %s", resp.Overall)
	}
}

func TestHealthHandler_DownWhenMissed(t *testing.T) {
	h, reg := newTestHealthHandler()
	reg.RecordSeen("nightly")
	reg.RecordMissed("nightly")
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	var resp HealthResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Overall != HealthDown {
		t.Errorf("expected overall down, got %s", resp.Overall)
	}
}

func TestHealthHandler_GeneratedAtIsRecent(t *testing.T) {
	h, _ := newTestHealthHandler()
	before := time.Now().UTC()
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	var resp HealthResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.GeneratedAt.Before(before) {
		t.Error("GeneratedAt should be after test start")
	}
}
