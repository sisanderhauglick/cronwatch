package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestTagIndex() *TagIndex {
	return NewTagIndex()
}

func TestTagIndex_LookupEmpty(t *testing.T) {
	ti := newTestTagIndex()
	result := ti.Lookup("env", "prod")
	if len(result) != 0 {
		t.Fatalf("expected empty, got %v", result)
	}
}

func TestTagIndex_AddAndLookup(t *testing.T) {
	ti := newTestTagIndex()
	ti.Add("backup", map[string]string{"env": "prod", "team": "infra"})
	ti.Add("cleanup", map[string]string{"env": "prod"})

	jobs := ti.Lookup("env", "prod")
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %v", jobs)
	}
	if jobs[0] != "backup" || jobs[1] != "cleanup" {
		t.Errorf("unexpected order: %v", jobs)
	}
}

func TestTagIndex_LookupByTeam(t *testing.T) {
	ti := newTestTagIndex()
	ti.Add("backup", map[string]string{"team": "infra"})
	ti.Add("deploy", map[string]string{"team": "platform"})

	result := ti.Lookup("team", "infra")
	if len(result) != 1 || result[0] != "backup" {
		t.Errorf("expected [backup], got %v", result)
	}
}

func TestTagIndex_Keys(t *testing.T) {
	ti := newTestTagIndex()
	ti.Add("job1", map[string]string{"env": "prod", "region": "us-east"})

	keys := ti.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %v", keys)
	}
	if keys[0] != "env" || keys[1] != "region" {
		t.Errorf("unexpected keys: %v", keys)
	}
}

func TestTagIndex_DuplicateAddIdempotent(t *testing.T) {
	ti := newTestTagIndex()
	ti.Add("job1", map[string]string{"env": "prod"})
	ti.Add("job1", map[string]string{"env": "prod"})

	result := ti.Lookup("env", "prod")
	if len(result) != 1 {
		t.Errorf("expected 1, got %d", len(result))
	}
}

func TestTagIndexHandler_ListsKeys(t *testing.T) {
	ti := newTestTagIndex()
	ti.Add("job1", map[string]string{"env": "staging"})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics/tags", nil)
	TagIndexHandler(ti)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
	var keys []string
	if err := json.NewDecoder(rec.Body).Decode(&keys); err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 || keys[0] != "env" {
		t.Errorf("unexpected keys: %v", keys)
	}
}

func TestTagIndexHandler_LookupByKeyValue(t *testing.T) {
	ti := newTestTagIndex()
	ti.Add("nightly", map[string]string{"env": "prod"})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics/tags?key=env&value=prod", nil)
	TagIndexHandler(ti)(rec, req)

	var jobs []string
	if err := json.NewDecoder(rec.Body).Decode(&jobs); err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 || jobs[0] != "nightly" {
		t.Errorf("expected [nightly], got %v", jobs)
	}
}
