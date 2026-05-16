package metrics

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestCheckpointStore() *CheckpointStore {
	return NewCheckpointStore()
}

func TestCheckpoint_GetMissingReturnsNotFound(t *testing.T) {
	s := newTestCheckpointStore()
	_, ok := s.Get("nonexistent")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestCheckpoint_RecordAndGet(t *testing.T) {
	s := newTestCheckpointStore()
	now := time.Now().UTC().Truncate(time.Second)
	s.Record("backup", now)
	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry")
	}
	if !e.LastOK.Equal(now) {
		t.Fatalf("got %v want %v", e.LastOK, now)
	}
}

func TestCheckpoint_StaleBefore(t *testing.T) {
	s := newTestCheckpointStore()
	old := time.Now().Add(-2 * time.Hour)
	recent := time.Now().Add(-1 * time.Minute)
	s.Record("old-job", old)
	s.Record("new-job", recent)
	stale := s.StaleBefore(time.Now().Add(-30 * time.Minute))
	if len(stale) != 1 || stale[0].Job != "old-job" {
		t.Fatalf("unexpected stale entries: %+v", stale)
	}
}

func TestCheckpointHandler_GetAll(t *testing.T) {
	s := newTestCheckpointStore()
	s.Record("jobA", time.Now())
	h := NewCheckpointHandler(s)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/checkpoints", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type %q", ct)
	}
	var entries []CheckpointEntry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestCheckpointHandler_PostRecords(t *testing.T) {
	s := newTestCheckpointStore()
	h := NewCheckpointHandler(s)
	now := time.Now().UTC().Truncate(time.Second)
	body, _ := json.Marshal(map[string]interface{}{"last_ok": now})
	req := httptest.NewRequest(http.MethodPost, "/checkpoints?job=deploy", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status %d", rec.Code)
	}
	e, ok := s.Get("deploy")
	if !ok || !e.LastOK.Equal(now) {
		t.Fatalf("checkpoint not recorded correctly: %+v", e)
	}
}

func TestCheckpointHandler_PostMissingJobParam(t *testing.T) {
	s := newTestCheckpointStore()
	h := NewCheckpointHandler(s)
	body, _ := json.Marshal(map[string]interface{}{"last_ok": time.Now()})
	req := httptest.NewRequest(http.MethodPost, "/checkpoints", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
