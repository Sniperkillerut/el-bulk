package external

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
		Name      string `json:"name"`
		TypeLine  string `json:"type_line"`
		OracleText string `json:"oracle_text"`
		Artist     string `json:"artist"`
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
	
	// MTG Metadata
	Lang          string   `json:"lang"`
	ColorIdentity []string `json:"color_identity"`
	Rarity        string   `json:"rarity"`
	CMC           float64  `json:"cmc"`
	TypeLine      string   `json:"type_line"`
	OracleText    string   `json:"oracle_text"`
	Artist        string   `json:"artist"`
	Variation     bool     `json:"variation"`
	BorderColor   string   `json:"border_color"`
	Frame         string   `json:"frame"`
	FullArt       bool     `json:"full_art"`
	Textless      bool     `json:"textless"`
}

type CardIdentifier struct {
	Name            string `json:"name,omitempty"`
	Set             string `json:"set,omitempty"`
	CollectorNumber string `json:"collector_number,omitempty"`
}

type scryfallCollectionRequest struct {
	Identifiers []CardIdentifier `json:"identifiers"`
}

type scryfallCollectionResponse struct {
	Data []scryfallCard `json:"data"`
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

// LookupMTGCard queries Scryfall for an MTG card with multiple fallbacks:
// 1. By set code and collector number (exact match)
// 2. By name and set code (exact)
// 3. By name and set code (fuzzy)
// 4. By name only (fuzzy)
func LookupMTGCard(name, setCode, collectorNumber, foilTreatment string) (*CardLookupResult, error) {
	if name == "" && (setCode == "" || collectorNumber == "") {
		return nil, errors.New("card name or set/collector number is required")
	}

	// Step 1: Exact Set + Collector Number
	if setCode != "" && collectorNumber != "" {
		res, err := scryfallGet(fmt.Sprintf("%s/cards/%s/%s", scryfallBase, url.PathEscape(setCode), url.PathEscape(collectorNumber)), foilTreatment)
		if err == nil {
			return res, nil
		}
	}

	// Step 2: Named Exact + Set
	if name != "" && setCode != "" {
		params := url.Values{}
		params.Set("exact", name)
		params.Set("set", setCode)
		res, err := scryfallGet(fmt.Sprintf("%s/cards/named?%s", scryfallBase, params.Encode()), foilTreatment)
		if err == nil {
			return res, nil
		}
	}

	// Step 3: Named Fuzzy + Set
	if name != "" && setCode != "" {
		params := url.Values{}
		params.Set("fuzzy", name)
		params.Set("set", setCode)
		res, err := scryfallGet(fmt.Sprintf("%s/cards/named?%s", scryfallBase, params.Encode()), foilTreatment)
		if err == nil {
			return res, nil
		}
	}

	// Step 4: Named Fuzzy (Global fallback)
	if name != "" {
		params := url.Values{}
		params.Set("fuzzy", name)
		return scryfallGet(fmt.Sprintf("%s/cards/named?%s", scryfallBase, params.Encode()), foilTreatment)
	}

	return nil, errors.New("card not found")
}

// BatchLookupMTGCard fetches multiple cards from Scryfall using the collection endpoint.
func BatchLookupMTGCard(identifiers []CardIdentifier) ([]CardLookupResult, error) {
	if len(identifiers) == 0 {
		return nil, nil
	}

	reqBody, _ := json.Marshal(scryfallCollectionRequest{Identifiers: identifiers})
	req, err := http.NewRequest(http.MethodPost, scryfallBase+"/cards/collection", strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ElBulkTCGStore/1.0 (contact@elbulk.com)")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := scryfallClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("scryfall batch status %d", resp.StatusCode)
	}

	var batchResp scryfallCollectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&batchResp); err != nil {
		return nil, err
	}

	results := make([]CardLookupResult, len(batchResp.Data))
	for i, card := range batchResp.Data {
		// Note: foil_treatment is tricky in batch, as it's per card.
		// For now, we return all metadata and the caller can pick prices later.
		results[i] = *mapScryfallToResult(&card, "non_foil")
	}

	return results, nil
}

func scryfallGet(reqURL string, foilTreatment string) (*CardLookupResult, error) {
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ElBulkTCGStore/1.0 (contact@elbulk.com)")
	req.Header.Set("Accept", "application/json")

	time.Sleep(100 * time.Millisecond) // Respect rate limits

	resp, err := scryfallClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, errors.New("card not found")
		}
		return nil, fmt.Errorf("scryfall status %d", resp.StatusCode)
	}

	var card scryfallCard
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, err
	}

	return mapScryfallToResult(&card, foilTreatment), nil
}

func mapScryfallToResult(card *scryfallCard, foilTreatment string) *CardLookupResult {
	imageURL := card.bestImageURL()
	tcgUSD, cmEUR := card.scryfallPrices(foilTreatment)

	var colorStr *string
	if len(card.ColorIdentity) > 0 {
		cs := strings.Join(card.ColorIdentity, ",")
		colorStr = &cs
	}

	lowerType := strings.ToLower(card.TypeLine)
	isLegendary := strings.Contains(lowerType, "legendary")
	isHistoric := isLegendary || strings.Contains(lowerType, "artifact") || strings.Contains(lowerType, "saga")
	
	var artVar *string
	if card.Variation {
		v := "Variation"
		artVar = &v
	}

	oracle := card.OracleText
	artist := card.Artist
	typeLine := card.TypeLine

	if oracle == "" && len(card.CardFaces) > 0 {
		var oParts, aParts, tParts []string
		for _, f := range card.CardFaces {
			if f.OracleText != "" {
				oParts = append(oParts, f.OracleText)
			}
			if f.Artist != "" {
				aParts = append(aParts, f.Artist)
			}
			if f.TypeLine != "" {
				tParts = append(tParts, f.TypeLine)
			}
		}
		oracle = strings.Join(oParts, "\n\n---\n\n")
		artist = strings.Join(aParts, " & ")
		typeLine = strings.Join(tParts, " // ")
	}

	return &CardLookupResult{
		ImageURL:        imageURL,
		SetName:         card.SetName,
		SetCode:         card.Set,
		CollectorNumber: card.CollectorNumber,
		PriceTCGPlayer:  tcgUSD,
		PriceCardmarket: cmEUR,
		Language:        card.Lang,
		ColorIdentity:   colorStr,
		Rarity:          &card.Rarity,
		CMC:             &card.CMC,
		IsLegendary:     isLegendary,
		IsHistoric:      isHistoric,
		IsLand:          strings.Contains(lowerType, "land"),
		IsBasicLand:     strings.Contains(lowerType, "basic land"),
		ArtVariation:    artVar,
		OracleText:      &oracle,
		Artist:          &artist,
		TypeLine:        &typeLine,
		BorderColor:     &card.BorderColor,
		Frame:           &card.Frame,
		FullArt:         card.FullArt,
		Textless:        card.Textless,
	}
}
