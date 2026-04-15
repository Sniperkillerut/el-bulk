package external

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
	"sync"
	"bytes"
)

var ScryfallBase = "https://api.scryfall.com"

// scryfallClient is a shared HTTP client with a reasonable timeout.
var scryfallClient = &http.Client{Timeout: 30 * time.Second}

// scryfallCard is the minimal subset of the Scryfall card object we need.
type scryfallCard struct {
	ID              string `json:"id"`
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
	PromoTypes    []string `json:"promo_types"`
	Finishes      []string `json:"finishes"`
	FrameEffects  []string `json:"frame_effects"`
	Legalities    map[string]string `json:"legalities"`
	CardKingdomID     *string `json:"card_kingdom_id"`
	CardKingdomFoilID *string `json:"card_kingdom_foil_id"`
}

type CardIdentifier struct {
	ScryfallID      string `json:"id,omitempty"`
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
// 0. By Scryfall ID (direct)
// 1. By set code and collector number (exact match)
// 2. By name and set code (exact)
// 3. By name and set code (fuzzy)
// 4. By name only (fuzzy)
func LookupMTGCard(ctx context.Context, scryfallID, name, setCode, collectorNumber, foilTreatment string) (*CardLookupResult, error) {
	if scryfallID == "" && name == "" && (setCode == "" || collectorNumber == "") {
		return nil, errors.New("scryfall id, card name or set/collector number is required")
	}

	// Step 0: Direct Scryfall ID Lookup
	if scryfallID != "" {
		res, err := scryfallGet(ctx, fmt.Sprintf("%s/cards/%s", ScryfallBase, url.PathEscape(scryfallID)), foilTreatment)
		if err == nil {
			return res, nil
		}
	}

	// Step 1: Exact Set + Collector Number
	if setCode != "" && collectorNumber != "" {
		res, err := scryfallGet(ctx, fmt.Sprintf("%s/cards/%s/%s", ScryfallBase, url.PathEscape(setCode), url.PathEscape(collectorNumber)), foilTreatment)
		if err == nil {
			return res, nil
		}
	}

	// Step 2: Named Exact + Set
	if name != "" && setCode != "" {
		params := url.Values{}
		params.Set("exact", name)
		params.Set("set", setCode)
		res, err := scryfallGet(ctx, fmt.Sprintf("%s/cards/named?%s", ScryfallBase, params.Encode()), foilTreatment)
		if err == nil {
			return res, nil
		}
	}

	// Step 3: Named Fuzzy + Set
	if name != "" && setCode != "" {
		params := url.Values{}
		params.Set("fuzzy", name)
		params.Set("set", setCode)
		res, err := scryfallGet(ctx, fmt.Sprintf("%s/cards/named?%s", ScryfallBase, params.Encode()), foilTreatment)
		if err == nil {
			return res, nil
		}
	}

	// Step 4: Named Fuzzy (Global fallback)
	if name != "" {
		params := url.Values{}
		params.Set("fuzzy", name)
		return scryfallGet(ctx, fmt.Sprintf("%s/cards/named?%s", ScryfallBase, params.Encode()), foilTreatment)
	}

	return nil, errors.New("card not found")
}

// BatchLookupMTGCard fetches multiple cards from Scryfall using the collection endpoint.
func BatchLookupMTGCard(ctx context.Context, identifiers []CardIdentifier) ([]CardLookupResult, error) {
	if len(identifiers) == 0 {
		return nil, nil
	}

	reqBody, _ := json.Marshal(scryfallCollectionRequest{Identifiers: identifiers})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ScryfallBase+"/cards/collection", strings.NewReader(string(reqBody)))
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

func scryfallGet(ctx context.Context, reqURL string, foilTreatment string) (*CardLookupResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
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
	} else {
		colorless := "C"
		colorStr = &colorless
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

	var promoType *string
	if len(card.PromoTypes) > 0 {
		pt := strings.Join(card.PromoTypes, ", ")
		promoType = &pt
	}

	return &CardLookupResult{
		Name:            card.Name,
		ImageURL:        imageURL,
		PriceTCGPlayer:  tcgUSD,
		PriceCardmarket: cmEUR,
		MTGMetadata: models.MTGMetadata{
			SetName:         &card.SetName,
			SetCode:         &card.Set,
			CollectorNumber: &card.CollectorNumber,
			FoilTreatment:   resolveFoilTreatment(card, foilTreatment),
			CardTreatment:   resolveCardTreatment(card),
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
			PromoType:       promoType,
			ScryfallID:      &card.ID,
			Legalities:      castLegalities(card.Legalities),
		},
	}
}

func castLegalities(m map[string]string) models.JSONB {
	if m == nil {
		return nil
	}
	res := make(models.JSONB)
	for k, v := range m {
		res[k] = v
	}
	return res
}

func resolveCardTreatment(card *scryfallCard) models.CardTreatment {
	for _, effect := range card.FrameEffects {
		switch effect {
		case "showcase":
			return models.TreatmentShowcase
		case "extendedart":
			return models.TreatmentExtendedArt
		case "etched":
			return models.TreatmentEtched
		}
	}
	if card.BorderColor == "borderless" {
		return models.TreatmentBorderless
	}
	if card.FullArt {
		return models.TreatmentFullArt
	}
	if card.Textless {
		return models.TreatmentTextless
	}
	return models.TreatmentNormal
}

func resolveFoilTreatment(card *scryfallCard, requestedFoil string) models.FoilTreatment {
	hasFoil := false
	hasNonFoil := false
	hasEtched := false
	hasGlossy := false

	for _, f := range card.Finishes {
		switch f {
		case "foil":
			hasFoil = true
		case "nonfoil":
			hasNonFoil = true
		case "etched":
			hasEtched = true
		case "glossy":
			hasGlossy = true
		}
	}

	// 1. Check for specialized finishes in promo_types if the card supports foil
	if hasFoil {
		for _, pt := range card.PromoTypes {
			switch pt {
			case "ripplefoil":
				return models.FoilRippleFoil
			case "surgefoil":
				return models.FoilSurgeFoil
			case "confettifoil":
				return models.FoilConfettiFoil
			case "textured":
				return models.FoilTexturedFoil
			case "stepandcompleat":
				return models.FoilStepAndCompleat
			case "oilSlick":
				return models.FoilOilSlick
			case "neonink":
				return models.FoilNeonInk
			case "galaxyfoil":
				return models.FoilGalaxyFoil
			}
		}
	}

	// Double-check requested foil from caller (e.g. from CSV parser)
	if requestedFoil != "" {
		rf := models.FoilTreatment(requestedFoil)
		// If they explicitly asked for a specialized foil that matches our system,
		// and the card has the base finish (usually "foil"), return it.
		if hasFoil && (rf == models.FoilRippleFoil || rf == models.FoilSurgeFoil ||
			rf == models.FoilConfettiFoil || rf == models.FoilGalaxyFoil ||
			rf == models.FoilTexturedFoil || rf == models.FoilStepAndCompleat ||
			rf == models.FoilOilSlick || rf == models.FoilNeonInk) {
			return rf
		}

		if (rf == models.FoilFoil && hasFoil) || (rf == models.FoilNonFoil && hasNonFoil) {
			return rf
		}
	}

	// Standard fallbacks
	if hasEtched {
		return models.FoilEtchedFoil
	}
	if hasGlossy {
		return models.FoilGalaxyFoil
	}

	if hasNonFoil {
		return models.FoilNonFoil
	}
	if hasFoil {
		return models.FoilFoil
	}

	return models.FoilNonFoil
}

// ─── Scryfall bulk data structures ──────────────────────────────────────────

// ScryfallBulkMeta is the top-level response from GET /bulk-data.
type ScryfallBulkMeta struct {
	Data []struct {
		Type        string `json:"type"`
		DownloadURI string `json:"download_uri"`
		UpdatedAt   string `json:"updated_at"`
	} `json:"data"`
}

// ScryfallBulkCard is the price-relevant subset of each card in the bulk file.
type ScryfallBulkCard struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Set             string            `json:"set"` // set code, e.g. "m11"
	CollectorNumber string            `json:"collector_number"`
	TypeLine        string            `json:"type_line"`
	OracleText string            `json:"oracle_text"`
	Legalities map[string]string `json:"legalities"`
	ImageURIs  struct {
		Normal string `json:"normal"`
	} `json:"image_uris"`
	Prices struct {
		USD       *string `json:"usd"`
		USDFoil   *string `json:"usd_foil"`
		USDEtched *string `json:"usd_etched"`
		EUR       *string `json:"eur"`
		EURFoil   *string `json:"eur_foil"`
	} `json:"prices"`
	CardKingdomID     *string `json:"card_kingdom_id"`
	CardKingdomFoilID *string `json:"card_kingdom_foil_id"`
}

