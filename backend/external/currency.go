package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// currencyClient is used for external exchange rate API calls.
var currencyClient = &http.Client{Timeout: 15 * time.Second}

// ExchangeRateResponse is the standard response structure for ExchangeRate-API v6.
type ExchangeRateResponse struct {
	Result             string             `json:"result"`
	BaseCode           string             `json:"base_code"`
	ConversionRates    map[string]float64 `json:"conversion_rates"`
	TimeLastUpdateUnix int64              `json:"time_last_update_unix"`
}

var currencyBaseURL = "https://v6.exchangerate-api.com/v6"

// FetchExchangeRates fetches the latest conversion rates from ExchangeRate-API.
// It requires EXCHANGERATE_API_KEY to be set in the environment.
func FetchExchangeRates(ctx context.Context) (map[string]float64, error) {
	apiKey := os.Getenv("EXCHANGERATE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("EXCHANGERATE_API_KEY environment variable is not set")
	}

	// We use USD as the base for the free tier compatibility
	url := fmt.Sprintf("%s/%s/latest/USD", currencyBaseURL, apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := currencyClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("exchange rate api returned status: %d", resp.StatusCode)
	}

	var data ExchangeRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode exchange rate response: %w", err)
	}

	if data.Result != "success" {
		return nil, fmt.Errorf("exchange rate api returned result error: %s", data.Result)
	}

	return data.ConversionRates, nil
}
