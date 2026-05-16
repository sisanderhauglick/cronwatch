package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestFingerprintStore() *FingerprintStore {
	return NewFingerprintStore(24 * time.Hour)
}

func TestFingerprint_RecordNewEntry(t *testing.T) {
	s := newTestFingerprintStore()
	now := time.Now()
	hash := s.Record("backup", "missed", now)
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	rec := s.Get(hash)
	if rec == nil {
		t.Fatal("expected record, got nil")
	}
	if rec.Job != "backup" || rec.Reason != "missed" {
		t.Errorf("unexpected record: %+v", rec)
	}
	if rec.Count != 1 {
		t.Errorf("expected count 1, got %d", rec.Count)
	}
}

func TestFingerprint_IncrementOnDuplicate(t *testing.T) {
	s := newTestFingerprintStore()
	now := time.Now()
	h1 := s.Record("backup", "missed", now)
	h2 := s.Record("backup", "missed", now.Add(time.Minute))
	if h1 != h2 {
		t.Errorf("expected same hash, got %s and %s", h1, h2)
	}
	rec := s.Get(h1)
	if rec.Count != 2 {
		t.Errorf("expected count 2, got %d", rec.Count)
	}
}

func TestFingerprint_DifferentReasonDifferentHash(t *testing.T) {
	s := newTestFingerprintStore()
	now := time.Now()
	h1 := s.Record("backup", "missed", now)
	h2 := s.Record("backup", "failed", now)
	if h1 == h2 {
		t.Error("expected different hashes for different reasons")
	}
}

func TestFingerprint_PrunesOldEntries(t *testing.T) {
	s := NewFingerprintStore(time.Hour)
	old := time.Now().Add(-2 * time.Hour)
	s.Record("backup", "missed", old)
	all := s.All(time.Now())
	if len(all) != 0 {
		t.Errorf("expected 0 entries after prune, got %d", len(all))
	}
}

func TestFingerprintHandler_ContentType(t *testing.T) {
	s := newTestFingerprintStore()
	h := FingerprintHandler(s)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/fingerprints", nil))
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("unexpected content-type: %s", ct)
	}
}

func TestFingerprintHandler_EmptyReturnsEmptySlice(t *testing.T) {
	s := newTestFingerprintStore()
	h := FingerprintHandler(s)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/fingerprints", nil))
	body := rec.Body.String()
	if !strings.Contains(body, "[]") {
		t.Errorf("expected empty JSON array, got: %s", body)
	}
}

func TestFingerprintHandler_DeleteRemovesEntry(t *testing.T) {
	s := newTestFingerprintStore()
	hash := s.Record("cleanup", "failed", time.Now())
	h := FingerprintHandler(s)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodDelete, "/fingerprints?hash="+hash, nil))
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
	if s.Get(hash) != nil {
		t.Error("expected entry to be deleted")
	}
}
