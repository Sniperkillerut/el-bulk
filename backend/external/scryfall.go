package external

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"bytes"
	"sync"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

// scryfallLookupClient is used for individual card lookups and small metadata requests.
var scryfallLookupClient = &http.Client{Timeout: 30 * time.Second}

// scryfallBulkClient is used for massive bulk data downloads that require long-lived streams.
var scryfallBulkClient = &http.Client{Timeout: 0}

// scryfallCard is the minimal subset of the Scryfall card object we need.
type scryfallCard struct {
	ID              string `json:"id"`
	OracleID        string `json:"oracle_id"`
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
		Name       string `json:"name"`
		TypeLine   string `json:"type_line"`
		OracleText string `json:"oracle_text"`
		Artist     string `json:"artist"`
		ImageURIs  struct {
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
	Lang              string            `json:"lang"`
	ColorIdentity     []string          `json:"color_identity"`
	Rarity            string            `json:"rarity"`
	CMC               float64           `json:"cmc"`
	TypeLine          string            `json:"type_line"`
	OracleText        string            `json:"oracle_text"`
	Artist            string            `json:"artist"`
	Variation         bool              `json:"variation"`
	BorderColor       string            `json:"border_color"`
	Frame             string            `json:"frame"`
	FullArt           bool              `json:"full_art"`
	Textless          bool              `json:"textless"`
	PromoTypes        []string          `json:"promo_types"`
	Finishes          []string          `json:"finishes"`
	FrameEffects      []string          `json:"frame_effects"`
	Legalities        map[string]string `json:"legalities"`
	CardKingdomID     *string           `json:"card_kingdom_id"`
	CardKingdomFoilID *string           `json:"card_kingdom_foil_id"`
}

type CardIdentifier struct {
	ScryfallID      string `json:"id,omitempty"`
	Name            string `json:"name,omitempty"`
	SetCode         string `json:"set_code,omitempty"`
	CollectorNumber string `json:"collector_number,omitempty"`
	Foil            string `json:"foil,omitempty"`
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
// scryfallPrices extracts TCGPlayer and Cardmarket prices appropriate for the
// given foil treatment. foilTreatment should be one of the foil_treatment_type values.
func (c *scryfallCard) scryfallPrices(foilTreatment string) (tcgUSD, cmEUR *float64) {
	foilTreatment = strings.ToLower(foilTreatment)

	// 1. Try to get the exact price for the requested treatment
	switch foilTreatment {
	case "etched_foil":
		tcgUSD = parsePrice(c.Prices.USDEtched)
		cmEUR = parsePrice(c.Prices.EURFoil) // Cardmarket typically uses EURFoil for all foil types
	case "non_foil", "":
		tcgUSD = parsePrice(c.Prices.USD)
		cmEUR = parsePrice(c.Prices.EUR)
	default:
		// Generic foil or specialized foil (surge, galaxy, etc.)
		tcgUSD = parsePrice(c.Prices.USDFoil)
		cmEUR = parsePrice(c.Prices.EURFoil)
	}

	// 2. Intelligent Fallback: If we requested a foil price but that specific
	// field is empty, try the standard foil field (or vice versa).
	// This handles cards that only have one type of foil finish.
	if foilTreatment != "non_foil" && foilTreatment != "" {
		if tcgUSD == nil {
			tcgUSD = parsePrice(c.Prices.USDFoil)
			if tcgUSD == nil {
				tcgUSD = parsePrice(c.Prices.USDEtched)
			}
		}
		if cmEUR == nil {
			cmEUR = parsePrice(c.Prices.EURFoil)
		}
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
			// VALIDATION: Ensure the card found by ID actually matches the required finish.
			// Normalizes empty/non_foil to ensure consistency.
			requestedFinish := strings.ToLower(foilTreatment)
			if requestedFinish == "" {
				requestedFinish = "non_foil"
			}
			foundFinish := string(res.FoilTreatment)

			if foundFinish == requestedFinish {
				return res, nil
			}
			logger.WarnCtx(ctx, "Self-Healing: Scryfall ID %s finish mismatch (%s vs %s). Forcing resolution by name.",
				scryfallID, foundFinish, requestedFinish)
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
// It transparently handles Scryfall's limit of 75 identifiers per request.
func BatchLookupMTGCard(ctx context.Context, identifiers []CardIdentifier) ([]CardLookupResult, error) {
	if len(identifiers) == 0 {
		return nil, nil
	}

	const scryfallBatchLimit = 75
	results := make([]CardLookupResult, 0, len(identifiers))

	for i := 0; i < len(identifiers); i += scryfallBatchLimit {
		end := i + scryfallBatchLimit
		if end > len(identifiers) {
			end = len(identifiers)
		}
		chunk := identifiers[i:end]

		reqBody, _ := json.Marshal(scryfallCollectionRequest{Identifiers: chunk})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, ScryfallBase+"/cards/collection", strings.NewReader(string(reqBody)))
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "ElBulkTCGStore/1.0 (contact@elbulk.com)")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := scryfallLookupClient.Do(req)
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

		for _, card := range batchResp.Data {
			// Match back to original identifier to get the requested foil treatment
			requestedFoil := ""
			for _, id := range chunk {
				if (id.ScryfallID != "" && id.ScryfallID == card.ID) ||
					(id.SetCode != "" && strings.EqualFold(id.SetCode, card.Set) && id.CollectorNumber != "" && id.CollectorNumber == card.CollectorNumber) {
					requestedFoil = id.Foil
					break
				}
			}
			results = append(results, *mapScryfallToResult(&card, requestedFoil))
		}

		// Brief sleep to respect Scryfall rate limits during multiple chunks
		if end < len(identifiers) {
			time.Sleep(100 * time.Millisecond)
		}
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

	resp, err := scryfallLookupClient.Do(req)
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
		ScryfallID:      card.ID,
		OracleID:        card.OracleID,
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
			FrameEffects:    card.FrameEffects,
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

	// 1. Identify the print's specific foil treatment if it has any foil finishes
	specificFoil := models.FoilNonFoil
	if hasFoil || hasEtched {
		// Start with standard fallbacks
		if hasEtched {
			specificFoil = models.FoilEtchedFoil
		} else if hasFoil {
			specificFoil = models.FoilFoil
		} else if hasGlossy {
			specificFoil = models.FoilGalaxyFoil
		}

		// Check for specialized tags which take precedence over standard "foil" finish
		for _, pt := range card.PromoTypes {
			switch pt {
			case "ripplefoil":
				specificFoil = models.FoilRippleFoil
			case "surgefoil":
				specificFoil = models.FoilSurgeFoil
			case "confettifoil":
				specificFoil = models.FoilConfettiFoil
			case "textured":
				specificFoil = models.FoilTexturedFoil
			case "stepandcompleat":
				specificFoil = models.FoilStepAndCompleat
			case "oilslick":
				specificFoil = models.FoilOilSlick
			case "neonink":
				specificFoil = models.FoilNeonInk
			case "galaxyfoil":
				specificFoil = models.FoilGalaxyFoil
			case "doublerainbow":
				specificFoil = models.FoilDoubleRainbow
			case "platinumfoil":
				specificFoil = models.FoilPlatinumFoil
			}
		}
	}

	// 2. If caller requested a specific treatment, respect it if the print supports it
	if requestedFoil != "" {
		rf := models.FoilTreatment(requestedFoil)
		// If they asked for the specific specialized foil we found, excellent
		if rf == specificFoil && specificFoil != models.FoilNonFoil {
			return rf
		}
		// If they asked for generic "foil" but we found a more specific one, give them the specific one
		if rf == models.FoilFoil && (specificFoil != models.FoilNonFoil && specificFoil != models.FoilEtchedFoil) {
			return specificFoil
		}
		// Basic finish matches
		if rf == models.FoilFoil && hasFoil {
			return models.FoilFoil
		}
		if rf == models.FoilEtchedFoil && hasEtched {
			return models.FoilEtchedFoil
		}
		if rf == models.FoilNonFoil && hasNonFoil {
			return models.FoilNonFoil
		}
	}

	// 3. Default to Non-Foil if supported (Store preference for general lookups)
	if hasNonFoil {
		return models.FoilNonFoil
	}

	// 4. Final fallback to whatever specific foil we identified
	return specificFoil
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
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Set             string            `json:"set"` // set code, e.g. "m11"
	CollectorNumber string            `json:"collector_number"`
	TypeLine        string            `json:"type_line"`
	OracleText      string            `json:"oracle_text"`
	Legalities      map[string]string `json:"legalities"`
	ImageURIs       struct {
		Normal string `json:"normal"`
	} `json:"image_uris"`
	Prices struct {
		USD       *string `json:"usd"`
		USDFoil   *string `json:"usd_foil"`
		USDEtched *string `json:"usd_etched"`
		EUR       *string `json:"eur"`
		EURFoil   *string `json:"eur_foil"`
	} `json:"prices"`
	FrameEffects      []string `json:"frame_effects"`
	CardKingdomID     *string `json:"card_kingdom_id"`
	CardKingdomFoilID *string `json:"card_kingdom_foil_id"`
}

var (
	scryCache      map[PriceKey]CardMetadata
	idCache        map[string]CardMetadata // NEW: Fast ID lookup
	scryCacheMutex sync.RWMutex
	scryCacheTime  time.Time
)

// ResetScryfallCache clears the in-memory Scryfall cache (primarily for unit tests).
func ResetScryfallCache() {
	scryCacheMutex.Lock()
	defer scryCacheMutex.Unlock()
	scryCache = nil
	idCache = nil
	scryCacheTime = time.Time{}
}

// BuildPriceMap loads Scryfall metadata from the external_scryfall table.
// Uses an in-memory cache valid for 1 hour.
func BuildPriceMap(ctx context.Context, db *sqlx.DB) (map[PriceKey]CardMetadata, map[string]CardMetadata, error) {
	scryCacheMutex.RLock()
	if scryCache != nil && idCache != nil && time.Since(scryCacheTime) < 1*time.Hour {
		cache := scryCache
		iCache := idCache
		scryCacheMutex.RUnlock()
		return cache, iCache, nil
	}
	scryCacheMutex.RUnlock()

	scryCacheMutex.Lock()
	defer scryCacheMutex.Unlock()

	// Double check
	if scryCache != nil && idCache != nil && time.Since(scryCacheTime) < 1*time.Hour {
		return scryCache, idCache, nil
	}

	logger.InfoCtx(ctx, "Loading Scryfall metadata from database...")

	type row struct {
		ScryfallID      string   `db:"scryfall_id"`
		Name            string   `db:"name"`
		SetCode         string   `db:"set_code"`
		CollectorNumber string   `db:"collector_number"`
		PriceUSD        *float64 `db:"price_usd"`
		PriceUSDFoil    *float64 `db:"price_usd_foil"`
		PriceEUR        *float64 `db:"price_eur"`
		ImageURL        string   `db:"image_url"`
	}

	var rows []row
	err := db.SelectContext(ctx, &rows, "SELECT scryfall_id, name, set_code, collector_number, price_usd, price_usd_foil, price_eur, image_url FROM external_scryfall")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query external_scryfall: %w", err)
	}

	priceMap := make(map[PriceKey]CardMetadata, len(rows))
	localIdCache := make(map[string]CardMetadata, len(rows))

	for _, r := range rows {
		meta := CardMetadata{
			ScryfallID:       r.ScryfallID,
			TCGPlayerUSD:     r.PriceUSD,
			TCGPlayerUSDFoil: r.PriceUSDFoil,
			CardmarketEUR:    r.PriceEUR,
			ImageURL:         r.ImageURL,
		}

		// 1. Exact Match Key (Name + Set + Collector)
		key := PriceKey{
			Name:      strings.ToLower(r.Name),
			SetCode:   strings.ToLower(r.SetCode),
			Collector: r.CollectorNumber,
		}
		priceMap[key] = meta

		// 2. Fallback: Name + Set (Highest priority printing in set)
		setKey := PriceKey{
			Name:    strings.ToLower(r.Name),
			SetCode: strings.ToLower(r.SetCode),
		}
		if _, ok := priceMap[setKey]; !ok {
			priceMap[setKey] = meta
		}

		// 3. Fallback: Name Global (Highest priority printing found)
		nameKey := PriceKey{
			Name: strings.ToLower(r.Name),
		}
		if _, ok := priceMap[nameKey]; !ok {
			priceMap[nameKey] = meta
		}

		localIdCache[r.ScryfallID] = meta
	}

	scryCache = priceMap
	idCache = localIdCache
	scryCacheTime = time.Now()

	logger.InfoCtx(ctx, "Loaded %d Scryfall cards from database", len(rows))
	return scryCache, idCache, nil
}

func FetchBulkDataURL(ctx context.Context, client *http.Client) (string, error) {
	if client == nil {
		client = scryfallLookupClient
	}
	metaReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, ScryfallBase+"/bulk-data", nil)
	metaReq.Header.Set("User-Agent", "ElBulkTCGStore/1.0 (contact@elbulk.com)")
	metaReq.Header.Set("Accept", "application/json")

	metaResp, err := client.Do(metaReq)
	if err != nil {
		return "", fmt.Errorf("fetching bulk-data index: %w", err)
	}
	defer metaResp.Body.Close()

	var meta ScryfallBulkMeta
	if err := json.NewDecoder(metaResp.Body).Decode(&meta); err != nil {
		return "", fmt.Errorf("decoding bulk-data index: %w", err)
	}

	for _, item := range meta.Data {
		if item.Type == "default_cards" {
			return item.DownloadURI, nil
		}
	}
	return "", fmt.Errorf("default_cards bulk file not found in scryfall response")
}

func DownloadBulkData(ctx context.Context, client *http.Client) (io.ReadCloser, error) {
	// Step 1: discover today's bulk-data download URL
	downloadURL, err := FetchBulkDataURL(ctx, client)
	if err != nil {
		return nil, err
	}

	if client == nil {
		client = scryfallBulkClient
	}
	dlReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	dlReq.Header.Set("User-Agent", "ElBulkTCGStore/1.0 (contact@elbulk.com)")

	dlResp, err := client.Do(dlReq)
	if err != nil {
		return nil, fmt.Errorf("downloading bulk data: %w", err)
	}
	return dlResp.Body, nil
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

	resp, err := scryfallLookupClient.Do(req)
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

		resp, err := scryfallLookupClient.Do(req)
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

// SyncScryfallToDB streams the Scryfall default_cards JSON into the external_scryfall table.
// If r is nil, it will automatically download the bulk data.
func SyncScryfallToDB(ctx context.Context, db *sqlx.DB, r io.Reader) error {
	logger.InfoCtx(ctx, "Syncing Scryfall bulk data to database...")

	if r == nil {
		body, err := DownloadBulkData(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to download scryfall bulk data: %w", err)
		}
		defer body.Close()
		r = body
	}

	// 1. Clear existing data (Full refresh pattern)
	if _, err := db.ExecContext(ctx, "TRUNCATE TABLE external_scryfall"); err != nil {
		return fmt.Errorf("failed to truncate external_scryfall: %w", err)
	}
	decoder := json.NewDecoder(r)

	// Read opening '['
	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("failed to read scryfall JSON opening: %w", err)
	}

	const batchSize = 1000
	var batch []ScryfallBulkCard

	for decoder.More() {
		var card ScryfallBulkCard
		if err := decoder.Decode(&card); err != nil {
			logger.WarnCtx(ctx, "Skipping malformed Scryfall card: %v", err)
			continue
		}
		batch = append(batch, card)

		if len(batch) >= batchSize {
			if err := flushScryBatch(ctx, db, batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		if err := flushScryBatch(ctx, db, batch); err != nil {
			return err
		}
	}

	logger.InfoCtx(ctx, "Successfully synced Scryfall bulk data to database")
	return nil
}

func flushScryBatch(ctx context.Context, db *sqlx.DB, batch []ScryfallBulkCard) error {
	query := `
		INSERT INTO external_scryfall (scryfall_id, name, set_code, collector_number, price_usd, price_usd_foil, price_eur, image_url, frame_effects)
		VALUES (:scryfall_id, :name, :set_code, :collector_number, :price_usd, :price_usd_foil, :price_eur, :image_url, :frame_effects)
	`
	type row struct {
		ScryfallID      string   `db:"scryfall_id"`
		Name            string   `db:"name"`
		SetCode         string   `db:"set_code"`
		CollectorNumber string   `db:"collector_number"`
		PriceUSD        *float64 `db:"price_usd"`
		PriceUSDFoil    *float64 `db:"price_usd_foil"`
		PriceEUR        *float64 `db:"price_eur"`
		ImageURL        string   `db:"image_url"`
		FrameEffects    string   `db:"frame_effects"`
	}

	rows := make([]row, len(batch))
	for i, c := range batch {
		rows[i] = row{
			ScryfallID:      c.ID,
			Name:            c.Name,
			SetCode:         c.Set,
			CollectorNumber: c.CollectorNumber,
			PriceUSD:        parsePrice(c.Prices.USD),
			PriceUSDFoil:    parsePrice(c.Prices.USDFoil),
			PriceEUR:        parsePrice(c.Prices.EUR),
			ImageURL:        c.ImageURIs.Normal,
			FrameEffects:    "[]", // Default to empty array
		}
		if len(c.FrameEffects) > 0 {
			feJSON, _ := json.Marshal(c.FrameEffects)
			rows[i].FrameEffects = string(feJSON)
		}
	}

	_, err := db.NamedExecContext(ctx, query, rows)
	return err
}
