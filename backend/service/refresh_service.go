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
			logger.InfoCtx(ctx, "[price-refresh] triggering fresh Scryfall sync to DB...")
			if err := external.SyncScryfallToDB(ctx, s.Store.DB, nil); err != nil {
				logger.WarnCtx(ctx, "[price-refresh] Scryfall sync failed, falling back to existing DB data: %v", err)
			}
			scryPriceMap, idMap, err = external.BuildPriceMap(ctx, s.Store.DB)
			if err != nil {
				logger.ErrorCtx(ctx, "[price-refresh] failed to build scryfall map from DB: %v", err)
				return 0, len(mtgRows)
			}
		}

		if needsCK {
			logger.InfoCtx(ctx, "[price-refresh] triggering fresh Card Kingdom sync to DB...")
			if err := external.SyncCardKingdomToDB(ctx, s.Store.DB, nil); err != nil {
				logger.WarnCtx(ctx, "[price-refresh] CK sync failed, falling back to existing DB data: %v", err)
			}
			ckPriceMap, err = external.BuildCardKingdomPriceMap(ctx, s.Store.DB)
			if err != nil {
				logger.ErrorCtx(ctx, "[price-refresh] failed to build CK map from DB: %v", err)
				errs += len(mtgRows)
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

func (s *RefreshService) SyncCurrencyRates(ctx context.Context) error {
	rates, err := external.FetchExchangeRates(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch rates: %w", err)
	}

	usdToCop, okUSD := rates["COP"]
	eurToUSD, okEUR := rates["EUR"]
	if !okUSD || !okEUR {
		return fmt.Errorf("COP or EUR rates missing from response")
	}

	// EUR -> COP = (1 / EUR_to_USD) * USD_to_COP
	eurToCop := (1.0 / eurToUSD) * usdToCop

	// Update settings
	if err := s.Settings.Upsert(ctx, "usd_to_cop_rate", fmt.Sprintf("%.2f", usdToCop)); err != nil {
		return err
	}
	if err := s.Settings.Upsert(ctx, "eur_to_cop_rate", fmt.Sprintf("%.2f", eurToCop)); err != nil {
		return err
	}

	// Fetch current CK rate and other settings for recomputation
	settings, err := s.Settings.GetSettings(ctx)
	if err != nil {
		return err
	}

	// Recompute all prices in the database to keep materialized column in sync
	return s.Store.RecalculateAllPrices(ctx, usdToCop, eurToCop, settings.CKToCOPRate)
}

func (s *RefreshService) GetSuggestedPrice(ctx context.Context, scryfallID, name, set, setName, collector, foil, treatment, source, ckEdition string) (*float64, error) {
	foil = strings.ToLower(foil)

	// 1. Fetch real-time Scryfall metadata (Fast-Path)
	res, err := external.LookupMTGCard(ctx, scryfallID, name, set, collector, foil)
	if err != nil {
		return nil, err
	}

	// 2. Prepare resolution maps
	scryBatch := map[string]external.CardMetadata{
		*res.MTGMetadata.ScryfallID: res.ToCardMetadata(),
	}

	var ckMap map[string]*float64
	if source == "cardkingdom" {
		ckMap, _ = external.BuildCardKingdomPriceMap(ctx, s.Store.DB)
	}

	// 3. Resolve curated metadata and use 'ground truth' textures from Scryfall
	// Fall back to requested foil/treatment if Scryfall doesn't specify one
	resFoil := string(res.FoilTreatment)
	resTreatment := string(res.CardTreatment)

	finalFoil := foil
	if finalFoil == "" {
		finalFoil = resFoil
	}
	finalTreatment := treatment
	if finalTreatment == "" {
		finalTreatment = resTreatment
	}

	if ckEdition == "" {
		ckEdition = external.NormalizeCKEdition(setName)
	}
	variation := external.MapFoilTreatmentToCKVariation(models.FoilTreatment(finalFoil), models.CardTreatment(finalTreatment))

	// 4. Unified Resolve (Single Source of Truth)
	pResult := external.ResolveMTGPrice(
		scryfallID, name, set, collector, finalFoil, finalTreatment,
		ckEdition, variation,
		nil, scryBatch, ckMap,
	)

	// 5. Select requested price (Use unified resolved fields)
	switch source {
	case "cardkingdom":
		if pResult.CardKingdomUSD != nil {
			return pResult.CardKingdomUSD, nil
		}
	case "cardmarket":
		if pResult.CardmarketEUR != nil {
			return pResult.CardmarketEUR, nil
		}
	case "tcgplayer":
		if pResult.TCGPlayerUSD != nil {
			return pResult.TCGPlayerUSD, nil
		}
	}

	return nil, fmt.Errorf("no %s price found for %s", source, name)
}
