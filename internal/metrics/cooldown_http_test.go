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

func newTestCooldownHandler() (*CooldownManager, http.Handler) {
	cm := NewCooldownManager()
	cm.now = func() time.Time { return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC) }
	return cm, NewCooldownHandler(cm)
}

func TestCooldownHandler_ContentType(t *testing.T) {
	_, h := newTestCooldownHandler()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content-type: %s", ct)
	}
}

func TestCooldownHandler_EmptyReturnsEmptySlice(t *testing.T) {
	_, h := newTestCooldownHandler()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	var out []CooldownEntry
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty slice, got %d", len(out))
	}
}

func TestCooldownHandler_PostActivates(t *testing.T) {
	cm, h := newTestCooldownHandler()
	body := bytes.NewBufferString(`{"duration":"10m"}`)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/myjob", body))
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if !cm.InCooldown("myjob") {
		t.Fatal("expected job to be in cooldown")
	}
}

func TestCooldownHandler_DeleteResets(t *testing.T) {
	cm, h := newTestCooldownHandler()
	cm.SetCooldown("myjob", 10*time.Minute)
	cm.Activate("myjob")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodDelete, "/myjob", nil))
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if cm.InCooldown("myjob") {
		t.Fatal("expected cooldown cleared")
	}
}

func TestCooldownHandler_PostInvalidDuration(t *testing.T) {
	_, h := newTestCooldownHandler()
	body := strings.NewReader(`{"duration":"bad"}`)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/myjob", body))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCooldownHandler_MethodNotAllowed(t *testing.T) {
	_, h := newTestCooldownHandler()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPatch, "/myjob", nil))
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}
