package metrics

import (
	"net/http"
	"sort"
	"sync"
)

// TagIndex maintains an inverted index from tag key/value pairs to job names,
// allowing fast lookups of jobs by tag.
type TagIndex struct {
	mu    sync.RWMutex
	index map[string]map[string]struct{} // "key=value" -> set of job names
}

// NewTagIndex creates an empty TagIndex.
func NewTagIndex() *TagIndex {
	return &TagIndex{
		index: make(map[string]map[string]struct{}),
	}
}

// Add registers a job name under each of its tags.
func (ti *TagIndex) Add(job string, tags map[string]string) {
	ti.mu.Lock()
	defer ti.mu.Unlock()
	for k, v := range tags {
		key := k + "=" + v
		if ti.index[key] == nil {
			ti.index[key] = make(map[string]struct{})
		}
		ti.index[key][job] = struct{}{}
	}
}

// Lookup returns all job names that have the given tag key=value pair.
func (ti *TagIndex) Lookup(tagKey, tagValue string) []string {
	ti.mu.RLock()
	defer ti.mu.RUnlock()
	key := tagKey + "=" + tagValue
	set, ok := ti.index[key]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(set))
	for job := range set {
		out = append(out, job)
	}
	sort.Strings(out)
	return out
}

// Keys returns all unique tag keys present in the index.
func (ti *TagIndex) Keys() []string {
	ti.mu.RLock()
	defer ti.mu.RUnlock()
	seen := make(map[string]struct{})
	for kv := range ti.index {
		// extract key before '='
		for i, c := range kv {
			if c == '=' {
				seen[kv[:i]] = struct{}{}
				break
			}
		}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TagIndexHandler serves a JSON view of the tag index.
// GET /metrics/tags?key=env returns jobs tagged with that key.
// GET /metrics/tags lists all tag keys.
func TagIndexHandler(ti *TagIndex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		key := r.URL.Query().Get("key")
		value := r.URL.Query().Get("value")
		if key != "" && value != "" {
			jobs := ti.Lookup(key, value)
			writeJSON(w, jobs)
			return
		}
		keys := ti.Keys()
		writeJSON(w, keys)
	}
}
