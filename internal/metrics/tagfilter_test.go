package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestTagFilter() (*TagFilter, *Collector) {
	tags := map[string]string{
		"backup":  "env=prod,team=ops",
		"reports": "env=staging,team=platform",
		"cleanup": "env=prod,team=platform",
	}
	tf := NewTagFilter(tags)
	c := NewCollector(DefaultRetentionPolicy())
	return tf, c
}

func TestTagFilter_MatchEmptyFilterAlwaysTrue(t *testing.T) {
	tf, _ := newTestTagFilter()
	if !tf.Match("backup", nil) {
		t.Fatal("empty filter should always match")
	}
}

func TestTagFilter_MatchExactTags(t *testing.T) {
	tf, _ := newTestTagFilter()
	if !tf.Match("backup", map[string]string{"env": "prod", "team": "ops"}) {
		t.Fatal("expected match for backup with env=prod,team=ops")
	}
}

func TestTagFilter_NoMatchWrongValue(t *testing.T) {
	tf, _ := newTestTagFilter()
	if tf.Match("backup", map[string]string{"env": "staging"}) {
		t.Fatal("expected no match for backup with env=staging")
	}
}

func TestTagFilter_NoMatchUnknownJob(t *testing.T) {
	tf, _ := newTestTagFilter()
	if tf.Match("unknown", map[string]string{"env": "prod"}) {
		t.Fatal("expected no match for unknown job")
	}
}

func TestTagFilter_SetTagsUpdates(t *testing.T) {
	tf, _ := newTestTagFilter()
	tf.SetTags("newjob", "env=prod,team=ops")
	if !tf.Match("newjob", map[string]string{"env": "prod"}) {
		t.Fatal("expected match after SetTags")
	}
}

func TestTagFilter_FilterSnapshots(t *testing.T) {
	tf, _ := newTestTagFilter()
	snaps := []Snapshot{
		{Job: "backup", Seen: 1},
		{Job: "reports", Seen: 2},
		{Job: "cleanup", Seen: 3},
	}
	result := tf.FilterSnapshots(snaps, map[string]string{"env": "prod"})
	if len(result) != 2 {
		t.Fatalf("expected 2 prod snapshots, got %d", len(result))
	}
}

func TestTagFilterHandler_ContentType(t *testing.T) {
	tf, c := newTestTagFilter()
	h := NewTagFilterHandler(c, tf)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics/tags", nil))
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}

func TestTagFilterHandler_FiltersByQueryParam(t *testing.T) {
	tf, c := newTestTagFilter()
	now := time.Now()
	_ = c.Collect(Snapshot{Job: "backup", Seen: 1, Timestamp: now})
	_ = c.Collect(Snapshot{Job: "reports", Seen: 2, Timestamp: now})

	h := NewTagFilterHandler(c, tf)
	req := httptest.NewRequest(http.MethodGet, "/metrics/tags?env=prod", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var snaps []Snapshot
	if err := json.NewDecoder(rec.Body).Decode(&snaps); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(snaps) != 1 || snaps[0].Job != "backup" {
		t.Fatalf("expected only backup snapshot, got %+v", snaps)
	}
}
