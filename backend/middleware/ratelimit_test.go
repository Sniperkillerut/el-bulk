package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimit(t *testing.T) {
	// For testing, we use a small limit and a long window
	limit := 3
	window := time.Second
	handler := RateLimit(limit, window)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	testHandler := handler(nextHandler)

	// Clean up previous test state (if any)
	mu.Lock()
	clients = make(map[string]*client)
	mu.Unlock()

	// 1. Success cases
	for i := 0; i < limit; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	}

	// 2. Exceed limit
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)

	// 3. Different IP should still work
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "5.6.7.8:5678"
	rr2 := httptest.NewRecorder()
	testHandler.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)

	// 4. Wait for window to expire
	time.Sleep(window + 100*time.Millisecond)
	req3 := httptest.NewRequest("GET", "/", nil)
	req3.RemoteAddr = "1.2.3.4:1234"
	rr3 := httptest.NewRecorder()
	testHandler.ServeHTTP(rr3, req3)
	assert.Equal(t, http.StatusOK, rr3.Code)
}
