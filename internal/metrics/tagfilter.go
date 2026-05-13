package metrics

import "strings"

// TagFilter allows filtering job snapshots by user-defined tags.
// Tags are arbitrary key=value pairs attached to job names via config labels.
type TagFilter struct {
	tags map[string]string // job name -> comma-separated key=value tags
}

// NewTagFilter creates a TagFilter seeded with a job->tags mapping.
func NewTagFilter(tags map[string]string) *TagFilter {
	copy := make(map[string]string, len(tags))
	for k, v := range tags {
		copy[k] = v
	}
	return &TagFilter{tags: copy}
}

// SetTags updates the tag string for a given job.
func (tf *TagFilter) SetTags(job, tagStr string) {
	tf.tags[job] = tagStr
}

// Match returns true if the job's tags contain all of the requested key=value pairs.
// An empty filter always matches.
func (tf *TagFilter) Match(job string, filter map[string]string) bool {
	if len(filter) == 0 {
		return true
	}
	raw, ok := tf.tags[job]
	if !ok {
		return false
	}
	parsed := parseTags(raw)
	for k, v := range filter {
		if parsed[k] != v {
			return false
		}
	}
	return true
}

// FilterSnapshots returns only those snapshots whose job name matches the filter.
func (tf *TagFilter) FilterSnapshots(snaps []Snapshot, filter map[string]string) []Snapshot {
	if len(filter) == 0 {
		return snaps
	}
	out := make([]Snapshot, 0, len(snaps))
	for _, s := range snaps {
		if tf.Match(s.Job, filter) {
			out = append(out, s)
		}
	}
	return out
}

// parseTags splits "key=value,key2=value2" into a map.
func parseTags(raw string) map[string]string {
	m := make(map[string]string)
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			m[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return m
}
