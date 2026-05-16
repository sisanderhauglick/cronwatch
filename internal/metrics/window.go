package metrics

import (
	"net/http"
	"sync"
	"time"
)

// WindowConfig holds a sliding-window duration that can be read and updated
// concurrently. It is used by analysers that accept a configurable look-back
// period so that operators can tune the window at runtime without restarting
// the daemon.
type WindowConfig struct {
	mu       sync.RWMutex
	duration time.Duration
}

// NewWindowConfig returns a WindowConfig initialised to d.
func NewWindowConfig(d time.Duration) *WindowConfig {
	return &WindowConfig{duration: d}
}

// Get returns the current window duration.
func (w *WindowConfig) Get() time.Duration {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.duration
}

// Set replaces the window duration. Returns false if d <= 0.
func (w *WindowConfig) Set(d time.Duration) bool {
	if d <= 0 {
		return false
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.duration = d
	return true
}

// WindowHandler exposes GET / PUT for the window duration over HTTP.
// GET returns the current value as plain text (e.g. "1h0m0s").
// PUT expects a duration string in the request body (e.g. "30m").
func WindowHandler(w *WindowConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			rw.Header().Set("Content-Type", "text/plain")
			_, _ = rw.Write([]byte(w.Get().String()))

		case http.MethodPut:
			var buf [64]byte
			n, _ := r.Body.Read(buf[:])
			body := string(buf[:n])
			d, err := time.ParseDuration(body)
			if err != nil || !w.Set(d) {
				http.Error(rw, "invalid duration", http.StatusBadRequest)
				return
			}
			rw.Header().Set("Content-Type", "text/plain")
			_, _ = rw.Write([]byte(w.Get().String()))

		default:
			http.Error(rw, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
