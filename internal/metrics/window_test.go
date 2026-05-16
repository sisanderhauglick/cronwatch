package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestWindowConfig(d time.Duration) *WindowConfig {
	return NewWindowConfig(d)
}

func TestWindowConfig_GetDefault(t *testing.T) {
	w := newTestWindowConfig(time.Hour)
	if w.Get() != time.Hour {
		t.Fatalf("expected 1h, got %s", w.Get())
	}
}

func TestWindowConfig_SetValid(t *testing.T) {
	w := newTestWindowConfig(time.Hour)
	if !w.Set(30 * time.Minute) {
		t.Fatal("expected Set to return true")
	}
	if w.Get() != 30*time.Minute {
		t.Fatalf("expected 30m, got %s", w.Get())
	}
}

func TestWindowConfig_SetInvalidReturnsFalse(t *testing.T) {
	w := newTestWindowConfig(time.Hour)
	if w.Set(0) {
		t.Fatal("expected Set(0) to return false")
	}
	if w.Set(-time.Minute) {
		t.Fatal("expected Set(-1m) to return false")
	}
	// original value unchanged
	if w.Get() != time.Hour {
		t.Fatalf("expected 1h unchanged, got %s", w.Get())
	}
}

func TestWindowHandler_GetContentType(t *testing.T) {
	w := newTestWindowConfig(time.Hour)
	h := WindowHandler(w)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/window", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/plain" {
		t.Fatalf("unexpected content-type: %s", ct)
	}
	if body := rec.Body.String(); body != "1h0m0s" {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestWindowHandler_PutUpdates(t *testing.T) {
	w := newTestWindowConfig(time.Hour)
	h := WindowHandler(w)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/window", strings.NewReader("45m")))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if w.Get() != 45*time.Minute {
		t.Fatalf("expected 45m, got %s", w.Get())
	}
}

func TestWindowHandler_PutInvalidDuration(t *testing.T) {
	w := newTestWindowConfig(time.Hour)
	h := WindowHandler(w)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/window", strings.NewReader("not-a-duration")))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestWindowHandler_MethodNotAllowed(t *testing.T) {
	w := newTestWindowConfig(time.Hour)
	h := WindowHandler(w)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/window", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
