package metrics

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestInhibitHandler() (*InhibitHandler, *Registry) {
	r := New()
	rules := []InhibitRule{{SourceJob: "src", TargetJob: "tgt"}}
	m := NewInhibitManager(r, rules, 5*time.Minute)
	return NewInhibitHandler(m), r
}

func TestInhibitHandler_ContentType(t *testing.T) {
	h, _ := newTestInhibitHandler()
	req := httptest.NewRequest(http.MethodGet, "/metrics/inhibit?job=tgt", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if ct := rw.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("unexpected content-type: %s", ct)
	}
}

func TestInhibitHandler_NotInhibitedWhenSourceHealthy(t *testing.T) {
	h, r := newTestInhibitHandler()
	r.RecordSeen("src")

	req := httptest.NewRequest(http.MethodGet, "/metrics/inhibit?job=tgt", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	var resp inhibitCheckResponse
	if err := json.NewDecoder(rw.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Inhibited {
		t.Fatal("expected not inhibited")
	}
}

func TestInhibitHandler_InhibitedWhenSourceFailed(t *testing.T) {
	h, r := newTestInhibitHandler()
	r.RecordFailed("src")

	req := httptest.NewRequest(http.MethodGet, "/metrics/inhibit?job=tgt", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	var resp inhibitCheckResponse
	if err := json.NewDecoder(rw.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if !resp.Inhibited {
		t.Fatal("expected inhibited")
	}
}

func TestInhibitHandler_PostAddsRule(t *testing.T) {
	h, _ := newTestInhibitHandler()
	body := `{"source_job":"new-src","target_job":"new-tgt"}`
	req := httptest.NewRequest(http.MethodPost, "/metrics/inhibit", bytes.NewBufferString(body))
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rw.Code)
	}
	if len(h.manager.Rules()) != 2 {
		t.Fatalf("expected 2 rules after POST, got %d", len(h.manager.Rules()))
	}
}

func TestInhibitHandler_PostBadBody(t *testing.T) {
	h, _ := newTestInhibitHandler()
	req := httptest.NewRequest(http.MethodPost, "/metrics/inhibit", bytes.NewBufferString("not-json"))
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rw.Code)
	}
}

func TestInhibitHandler_MethodNotAllowed(t *testing.T) {
	h, _ := newTestInhibitHandler()
	req := httptest.NewRequest(http.MethodDelete, "/metrics/inhibit", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rw.Code)
	}
}
