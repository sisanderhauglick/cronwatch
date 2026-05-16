package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestRetentionPolicyHandler() (*RetentionPolicyHandler, *RetentionPolicy) {
	p := DefaultRetentionPolicy()
	return NewRetentionPolicyHandler(p), p
}

func TestRetentionPolicyHandler_GetContentType(t *testing.T) {
	h, _ := newTestRetentionPolicyHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/retention", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}

func TestRetentionPolicyHandler_GetReturnsDefaults(t *testing.T) {
	h, p := newTestRetentionPolicyHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/retention", nil))

	var resp retentionPolicyResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.MaxSnapshots != p.MaxSnapshots {
		t.Errorf("expected %d snapshots, got %d", p.MaxSnapshots, resp.MaxSnapshots)
	}
}

func TestRetentionPolicyHandler_PutUpdatesMaxAge(t *testing.T) {
	h, p := newTestRetentionPolicyHandler()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/retention?max_age=2h", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if p.MaxAge != 2*time.Hour {
		t.Errorf("expected 2h, got %s", p.MaxAge)
	}
}

func TestRetentionPolicyHandler_PutUpdatesMaxSnapshots(t *testing.T) {
	h, p := newTestRetentionPolicyHandler()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/retention?max_snapshots=50", nil)
	h.ServeHTTP(rec, req)

	if p.MaxSnapshots != 50 {
		t.Errorf("expected 50, got %d", p.MaxSnapshots)
	}
}

func TestRetentionPolicyHandler_PutInvalidMaxAge(t *testing.T) {
	h, _ := newTestRetentionPolicyHandler()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/retention?max_age=notaduration", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRetentionPolicyHandler_PutInvalidMaxSnapshots(t *testing.T) {
	h, _ := newTestRetentionPolicyHandler()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/retention?max_snapshots=0", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRetentionPolicyHandler_MethodNotAllowed(t *testing.T) {
	h, _ := newTestRetentionPolicyHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/retention", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
