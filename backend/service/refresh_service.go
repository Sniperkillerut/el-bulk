package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/el-bulk/backend/external"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/el-bulk/backend/utils/logger"
)

type RefreshService struct {
	Store    *store.RefreshStore
	Settings *SettingsService
}

func NewRefreshService(s *store.RefreshStore, settings *SettingsService) *RefreshService {
	return &RefreshService{Store: s, Settings: settings}
}

func (s *RefreshService) RunPriceRefresh(ctx context.Context, tcgID string) (updated int, errs int) {
	rows, err := s.Store.ListRefreshableProducts(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "[price-refresh] failed to query products: %v", err)
		return 0, 1
	}

	if len(rows) == 0 {
		logger.InfoCtx(ctx, "[price-refresh] no products with external price source, skipping")
		return 0, 0
	}

	// Filter by TCG if tcgID is provided
	var filteredRows []store.RefreshRow
	if tcgID != "" {
		for _, r := range rows {
			if r.TCG == tcgID {
				filteredRows = append(filteredRows, r)
			}
		}
	} else {
		filteredRows = rows
	}

	if len(filteredRows) == 0 {
		logger.InfoCtx(ctx, "[price-refresh] no products found for TCG %q, skipping", tcgID)
		return 0, 0
	}

	// Separate MTG products (use Scryfall) from non-MTG (not yet supported)
	var mtgRows []store.RefreshRow
	needsScry := false
	needsCK := false
	for _, r := range filteredRows {
		if r.TCG == "mtg" {
			mtgRows = append(mtgRows, r)
			switch r.PriceSource {
			case "cardkingdom":
				needsCK = true
			case "tcgplayer", "cardmarket":
				needsScry = true
			}
		}
	}

	var scryPriceMap map[external.PriceKey]external.CardMetadata
	var ckPriceMap map[string]*float64
	if len(mtgRows) > 0 {
		var err error

		if needsScry {
			scryPriceMap, err = external.BuildPriceMap(ctx)
			if err != nil {
				logger.ErrorCtx(ctx, "[price-refresh] scryfall bulk download failed: %v", err)
				return 0, len(mtgRows)
			}
		}

		if needsCK {
			ckPriceMap, err = external.BuildCardKingdomPriceMap(ctx)
			if err != nil {
				logger.ErrorCtx(ctx, "[price-refresh] cardkingdom pricelist download failed: %v", err)
				errs += len(mtgRows) // Consider them all errors if we can't get the source
			}
		}
	}

	updates, resolveErrs := store.BuildPriceUpdates(mtgRows, scryPriceMap, ckPriceMap)
	errs += resolveErrs

	settings, err := s.Settings.GetSettings(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "[price-refresh] failed to fetch settings for rates: %v", err)
		return 0, len(mtgRows)
	}

	updated, updateErrs := s.Store.BulkUpdateMetadata(ctx, updates, settings.USDToCOPRate, settings.EURToCOPRate, settings.CKToCOPRate)
	errs += updateErrs

	logger.InfoCtx(ctx, "[price-refresh] complete: %d updated, %d errors", updated, errs)
	return updated, errs
}

func (s *RefreshService) GetSuggestedPrice(ctx context.Context, name, set, setName, collector, foil, treatment, source string) (*float64, error) {
	foil = strings.ToLower(foil)
	
	if source == "cardkingdom" {
		ckMap, err := external.BuildCardKingdomPriceMap(ctx)
		if err != nil {
			return nil, err
		}

		// 1. Fallback to Name + Edition + Variation matching
		variation := external.MapFoilTreatmentToCKVariation(models.FoilTreatment(foil), models.CardTreatment(treatment))
		isFoil := foil != "non_foil"

		ckKey := fmt.Sprintf("%s|%s|%s|%s",
			strings.ToLower(name),
			strings.ToLower(setName),
			strings.ToLower(variation),
			func() string {
				if isFoil {
					return "foil"
				}
				return "non_foil"
			}())

		if price, ok := ckMap[ckKey]; ok {
			return price, nil
		}

		// 2. Fuzzy Edition Fallback
		nameKeyPrefix := strings.ToLower(name) + "|"
		foilSuffix := "|non_foil"
		if isFoil {
			foilSuffix = "|foil"
		}
		targetEdition := strings.ToLower(setName)
		targetCollector := strings.ToLower(collector)
		
		var bestMatch *float64

		for k, p := range ckMap {
			if strings.HasPrefix(k, nameKeyPrefix) && strings.HasSuffix(k, foilSuffix) {
				parts := strings.Split(k, "|")
				if len(parts) >= 3 {
					ckEdition := parts[1]
					ckVariation := parts[2]
					
					// Skip non-standard cards that often fluctuate valuation incorrectly
					if strings.Contains(ckVariation, "art card") || strings.Contains(ckVariation, "token") {
						continue
					}

					editionMatches := targetEdition != "" && (ckEdition == targetEdition || strings.Contains(ckEdition, targetEdition) || strings.Contains(targetEdition, ckEdition))
					collectorMatches := targetCollector != "" && (ckVariation == targetCollector || strings.Contains(ckVariation, targetCollector))
					
					if editionMatches {
						if collectorMatches {
							// Found exact variation in correct set - can't get better than this
							return p, nil
						}
						// If we haven't found a match yet, or this one is more expensive, take it
						if bestMatch == nil || (p != nil && *p > *bestMatch) {
							bestMatch = p
						}
					} else if collectorMatches {
						if bestMatch == nil || (p != nil && *p > *bestMatch) {
							bestMatch = p
						}
					}
				}
			}
		}

		if bestMatch != nil {
			return bestMatch, nil
		}

		// 3. Absolute Last Fallback: Just return the highest available CK price for this card name + foil
		for k, p := range ckMap {
			if strings.HasPrefix(k, nameKeyPrefix) && strings.HasSuffix(k, foilSuffix) {
				if bestMatch == nil || (p != nil && *p > *bestMatch) {
					bestMatch = p
				}
			}
		}

		if bestMatch != nil {
			return bestMatch, nil
		}

		return nil, fmt.Errorf("no cardkingdom price found for %s", name)
	}

	if source == "tcgplayer" || source == "cardmarket" {
		scryMap, err := external.BuildPriceMap(ctx)
		if err != nil {
			return nil, err
		}

		key := external.PriceKey{
			Name:      strings.ToLower(name),
			SetCode:   strings.ToLower(set),
			Collector: strings.TrimSpace(collector),
			Foil:      foil,
		}

		var scryMeta external.CardMetadata
		var hasScry bool
		
		if meta, ok := scryMap[key]; ok {
			scryMeta = meta
			hasScry = true
		} else {
			key.Collector = ""
			if meta, ok := scryMap[key]; ok {
				scryMeta = meta
				hasScry = true
			} else {
				key.SetCode = ""
				if meta, ok := scryMap[key]; ok {
					scryMeta = meta
					hasScry = true
				}
			}
		}
		if hasScry {
			if source == "cardmarket" {
				return scryMeta.CardmarketEUR, nil
			}
			return scryMeta.TCGPlayerUSD, nil
		}
	}

	return nil, fmt.Errorf("no price found for %s", source)
}
