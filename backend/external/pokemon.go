package external

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

const pokemonTCGBase = "https://api.pokemontcg.io/v2"

var pokemonClient = &http.Client{Timeout: 10 * time.Second}

// pokemonAPICard is the minimal subset of the Pokémon TCG API card object.
type pokemonAPICard struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Number string `json:"number"`
	Set    struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"set"`
	Images struct {
		Small string `json:"small"`
		Large string `json:"large"`
	} `json:"images"`
}

type pokemonAPIResponse struct {
	Data       []pokemonAPICard `json:"data"`
	TotalCount int              `json:"totalCount"`
}

// LookupPokemonCard queries the Pokémon TCG API for a card by name and optional set ID.
// Set IDs follow the API format, e.g. "base1", "swsh12", "sv3pt5".
// If POKEMON_TCG_API_KEY env var is set it will be sent as X-Api-Key for higher rate limits.
// The best match is the first result from the API (sorted by set release, newest first by default).
func LookupPokemonCard(name, setID string) (*CardLookupResult, error) {
	if name == "" {
		return nil, errors.New("card name is required")
	}

	// Build the search query: `name:"Charizard" set.id:base1`
	q := fmt.Sprintf(`name:"%s"`, name)
	if setID != "" {
		q += fmt.Sprintf(` set.id:%s`, setID)
	}

	params := url.Values{}
	params.Set("q", q)
	params.Set("pageSize", "1")
	params.Set("orderBy", "-set.releaseDate") // newest set first
	reqURL := fmt.Sprintf("%s/cards?%s", pokemonTCGBase, params.Encode())

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building pokemon request: %w", err)
	}

	req.Header.Set("User-Agent", "ElBulkTCGStore/1.0")
	req.Header.Set("Accept", "application/json")

	// Optional API key for higher rate limits
	if key := os.Getenv("POKEMON_TCG_API_KEY"); key != "" {
		req.Header.Set("X-Api-Key", key)
	}

	resp, err := pokemonClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pokemon TCG API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, errors.New("pokémon TCG API rate limit exceeded; add POKEMON_TCG_API_KEY to .env for higher limits")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pokémon TCG API returned status %d", resp.StatusCode)
	}

	var apiResp pokemonAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding pokémon TCG API response: %w", err)
	}

	if len(apiResp.Data) == 0 {
		return nil, errors.New("card not found")
	}

	card := apiResp.Data[0]
	imageURL := card.Images.Large
	if imageURL == "" {
		imageURL = card.Images.Small
	}
	if imageURL == "" {
		return nil, errors.New("pokémon TCG API returned card with no image")
	}

	return &CardLookupResult{
		ImageURL:        imageURL,
		SetName:         card.Set.Name,
		SetCode:         card.Set.ID,
		CollectorNumber: card.Number,
	}, nil
}