// PriceKey uniquely identifies a card+foil combination for the in-memory map.
type PriceKey struct {
	Name      string // lowercase
	SetCode   string // lowercase; empty = any set
	Collector string // NEW: collector number
	Foil      string // foil_treatment value
}

// CardMetadata holds extracted USD/EUR prices and MTG card data for one PriceKey.
type CardMetadata struct {
	// Non-foil prices (TCGPlayer / Cardmarket)
	TCGPlayerUSD  *float64
	CardmarketEUR *float64
	// Foil prices — populated by BatchLookupMTG; callers should pick based on FoilTreatment
	TCGPlayerUSDFoil  *float64
	CardmarketEURFoil *float64
	Legalities     models.JSONB
	OracleText     string
	ScryfallID     string
	TypeLine       string
	ImageURL       string
	CardKingdomID  string
	// Foil CK ID (card_kingdom_foil_id from Scryfall)
	CardKingdomFoilID string
}

var (
	scryCache      map[PriceKey]CardMetadata
	scryCacheMutex sync.RWMutex
	scryCacheTime  time.Time
)

// BuildPriceMap downloads Scryfall's "default_cards" bulk file and
// builds a lookup map. Uses an in-memory cache valid for 1 hour.
func BuildPriceMap(ctx context.Context) (map[PriceKey]CardMetadata, error) {
	scryCacheMutex.RLock()
	if scryCache != nil && time.Since(scryCacheTime) < 1*time.Hour {
		cache := scryCache
		scryCacheMutex.RUnlock()
		return cache, nil
	}
	scryCacheMutex.RUnlock()

	scryCacheMutex.Lock()
	defer scryCacheMutex.Unlock()

	if scryCache != nil && time.Since(scryCacheTime) < 1*time.Hour {
		return scryCache, nil
	}

	logger.InfoCtx(ctx, "Downloading Scryfall bulk data (cache empty or expired)...")
	client := &http.Client{Timeout: 5 * time.Minute} 

	// Step 1: discover today's bulk-data download URL
	metaReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, ScryfallBase+"/bulk-data", nil)
	metaReq.Header.Set("User-Agent", "ElBulkTCGStore/1.0 (contact@elbulk.com)")
	metaReq.Header.Set("Accept", "application/json")

	metaResp, err := client.Do(metaReq)
	if err != nil {
		return nil, fmt.Errorf("fetching bulk-data index: %w", err)
	}
	defer metaResp.Body.Close()

	var meta ScryfallBulkMeta
	if err := json.NewDecoder(metaResp.Body).Decode(&meta); err != nil {
		return nil, fmt.Errorf("decoding bulk-data index: %w", err)
	}

	downloadURL := ""
	for _, item := range meta.Data {
		if item.Type == "default_cards" {
			downloadURL = item.DownloadURI
			break
		}
	}
	if downloadURL == "" {
		return nil, fmt.Errorf("default_cards bulk file not found in scryfall response")
	}

	// Step 2: stream the bulk card JSON array
	dlReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	dlReq.Header.Set("User-Agent", "ElBulkTCGStore/1.0 (contact@elbulk.com)")

	dlResp, err := client.Do(dlReq)
	if err != nil {
		return nil, fmt.Errorf("downloading bulk data: %w", err)
	}
	defer dlResp.Body.Close()

	// Step 3: stream-decode the JSON array without loading 600MB into memory at once
	priceMap := make(map[PriceKey]CardMetadata, 300_000)
	decoder := json.NewDecoder(dlResp.Body)

	// Read opening '['
	if _, err := decoder.Token(); err != nil {
		return nil, fmt.Errorf("reading bulk JSON opening token: %w", err)
	}

	for decoder.More() {
		var card ScryfallBulkCard
		if err := decoder.Decode(&card); err != nil {
			// Skip malformed cards rather than failing the whole run
			continue
		}

		name := strings.ToLower(card.Name)
		set := strings.ToLower(card.Set)
		meta := CardMetadata{
			Legalities:  castLegalities(card.Legalities),
			OracleText:  card.OracleText,
			ScryfallID:  card.ID,
			TypeLine:    card.TypeLine,
			ImageURL:    card.ImageURIs.Normal,
		}

		// Register entries for each foil variant this print has prices for
		variants := []struct {
			foil string
			usd  *string
			eur  *string
		}{
			{"non_foil", card.Prices.USD, card.Prices.EUR},
			{"foil", card.Prices.USDFoil, card.Prices.EURFoil},
			{"holo_foil", card.Prices.USDFoil, card.Prices.EURFoil},
			{"ripple_foil", card.Prices.USDFoil, card.Prices.EURFoil},
			{"galaxy_foil", card.Prices.USDFoil, card.Prices.EURFoil},
			{"platinum_foil", card.Prices.USDFoil, card.Prices.EURFoil},
			{"etched_foil", card.Prices.USDEtched, card.Prices.EURFoil},
		}

		for _, v := range variants {
			tcg := parsePrice(v.usd)
			cm := parsePrice(v.eur)

			// Fallback: if variant-specific price is missing, use the non-foil price.
			// Scryfall often returns null for specialties (e.g. ripple foil) when 
			// the market only tracks one price.
			if tcg == nil && v.foil != "non_foil" {
				tcg = parsePrice(card.Prices.USD)
			}
			if cm == nil && v.foil != "non_foil" {
				cm = parsePrice(card.Prices.EUR)
			}

			if tcg == nil && cm == nil {
				continue
			}

			// Copy metadata and add specific prices
			entry := meta
			entry.TCGPlayerUSD = tcg
			entry.CardmarketEUR = cm

			// Add CK ID if applicable for this finish
			if v.foil == "non_foil" && card.CardKingdomID != nil {
				entry.CardKingdomID = *card.CardKingdomID
			} else if (v.foil == "foil" || v.foil == "holo_foil") && card.CardKingdomFoilID != nil {
				entry.CardKingdomID = *card.CardKingdomFoilID
			}

			// Index by specific set + collector number
			priceMap[PriceKey{Name: name, SetCode: set, Collector: card.CollectorNumber, Foil: v.foil}] = entry
			
			// Maintain previous set-only/name-only fallbacks
			priceMap[PriceKey{Name: name, SetCode: set, Collector: "", Foil: v.foil}] = entry
			priceMap[PriceKey{Name: name, SetCode: "", Collector: "", Foil: v.foil}] = entry
		}
	}

	scryCache = priceMap
	scryCacheTime = time.Now()

	logger.InfoCtx(ctx, "Parsed and cached %d Scryfall cards in price map", len(priceMap))
	return priceMap, nil
}
// ─── Scryfall sets data structures ──────────────────────────────────────────

