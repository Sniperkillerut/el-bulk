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

		var idMap map[string]external.CardMetadata
		if needsScry {
			scryPriceMap, idMap, err = external.BuildPriceMap(ctx)
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

		updates, resolveErrs := store.BuildPriceUpdates(mtgRows, scryPriceMap, idMap, ckPriceMap)
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

	return 0, 0
}

func (s *RefreshService) GetSuggestedPrice(ctx context.Context, scryfallID, name, set, setName, collector, foil, treatment, source string) (*float64, error) {
	foil = strings.ToLower(foil)
	
	if source == "cardkingdom" {
		ckMap, err := external.BuildCardKingdomPriceMap(ctx)
		if err != nil {
			return nil, err
		}

		variation := external.MapFoilTreatmentToCKVariation(models.FoilTreatment(foil), models.CardTreatment(treatment))
		isFoil := foil != "non_foil"

		// Prefer curated ck_name (setName from DB); fall back to normalized Scryfall name.
		ckEdition := external.NormalizeCKEdition(setName)

		// LookupCKPrice uses scryfallID for direct O(1) match when available.
		price := external.LookupCKPrice(scryfallID, name, ckEdition, variation, isFoil, ckMap)
		if price != nil {
			return price, nil
		}

		return nil, fmt.Errorf("no cardkingdom price found for %s", name)
	}

	if source == "tcgplayer" || source == "cardmarket" {
		// FAST-PATH: Use targeted API lookup instead of 600MB bulk file for single product refreshes
		res, err := external.LookupMTGCard(ctx, scryfallID, name, set, collector, foil)
		if err != nil {
			return nil, err
		}

		// Convert result to the format ResolveMTGPrice expects
		scryBatch := map[string]external.CardMetadata{
			*res.MTGMetadata.ScryfallID: res.ToCardMetadata(),
		}

		// Unified Resolve: implements ID > Set|CN > Set|Foil hierarchy
		pResult := external.ResolveMTGPrice(
			scryfallID, name, set, collector, foil, treatment, "", "",
			nil, scryBatch, nil,
		)

		if pResult.Metadata != nil {
			if source == "cardmarket" {
				return pResult.Metadata.CardmarketEUR, nil
			}
			return pResult.Metadata.TCGPlayerUSD, nil
		}
	}

	return nil, fmt.Errorf("no price found for %s", source)
}
