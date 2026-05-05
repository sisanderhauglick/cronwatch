package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandler_EmptyRegistry(t *testing.T) {
	r := New()
	w := httptest.NewRecorder()
	Handler(r).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var got []JobStats
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d items", len(got))
	}
}

func TestHandler_ReturnsJobStats(t *testing.T) {
	r := New()
	r.RecordSeen("alpha", time.Now())
	r.RecordMissed("beta", time.Now())

	w := httptest.NewRecorder()
	Handler(r).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var got []JobStats
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(got))
	}
	// Handler sorts alphabetically.
	if got[0].Name != "alpha" {
		t.Errorf("first job = %q, want alpha", got[0].Name)
	}
	if got[0].SeenCount != 1 {
		t.Errorf("alpha SeenCount = %d, want 1", got[0].SeenCount)
	}
	if got[1].Name != "beta" {
		t.Errorf("second job = %q, want beta", got[1].Name)
	}
	if got[1].MissedCount != 1 {
		t.Errorf("beta MissedCount = %d, want 1", got[1].MissedCount)
	}
}

func TestHandler_ContentTypeJSON(t *testing.T) {
	r := New()
	w := httptest.NewRecorder()
	Handler(r).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}
