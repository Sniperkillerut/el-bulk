package external

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const scryfallBase = "https://api.scryfall.com"

// scryfallClient is a shared HTTP client with a reasonable timeout.
var scryfallClient = &http.Client{Timeout: 10 * time.Second}

// scryfallCard is the minimal subset of the Scryfall card object we need.
type scryfallCard struct {
	Name            string `json:"name"`
	CollectorNumber string `json:"collector_number"`
	SetName         string `json:"set_name"`
	Set             string `json:"set"` // lowercase set code, e.g. "m11"
	ImageURIs       struct {
		Normal string `json:"normal"`
		Large  string `json:"large"`
		PNG    string `json:"png"`
	} `json:"image_uris"`
	// Two-faced cards store images per face
	CardFaces []struct {
		ImageURIs struct {
			Normal string `json:"normal"`
			Large  string `json:"large"`
			PNG    string `json:"png"`
		} `json:"image_uris"`
	} `json:"card_faces"`
	// Prices from marketplace aggregators (all values are strings or null in JSON)
	Prices struct {
		USD       *string `json:"usd"`        // TCGPlayer non-foil USD
		USDFoil   *string `json:"usd_foil"`   // TCGPlayer foil USD
		USDEtched *string `json:"usd_etched"` // TCGPlayer etched foil USD
		EUR       *string `json:"eur"`        // Cardmarket non-foil EUR
		EURFoil   *string `json:"eur_foil"`   // Cardmarket foil EUR
	} `json:"prices"`
}

func (c *scryfallCard) bestImageURL() string {
	for _, u := range []string{c.ImageURIs.PNG, c.ImageURIs.Large, c.ImageURIs.Normal} {
		if u != "" {
			return u
		}
	}
	if len(c.CardFaces) > 0 {
		face := c.CardFaces[0].ImageURIs
		for _, u := range []string{face.PNG, face.Large, face.Normal} {
			if u != "" {
				return u
			}
		}
	}
	return ""
}

// parsePrice converts a nullable Scryfall price string to *float64.
func parsePrice(s *string) *float64 {
	if s == nil || *s == "" {
		return nil
	}
	v, err := strconv.ParseFloat(*s, 64)
	if err != nil {
		return nil
	}
	return &v
}

// scryfallPrices extracts TCGPlayer and Cardmarket prices appropriate for the
// given foil treatment. foilTreatment should be one of the foil_treatment_type values.
func (c *scryfallCard) scryfallPrices(foilTreatment string) (tcgUSD, cmEUR *float64) {
	switch foilTreatment {
	case "etched_foil":
		tcgUSD = parsePrice(c.Prices.USDEtched)
		cmEUR = parsePrice(c.Prices.EURFoil) // Cardmarket doesn't distinguish etched
	case "foil", "holo_foil", "ripple_foil", "galaxy_foil", "platinum_foil":
		tcgUSD = parsePrice(c.Prices.USDFoil)
		cmEUR = parsePrice(c.Prices.EURFoil)
	default: // non_foil
		tcgUSD = parsePrice(c.Prices.USD)
		cmEUR = parsePrice(c.Prices.EUR)
	}
	return
}

// LookupMTGCard queries Scryfall for an MTG card by name and optional set code.
// foilTreatment is used to select the correct price variant (pass empty string for non-foil default).
func LookupMTGCard(name, setCode, foilTreatment string) (*CardLookupResult, error) {
	if name == "" {
		return nil, errors.New("card name is required")
	}

	params := url.Values{}
	params.Set("fuzzy", name)
	if setCode != "" {
		params.Set("set", setCode)
	}
	reqURL := fmt.Sprintf("%s/cards/named?%s", scryfallBase, params.Encode())

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building scryfall request: %w", err)
	}
	req.Header.Set("User-Agent", "ElBulkTCGStore/1.0 (contact@elbulk.com)")
	req.Header.Set("Accept", "application/json")

	// Scryfall asks for 50–100ms between requests
	time.Sleep(100 * time.Millisecond)

	resp, err := scryfallClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("scryfall request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("card not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("scryfall returned status %d", resp.StatusCode)
	}

	var card scryfallCard
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, fmt.Errorf("decoding scryfall response: %w", err)
	}

	imageURL := card.bestImageURL()
	if imageURL == "" {
		return nil, errors.New("scryfall returned card with no image")
	}

	tcgUSD, cmEUR := card.scryfallPrices(foilTreatment)

	return &CardLookupResult{
		ImageURL:        imageURL,
		SetName:         card.SetName,
		SetCode:         card.Set,
		CollectorNumber: card.CollectorNumber,
		PriceTCGPlayer:  tcgUSD,
		PriceCardmarket: cmEUR,
	}, nil
}
