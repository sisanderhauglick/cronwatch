package metrics

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// DigestHandler serves a JSON digest report over the configured window.
// An optional query param ?window=<minutes> overrides the analyzer default.
type DigestHandler struct {
	analyzer *DigestAnalyzer
}

// NewDigestHandler returns an http.Handler backed by the given DigestAnalyzer.
func NewDigestHandler(a *DigestAnalyzer) http.Handler {
	return &DigestHandler{analyzer: a}
}

func (h *DigestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if q := r.URL.Query().Get("window"); q != "" {
		if mins, err := strconv.Atoi(q); err == nil && mins > 0 {
			orig := h.analyzer.window
			h.analyzer.window = time.Duration(mins) * time.Minute
			defer func() { h.analyzer.window = orig }()
		}
	}
	report := h.analyzer.Summarize()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(report)
}
