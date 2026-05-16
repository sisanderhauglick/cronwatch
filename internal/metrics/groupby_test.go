package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestGroupByAnalyzer() (*GroupByAnalyzer, *Collector, *TagIndex) {
	c := NewCollector(10)
	ti := NewTagIndex()
	a := NewGroupByAnalyzer(c, ti, time.Hour)
	return a, c, ti
}

func TestGroupBy_EmptyCollector(t *testing.T) {
	a, _, _ := newTestGroupByAnalyzer()
	res := a.Summarize("team")
	if len(res) != 0 {
		t.Fatalf("expected empty, got %d results", len(res))
	}
}

func TestGroupBy_GroupsByTagKey(t *testing.T) {
	a, c, ti := newTestGroupByAnalyzer()

	ti.Add("job-a", map[string]string{"team": "ops"})
	ti.Add("job-b", map[string]string{"team": "dev"})
	ti.Add("job-c", map[string]string{"team": "ops"})

	snap := Snapshot{
		CollectedAt: time.Now(),
		Jobs: map[string]JobStats{
			"job-a": {SeenCount: 5, FailedCount: 1},
			"job-b": {SeenCount: 3, FailedCount: 0},
			"job-c": {SeenCount: 4, MissedCount: 1},
		},
	}
	c.Collect(snap)

	res := a.Summarize("team")
	groups := map[string]GroupByResult{}
	for _, r := range res {
		groups[r.Group] = r
	}

	ops := groups["ops"]
	if ops.JobCount != 2 {
		t.Errorf("ops job count: want 2, got %d", ops.JobCount)
	}
	if ops.TotalRuns != 9 {
		t.Errorf("ops total runs: want 9, got %d", ops.TotalRuns)
	}
	if ops.TotalFailed != 1 {
		t.Errorf("ops failed: want 1, got %d", ops.TotalFailed)
	}

	dev := groups["dev"]
	if dev.JobCount != 1 {
		t.Errorf("dev job count: want 1, got %d", dev.JobCount)
	}
	if dev.SuccessRate != 1.0 {
		t.Errorf("dev success rate: want 1.0, got %f", dev.SuccessRate)
	}
}

func TestGroupBy_UntaggedFallback(t *testing.T) {
	a, c, _ := newTestGroupByAnalyzer()

	snap := Snapshot{
		CollectedAt: time.Now(),
		Jobs: map[string]JobStats{
			"job-x": {SeenCount: 2},
		},
	}
	c.Collect(snap)

	res := a.Summarize("team")
	if len(res) != 1 || res[0].Group != "(untagged)" {
		t.Fatalf("expected untagged group, got %+v", res)
	}
}

func TestGroupByHandler_ContentTypeAndBody(t *testing.T) {
	a, c, ti := newTestGroupByAnalyzer()
	ti.Add("job-a", map[string]string{"env": "prod"})
	c.Collect(Snapshot{
		CollectedAt: time.Now(),
		Jobs: map[string]JobStats{"job-a": {SeenCount: 1}},
	})

	h := GroupByHandler(a, time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/groupby?key=env", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type: want application/json, got %s", ct)
	}
	var results []GroupByResult
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
}

func TestGroupByHandler_MissingKeyReturns400(t *testing.T) {
	a, _, _ := newTestGroupByAnalyzer()
	h := GroupByHandler(a, time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/groupby", nil)
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
