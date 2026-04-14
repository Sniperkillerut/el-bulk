package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
	"sync"
)

const CardKingdomPricelistURL = "https://api.cardkingdom.com/api/v2/pricelist"

type CardKingdomProduct struct {
	ID          int    `json:"id"`
	SKU         string `json:"sku"`
	Name        string `json:"name"`
	Variation   string `json:"variation"`
	Edition     string `json:"edition"`
	IsFoil      string `json:"is_foil"`
	PriceRetail string `json:"price_retail"`
	QtyRetail   int    `json:"qty_retail"`
}

type CardKingdomResponse struct {
	Meta struct {
		CreatedAt   string `json:"created_at"`
		LastUpdated string `json:"last_updated"`
	} `json:"meta"`
	Data []CardKingdomProduct `json:"data"`
}

var ckClient = &http.Client{Timeout: 60 * time.Second}

var (
	ckCache      map[string]*float64
	ckCacheMutex sync.RWMutex
	ckCacheTime  time.Time
)

const CacheDuration = 1 * time.Hour

// BuildCardKingdomPriceMap downloads the CK pricelist and builds a lookup map.
// The map is keyed by a composite of (Name, Edition, Variation, IsFoil).
// Uses an in-memory cache valid for 1 hour to avoid excessive downloads.
func BuildCardKingdomPriceMap(ctx context.Context) (map[string]*float64, error) {
	ckCacheMutex.RLock()
	if ckCache != nil && time.Since(ckCacheTime) < CacheDuration {
		logger.DebugCtx(ctx, "Using cached CardKingdom pricelist (downloaded %s ago)", time.Since(ckCacheTime).Round(time.Minute))
		cacheCopy := ckCache // Reference copy is enough for read-only use
		ckCacheMutex.RUnlock()
		return cacheCopy, nil
	}
	ckCacheMutex.RUnlock()

	ckCacheMutex.Lock()
	defer ckCacheMutex.Unlock()

	// Double-check after acquiring lock
	if ckCache != nil && time.Since(ckCacheTime) < CacheDuration {
		return ckCache, nil
	}

	logger.InfoCtx(ctx, "Downloading CardKingdom pricelist (cache empty or expired)...")
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, CardKingdomPricelistURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ElBulkTCGStore/1.0 (contact@elbulk.com)")

	resp, err := ckClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CardKingdom pricelist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CardKingdom API returned status %d", resp.StatusCode)
	}

	var ckResp CardKingdomResponse
	if err := json.NewDecoder(resp.Body).Decode(&ckResp); err != nil {
		return nil, fmt.Errorf("failed to decode CardKingdom response: %w", err)
	}

	priceMap := make(map[string]*float64, len(ckResp.Data))
	for _, p := range ckResp.Data {
		price, err := strconv.ParseFloat(p.PriceRetail, 64)
		if err != nil {
			continue
		}

		isFoilBool := p.IsFoil == "true"
		key := generateCKKey(p.Name, p.Edition, p.Variation, isFoilBool)
		priceMap[key] = &price
		
		priceMap[fmt.Sprintf("ckid:%d", p.ID)] = &price
	}

	ckCache = priceMap
	ckCacheTime = time.Now()

	logger.InfoCtx(ctx, "Parsed and cached %d CardKingdom products", len(ckResp.Data))
	return priceMap, nil
}

func generateCKKey(name, edition, variation string, isFoil bool) string {
	name = strings.ToLower(strings.TrimSpace(name))
	edition = NormalizeCKEdition(strings.TrimSpace(edition))
	variation = strings.ToLower(strings.TrimSpace(variation))
	
	foilSuffix := "non_foil"
	if isFoil {
		foilSuffix = "foil"
	}

	return fmt.Sprintf("%s|%s|%s|%s", name, edition, variation, foilSuffix)
}

// NormalizeCKEdition maps common Scryfall/TCGPlayer set names to CardKingdom's edition names.
func NormalizeCKEdition(edition string) string {
	e := strings.ToLower(edition)
	
	// Common mappings
	switch e {
	case "revised":
		return "revised edition"
	case "unlimited":
		return "unlimited edition"
	case "antiquities":
		return "antiquities edition"
	case "arabian nights":
		return "arabian nights edition"
	case "legends":
		return "legends edition"
	case "the dark":
		return "the dark edition"
	case "fourth edition", "4th edition":
		return "4th edition"
	case "fifth edition", "5th edition":
		return "5th edition"
	case "sixth edition", "6th edition":
		return "6th edition"
	case "seventh edition", "7th edition":
		return "7th edition"
	case "eighth edition", "8th edition":
		return "8th edition"
	case "ninth edition", "9th edition":
		return "9th edition"
	case "tenth edition", "10th edition":
		return "10th edition"
	case "classic sixth edition":
		return "6th edition"
	case "media promos":
		return "promotional"
	case "pro tour promos":
		return "promotional"
	case "judge gift cards", "judge gift program":
		return "promotional"
	case "magic player rewards":
		return "promotional"
	case "modern horizons 1":
		return "modern horizons"
	case "modern horizons 2":
		return "modern horizons 2"
	case "time spiral remastered":
		return "time spiral remastered"
	}

	// General rules: CK often adds " Edition" to early sets if not present
	earlySets := []string{"unlimited", "antiquities", "arabian nights", "legends", "the dark", "ice age", "homelands"}
	for _, s := range earlySets {
		if e == s {
			return e + " edition"
		}
	}

	return e
}

func MapFoilTreatmentToCKVariation(foil models.FoilTreatment, treatment models.CardTreatment) string {
	// Most CK variations are things like "Etched", "Extended Art", etc.
	// They are often combined in the variation field.
	
	var parts []string
	
	switch treatment {
	case models.TreatmentBorderless:
		parts = append(parts, "borderless")
	case models.TreatmentExtendedArt:
		parts = append(parts, "extended art")
	case models.TreatmentShowcase:
		parts = append(parts, "showcase")
	case models.TreatmentEtched:
		parts = append(parts, "etched")
	case models.TreatmentFullArt:
		parts = append(parts, "full art")
	case models.TreatmentAlternateArt:
		parts = append(parts, "alternate art")
	case models.TreatmentLegacyBorder, "retro": // Retro is common in variations
		parts = append(parts, "retro")
	}

	switch foil {
	case models.FoilEtchedFoil:
		parts = append(parts, "etched")
	case models.FoilSurgeFoil:
		parts = append(parts, "surge foil")
	case models.FoilGalaxyFoil:
		parts = append(parts, "galaxy foil")
	case models.FoilTexturedFoil:
		parts = append(parts, "textured")
	}

	return strings.Join(parts, " ")
}
