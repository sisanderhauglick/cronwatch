package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestParsePageRequest_Defaults(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	pr := ParsePageRequest(r)
	if pr.Page != 1 {
		t.Fatalf("expected page 1, got %d", pr.Page)
	}
	if pr.PageSize != defaultPageSize {
		t.Fatalf("expected page_size %d, got %d", defaultPageSize, pr.PageSize)
	}
}

func TestParsePageRequest_ClampsMaxSize(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?page_size=9999", nil)
	pr := ParsePageRequest(r)
	if pr.PageSize != maxPageSize {
		t.Fatalf("expected page_size clamped to %d, got %d", maxPageSize, pr.PageSize)
	}
}

func TestParsePageRequest_InvalidFallsToDefault(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?page=abc&page_size=xyz", nil)
	pr := ParsePageRequest(r)
	if pr.Page != 1 || pr.PageSize != defaultPageSize {
		t.Fatalf("unexpected values: page=%d size=%d", pr.Page, pr.PageSize)
	}
}

func TestPaginate_SlicesCorrectly(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	resp := Paginate(items, PageRequest{Page: 2, PageSize: 2})
	if resp.TotalItems != 5 {
		t.Fatalf("expected total 5, got %d", resp.TotalItems)
	}
	if resp.TotalPages != 3 {
		t.Fatalf("expected 3 pages, got %d", resp.TotalPages)
	}
	slice, ok := resp.Items.([]int)
	if !ok || len(slice) != 2 || slice[0] != 3 {
		t.Fatalf("unexpected items: %v", resp.Items)
	}
}

func TestPaginate_PageBeyondEnd(t *testing.T) {
	items := []int{1, 2}
	resp := Paginate(items, PageRequest{Page: 10, PageSize: 5})
	slice := resp.Items.([]int)
	if len(slice) != 0 {
		t.Fatalf("expected empty slice for out-of-range page")
	}
}

func TestPaginatedRunLogHandler_ContentTypeAndBody(t *testing.T) {
	rl := NewRunLog(50)
	now := time.Now()
	for i := 0; i < 5; i++ {
		rl.Record(RunEntry{Job: "backup", StartedAt: now, Status: "ok"})
	}

	handler := PaginatedRunLogHandler(rl)
	r := httptest.NewRequest(http.MethodGet, "/?job=backup&page=1&page_size=3", nil)
	w := httptest.NewRecorder()
	handler(w, r)

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}

	var resp PageResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.TotalItems != 5 {
		t.Fatalf("expected 5 total items, got %d", resp.TotalItems)
	}
	if resp.PageSize != 3 {
		t.Fatalf("expected page_size 3, got %d", resp.PageSize)
	}
}
