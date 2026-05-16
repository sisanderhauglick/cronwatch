package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestDedupHandler() (*DedupManager, *DedupHandler) {
	m := NewDedupManager(5 * time.Minute)
	return m, NewDedupHandler(m)
}

func TestDedupHandler_ContentType(t *testing.T) {
	_, h := newTestDedupHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dedup", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}

func TestDedupHandler_EmptyReturnsEmptySlice(t *testing.T) {
	_, h := newTestDedupHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dedup", nil))
	var out []interface{}
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty slice, got %d entries", len(out))
	}
}

func TestDedupHandler_ReturnsEntries(t *testing.T) {
	m, h := newTestDedupHandler()
	m.IsDuplicate("backup", "missed", time.Now())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dedup", nil))
	var out []map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out))
	}
	if out[0]["job"] != "backup" {
		t.Fatalf("expected job=backup, got %v", out[0]["job"])
	}
}

func TestDedupHandler_DeleteResetsEntry(t *testing.T) {
	m, h := newTestDedupHandler()
	now := time.Now()
	m.IsDuplicate("backup", "missed", now)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/dedup?job=backup&reason=missed", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if m.IsDuplicate("backup", "missed", now.Add(time.Second)) {
		t.Fatal("expected entry to be cleared after DELETE")
	}
}

func TestDedupHandler_DeleteMissingParamsBadRequest(t *testing.T) {
	_, h := newTestDedupHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/dedup?job=backup", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
