package external

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchExchangeRates(t *testing.T) {
	// 1. Mock the API response
	mockResponse := `{
		"result": "success",
		"base_code": "USD",
		"conversion_rates": {
			"USD": 1,
			"COP": 4250.50,
			"EUR": 0.92
		},
		"time_last_update_unix": 1700000000
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/test-api-key/latest/USD")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, mockResponse)
	}))
	defer server.Close()

	// 2. Setup environment and override URL
	origKey := os.Getenv("EXCHANGERATE_API_KEY")
	origURL := currencyBaseURL
	defer func() {
		os.Setenv("EXCHANGERATE_API_KEY", origKey)
		currencyBaseURL = origURL
	}()

	os.Setenv("EXCHANGERATE_API_KEY", "test-api-key")
	currencyBaseURL = server.URL

	// 3. Execute
	rates, err := FetchExchangeRates(context.Background())

	// 4. Verify
	assert.NoError(t, err)
	assert.NotNil(t, rates)
	assert.Equal(t, 4250.50, rates["COP"])
	assert.Equal(t, 0.92, rates["EUR"])
	assert.Equal(t, 1.0, rates["USD"])
}

func TestFetchExchangeRates_NoKey(t *testing.T) {
	origKey := os.Getenv("EXCHANGERATE_API_KEY")
	os.Unsetenv("EXCHANGERATE_API_KEY")
	defer os.Setenv("EXCHANGERATE_API_KEY", origKey)

	rates, err := FetchExchangeRates(context.Background())
	assert.Error(t, err)
	assert.Nil(t, rates)
	assert.Contains(t, err.Error(), "EXCHANGERATE_API_KEY environment variable is not set")
}

func TestFetchExchangeRates_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	origKey := os.Getenv("EXCHANGERATE_API_KEY")
	origURL := currencyBaseURL
	defer func() {
		os.Setenv("EXCHANGERATE_API_KEY", origKey)
		currencyBaseURL = origURL
	}()

	os.Setenv("EXCHANGERATE_API_KEY", "invalid-key")
	currencyBaseURL = server.URL

	rates, err := FetchExchangeRates(context.Background())
	assert.Error(t, err)
	assert.Nil(t, rates)
	assert.Contains(t, err.Error(), "exchange rate api returned status: 401")
}
