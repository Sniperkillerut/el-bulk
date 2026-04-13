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
)

const CardKingdomPricelistURL = "https://api.cardkingdom.com/api/v2/pricelist"

type CardKingdomProduct struct {
	ID          int    `json:"id"`
	SKU         string `json:"sku"`
	Name        string `json:"name"`
	Variation   string `json:"variation"`
	Edition     string `json:"edition"`
	IsFoil      bool   `json:"is_foil"`
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

// BuildCardKingdomPriceMap downloads the CK pricelist and builds a lookup map.
// The map is keyed by a composite of (Name, Edition, Variation, IsFoil).
func BuildCardKingdomPriceMap(ctx context.Context) (map[string]*float64, error) {
	logger.InfoCtx(ctx, "Downloading CardKingdom pricelist...")
	
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

		// Create a unique key for the card. 
		// We use Name, Edition (Set), and normalized Variation.
		key := generateCKKey(p.Name, p.Edition, p.Variation, p.IsFoil)
		priceMap[key] = &price
		
		// Also store by CK ID for direct matching if available
		priceMap[fmt.Sprintf("ckid:%d", p.ID)] = &price
	}

	logger.InfoCtx(ctx, "Parsed %d CardKingdom products", len(ckResp.Data))
	return priceMap, nil
}

func generateCKKey(name, edition, variation string, isFoil bool) string {
	name = strings.ToLower(strings.TrimSpace(name))
	edition = strings.ToLower(strings.TrimSpace(edition))
	variation = strings.ToLower(strings.TrimSpace(variation))
	
	foilSuffix := "non_foil"
	if isFoil {
		foilSuffix = "foil"
	}

	return fmt.Sprintf("%s|%s|%s|%s", name, edition, variation, foilSuffix)
}

// MapFoilTreatmentToCKVariation attempts to find the appropriate CK variation string 
// for a given foil/card treatment.
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
	}

	switch foil {
	case models.FoilEtchedFoil:
		parts = append(parts, "etched")
	case models.FoilSurgeFoil:
		parts = append(parts, "surge foil")
	case models.FoilGalaxyFoil:
		parts = append(parts, "galaxy foil")
	// Add more mappings as discovered in CK data
	}

	return strings.Join(parts, " ")
}