type ScryfallSet struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	ReleasedAt string `json:"released_at"`
	SetType    string `json:"set_type"`
}

type ScryfallSetsResponse struct {
	Data []ScryfallSet `json:"data"`
}

// FetchSets retrieves all MTG sets from Scryfall.
func FetchSets(ctx context.Context) ([]ScryfallSet, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ScryfallBase+"/sets", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ElBulkTCGStore/1.0 (contact@elbulk.com)")
	req.Header.Set("Accept", "application/json")

	resp, err := scryfallClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("scryfall sets status %d", resp.StatusCode)
	}

	var setsResp ScryfallSetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&setsResp); err != nil {
		return nil, err
	}

	return setsResp.Data, nil
}

// BatchLookupMTG fetches metadata and prices for a batch of Scryfall IDs.
// Scryfall's /cards/collection endpoint accepts at most 75 identifiers per
// request; this function transparently chunks the input and merges results.
func BatchLookupMTG(ctx context.Context, scryfallIDs []string) (map[string]CardMetadata, error) {
	if len(scryfallIDs) == 0 {
		return nil, nil
	}

	type identifier struct {
		ID string `json:"id"`
	}
	type collectionReq struct {
		Identifiers []identifier `json:"identifiers"`
	}

	const pageSize = 75
	results := make(map[string]CardMetadata, len(scryfallIDs))

	for start := 0; start < len(scryfallIDs); start += pageSize {
		end := start + pageSize
		if end > len(scryfallIDs) {
			end = len(scryfallIDs)
		}
		page := scryfallIDs[start:end]

		ids := make([]identifier, len(page))
		for i, id := range page {
			ids[i] = identifier{ID: id}
		}

		body, _ := json.Marshal(collectionReq{Identifiers: ids})
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, ScryfallBase+"/cards/collection", bytes.NewReader(body))
		req.Header.Set("User-Agent", "ElBulkTCGStore/1.0 (contact@elbulk.com)")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := scryfallClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("scryfall collection page %d: %w", start/pageSize+1, err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("scryfall collection status %d (page %d)", resp.StatusCode, start/pageSize+1)
		}

		var scryResp struct {
			Data []scryfallCard `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&scryResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("scryfall collection decode page %d: %w", start/pageSize+1, err)
		}
		resp.Body.Close()

		for _, c := range scryResp.Data {
			usd := parsePrice(c.Prices.USD)
			eur := parsePrice(c.Prices.EUR)
			usdFoil := parsePrice(c.Prices.USDFoil)
			eurFoil := parsePrice(c.Prices.EURFoil)

			// Fallback: If foil price is missing, use non-foil as default.
			if usdFoil == nil {
				usdFoil = usd
			}
			if eurFoil == nil {
				eurFoil = eur
			}

			meta := CardMetadata{
				ScryfallID:        c.ID,
				TCGPlayerUSD:      usd,
				CardmarketEUR:     eur,
				TCGPlayerUSDFoil:  usdFoil,
				CardmarketEURFoil: eurFoil,
				OracleText:        c.OracleText,
				TypeLine:          c.TypeLine,
				Legalities:        castLegalities(c.Legalities),
				ImageURL:          c.bestImageURL(),
			}
			if c.CardKingdomID != nil {
				meta.CardKingdomID = *c.CardKingdomID
			}
			if c.CardKingdomFoilID != nil {
				meta.CardKingdomFoilID = *c.CardKingdomFoilID
			}
			results[c.ID] = meta
		}
	}

	return results, nil
}
