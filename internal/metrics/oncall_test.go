package metrics

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestOncallManager() *OncallManager {
	return NewOncallManager()
}

func TestOncall_ActiveAtNoShifts(t *testing.T) {
	m := newTestOncallManager()
	if got := m.ActiveAt(time.Now()); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestOncall_ActiveAtMatchesShift(t *testing.T) {
	m := newTestOncallManager()
	now := time.Now()
	m.AddShift(OncallShift{Name: "alice", Start: now.Add(-time.Hour), End: now.Add(time.Hour)})
	got := m.ActiveAt(now)
	if got == nil || got.Name != "alice" {
		t.Errorf("expected alice shift, got %v", got)
	}
}

func TestOncall_ActiveAtOutsideShift(t *testing.T) {
	m := newTestOncallManager()
	now := time.Now()
	m.AddShift(OncallShift{Name: "bob", Start: now.Add(time.Hour), End: now.Add(2 * time.Hour)})
	if got := m.ActiveAt(now); got != nil {
		t.Errorf("expected nil for future shift, got %v", got)
	}
}

func TestOncall_PruneRemovesExpired(t *testing.T) {
	m := newTestOncallManager()
	now := time.Now()
	m.AddShift(OncallShift{Name: "old", Start: now.Add(-3 * time.Hour), End: now.Add(-2 * time.Hour)})
	m.AddShift(OncallShift{Name: "current", Start: now.Add(-time.Hour), End: now.Add(time.Hour)})
	m.Prune(now)
	shifts := m.All()
	if len(shifts) != 1 || shifts[0].Name != "current" {
		t.Errorf("expected only current shift after prune, got %v", shifts)
	}
}

func TestOncallHandler_GetContentType(t *testing.T) {
	h := NewOncallHandler(newTestOncallManager())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/oncall", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestOncallHandler_PostAddsShift(t *testing.T) {
	m := newTestOncallManager()
	h := NewOncallHandler(m)
	body, _ := json.Marshal(map[string]string{
		"name":  "carol",
		"start": "2024-01-01T08:00:00Z",
		"end":   "2024-01-01T16:00:00Z",
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/oncall", bytes.NewReader(body)))
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	if len(m.All()) != 1 {
		t.Errorf("expected 1 shift, got %d", len(m.All()))
	}
}

func TestOncallHandler_PostMissingName(t *testing.T) {
	h := NewOncallHandler(newTestOncallManager())
	body, _ := json.Marshal(map[string]string{
		"start": "2024-01-01T08:00:00Z",
		"end":   "2024-01-01T16:00:00Z",
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/oncall", bytes.NewReader(body)))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
