package httputil

import (
    "net/http"
    "strconv"
)

// GetIntParam gets an integer parameter from the query string with a default value.
func GetIntParam(r *http.Request, key string, defaultValue int) int {
    v := r.URL.Query().Get(key)
    if v == "" {
        return defaultValue
    }
    i, err := strconv.Atoi(v)
    if err != nil {
        return defaultValue
    }
    return i
}

// GetPagination returns (page, pageSize, offset) normalized from query parameters.
func GetPagination(r *http.Request, defaultPageSize, maxPageSize int) (int, int, int) {
    page := GetIntParam(r, "page", 1)
    if page < 1 {
        page = 1
    }

    pageSize := GetIntParam(r, "page_size", defaultPageSize)
    if pageSize < 1 {
        pageSize = defaultPageSize
    }
    if pageSize > maxPageSize {
        pageSize = maxPageSize
    }

    offset := (page - 1) * pageSize
    return page, pageSize, offset
}

// GetStringParam gets a string parameter from the query string with a default value.
func GetStringParam(r *http.Request, key, defaultValue string) string {
    v := r.URL.Query().Get(key)
    if v == "" {
        return defaultValue
    }
    return v
}
