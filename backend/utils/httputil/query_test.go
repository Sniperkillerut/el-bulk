package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetIntParam(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		key          string
		defaultValue int
		want         int
	}{
		{
			name:         "valid integer",
			url:          "/?age=25",
			key:          "age",
			defaultValue: 10,
			want:         25,
		},
		{
			name:         "invalid integer",
			url:          "/?age=abc",
			key:          "age",
			defaultValue: 10,
			want:         10,
		},
		{
			name:         "missing parameter",
			url:          "/",
			key:          "age",
			defaultValue: 10,
			want:         10,
		},
		{
			name:         "empty parameter",
			url:          "/?age=",
			key:          "age",
			defaultValue: 10,
			want:         10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tt.url, nil)
			got := GetIntParam(r, tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetIntParam() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPagination(t *testing.T) {
	tests := []struct {
		name            string
		url             string
		defaultPageSize int
		maxPageSize     int
		wantPage        int
		wantPageSize    int
		wantOffset      int
	}{
		{
			name:            "valid pagination",
			url:             "/?page=2&page_size=20",
			defaultPageSize: 10,
			maxPageSize:     50,
			wantPage:        2,
			wantPageSize:    20,
			wantOffset:      20,
		},
		{
			name:            "defaults",
			url:             "/",
			defaultPageSize: 10,
			maxPageSize:     50,
			wantPage:        1,
			wantPageSize:    10,
			wantOffset:      0,
		},
		{
			name:            "page below 1",
			url:             "/?page=0",
			defaultPageSize: 10,
			maxPageSize:     50,
			wantPage:        1,
			wantPageSize:    10,
			wantOffset:      0,
		},
		{
			name:            "page_size below 1",
			url:             "/?page_size=0",
			defaultPageSize: 10,
			maxPageSize:     50,
			wantPage:        1,
			wantPageSize:    10,
			wantOffset:      0,
		},
		{
			name:            "page_size exceeds max",
			url:             "/?page_size=100",
			defaultPageSize: 10,
			maxPageSize:     50,
			wantPage:        1,
			wantPageSize:    50,
			wantOffset:      0,
		},
		{
			name:            "non-integer values",
			url:             "/?page=abc&page_size=def",
			defaultPageSize: 10,
			maxPageSize:     50,
			wantPage:        1,
			wantPageSize:    10,
			wantOffset:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tt.url, nil)
			page, pageSize, offset := GetPagination(r, tt.defaultPageSize, tt.maxPageSize)
			if page != tt.wantPage {
				t.Errorf("GetPagination() page = %v, want %v", page, tt.wantPage)
			}
			if pageSize != tt.wantPageSize {
				t.Errorf("GetPagination() pageSize = %v, want %v", pageSize, tt.wantPageSize)
			}
			if offset != tt.wantOffset {
				t.Errorf("GetPagination() offset = %v, want %v", offset, tt.wantOffset)
			}
		})
	}
}

func TestGetStringParam(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		key          string
		defaultValue string
		want         string
	}{
		{
			name:         "present parameter",
			url:          "/?name=John",
			key:          "name",
			defaultValue: "Guest",
			want:         "John",
		},
		{
			name:         "missing parameter",
			url:          "/",
			key:          "name",
			defaultValue: "Guest",
			want:         "Guest",
		},
		{
			name:         "empty parameter",
			url:          "/?name=",
			key:          "name",
			defaultValue: "Guest",
			want:         "Guest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tt.url, nil)
			got := GetStringParam(r, tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetStringParam() = %v, want %v", got, tt.want)
			}
		})
	}
}
