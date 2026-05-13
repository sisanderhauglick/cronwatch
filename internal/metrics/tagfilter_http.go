package metrics

import (
	"encoding/json"
	"net/http"
	"strings"
)

// TagFilterHandler serves filtered snapshots via HTTP query params.
// Query params: any number of key=value pairs become the tag filter.
// Reserved param "_job" is ignored (use job-level endpoints for that).
//
// GET /metrics/tags?env=prod&team=platform
type TagFilterHandler struct {
	collector *Collector
	filter    *TagFilter
}

// NewTagFilterHandler creates a handler backed by the given collector and filter.
func NewTagFilterHandler(c *Collector, tf *TagFilter) *TagFilterHandler {
	return &TagFilterHandler{collector: c, filter: tf}
}

func (h *TagFilterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tagFilter := parseQueryTags(r)

	all := h.collector.All()
	// Flatten all snapshots from all history slices.
	var flat []Snapshot
	for _, snaps := range all {
		flat = append(flat, snaps...)
	}

	matched := h.filter.FilterSnapshots(flat, tagFilter)

	w.Header().Set("Content-Type", "application/json")
	if matched == nil {
		matched = []Snapshot{}
	}
	_ = json.NewEncoder(w).Encode(matched)
}

// parseQueryTags converts URL query params into a tag filter map,
// skipping the reserved "_job" key.
func parseQueryTags(r *http.Request) map[string]string {
	m := make(map[string]string)
	for k, vals := range r.URL.Query() {
		if strings.HasPrefix(k, "_") || len(vals) == 0 {
			continue
		}
		m[k] = vals[0]
	}
	return m
}
