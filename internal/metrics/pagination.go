package metrics

import (
	"net/http"
	"strconv"
)

// PageRequest holds parsed pagination parameters.
type PageRequest struct {
	Page     int
	PageSize int
}

// PageResponse wraps a paginated result set.
type PageResponse struct {
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalItems int         `json:"total_items"`
	TotalPages int         `json:"total_pages"`
	Items      interface{} `json:"items"`
}

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 200
)

// ParsePageRequest extracts page and page_size query parameters from the
// request, applying defaults and clamping page_size to maxPageSize.
func ParsePageRequest(r *http.Request) PageRequest {
	page := queryInt(r, "page", defaultPage)
	size := queryInt(r, "page_size", defaultPageSize)

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = defaultPageSize
	}
	if size > maxPageSize {
		size = maxPageSize
	}
	return PageRequest{Page: page, PageSize: size}
}

// Paginate slices a generic slice according to the PageRequest and returns
// a PageResponse. items must be a []T for any T.
func Paginate[T any](items []T, req PageRequest) PageResponse {
	total := len(items)
	totalPages := (total + req.PageSize - 1) / req.PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	start := (req.Page - 1) * req.PageSize
	if start > total {
		start = total
	}
	end := start + req.PageSize
	if end > total {
		end = total
	}

	return PageResponse{
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalItems: total,
		TotalPages: totalPages,
		Items:      items[start:end],
	}
}

func queryInt(r *http.Request, key string, def int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
