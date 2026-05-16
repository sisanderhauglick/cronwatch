package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRegisterRoutes_DedupEndpoint(t *testing.T) {
	reg := New()
	col := NewCollector(10)
	tracker := NewAlertTracker(100)
	dedup := NewDedupManager(5 * time.Minute)

	mux := http.NewServeMux()
	RegisterRoutes(mux, reg, col, tracker)
	mux.Handle("/metrics/dedup", NewDedupHandler(dedup))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics/dedup", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from /metrics/dedup, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}
