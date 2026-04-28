package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	// HTTP Client for API requests
	externalClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	ScryfallBase = "https://api.scryfall.com"
)

// ResolveMTGPrice handles single-card resolution with hierarchical fallback.
func ResolveMTGPrice(
	sid, name, setCode, cn, foil, cardTreatment, ckEdition, ckVariation string,
	scryMap map[PriceKey]CardMetadata,
	idMap map[string]CardMetadata,
	ckMap map[string]*float64,
) ResolvedPrices {
	return ResolveMTGPriceBatch(sid, name, setCode, cn, foil, cardTreatment, ckEdition, ckVariation, scryMap, idMap, ckMap)
}

// ResolveMTGPriceBatch (Batch Resolution for RefreshService)
// This version uses pre-loaded maps to avoid N+1 database queries during bulk sync.
func ResolveMTGPriceBatch(
	sid, name, setCode, cn, foil, cardTreatment, ckEdition, ckVariation string,
	scryMap map[PriceKey]CardMetadata,
	idMap map[string]CardMetadata,
	ckMap map[string]*float64,
) ResolvedPrices {
	var result ResolvedPrices
	foil = strings.ToLower(foil)

	// 1. Resolve curated metadata from Scryfall
	var meta CardMetadata
	found := false

	// Try ID lookup first
	if sid != "" {
		meta, found = idMap[sid]
	}

	// Step 2: Name | Set | Collector (Exact printing)
	if !found && name != "" && setCode != "" && cn != "" {
		key := PriceKey{
			Name:      strings.ToLower(name),
			SetCode:   strings.ToLower(setCode),
			Collector: cn,
		}
		meta, found = scryMap[key]
	}

	// Step 3: Name | Set (Hierarchical fallback)
	if !found && name != "" && setCode != "" {
		key := PriceKey{
			Name:    strings.ToLower(name),
			SetCode: strings.ToLower(setCode),
		}
		meta, found = scryMap[key]
	}

	// Step 4: Name (Global Fallback)
	if !found && name != "" {
		key := PriceKey{
			Name: strings.ToLower(name),
		}
		meta, found = scryMap[key]
	}

	if found {
		result.ScryfallID = meta.ScryfallID
		result.Metadata = &meta
		// Pick relevant price based on foil
		if foil != "non_foil" && foil != "" {
			result.TCGPlayerUSD = meta.TCGPlayerUSDFoil
			result.CardmarketEUR = meta.CardmarketEURFoil

			// Intelligent Fallback: if we want foil but scryfall only has one price field populated
			if result.TCGPlayerUSD == nil {
				result.TCGPlayerUSD = meta.TCGPlayerUSD
			}
			if result.CardmarketEUR == nil {
				result.CardmarketEUR = meta.CardmarketEUR
			}
		} else {
			result.TCGPlayerUSD = meta.TCGPlayerUSD
			result.CardmarketEUR = meta.CardmarketEUR
		}
	}

	// 2. Resolve Card Kingdom Price (Independent source)
	if ckMap != nil {
		isFoil := (foil != "non_foil" && foil != "")
		result.CardKingdomUSD = LookupCKPrice(sid, name, ckEdition, ckVariation, isFoil, ckMap)
	}

	return result
}

// FetchLiveScryfallCard (Optional: Live fallback if cache misses)
func FetchLiveScryfallCard(ctx context.Context, sid string) (*CardMetadata, error) {
	if sid == "" {
		return nil, fmt.Errorf("empty scryfall id")
	}

	resp, err := externalClient.Get(fmt.Sprintf("%s/cards/%s", ScryfallBase, sid))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("scryfall api returned %d", resp.StatusCode)
	}

	var c CardMetadata
	if err := json.NewDecoder(resp.Body).Decode(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

// UpdatePricesFromDB updates the provided models.Product with latest prices.
func UpdatePricesFromDB(ctx context.Context, db *sqlx.DB, product interface{}) error {
	// Legacy bridge stub
	return nil
}
