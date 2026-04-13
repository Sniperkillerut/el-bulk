package service

import (
	"context"
	"github.com/el-bulk/backend/external"
	"github.com/el-bulk/backend/store"
	"github.com/el-bulk/backend/utils/logger"
)

type RefreshService struct {
	Store *store.RefreshStore
}

func NewRefreshService(s *store.RefreshStore) *RefreshService {
	return &RefreshService{Store: s}
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
	needsCK := false
	for _, r := range filteredRows {
		if r.TCG == "mtg" {
			mtgRows = append(mtgRows, r)
			if r.PriceSource == "cardkingdom" {
				needsCK = true
			}
		}
	}

	var scryPriceMap map[external.PriceKey]external.CardMetadata
	var ckPriceMap map[string]*float64
	if len(mtgRows) > 0 {
		var err error
		scryPriceMap, err = external.BuildPriceMap(ctx)
		if err != nil {
			logger.ErrorCtx(ctx, "[price-refresh] scryfall bulk download failed: %v", err)
			return 0, len(mtgRows)
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

	updated, updateErrs := s.Store.BulkUpdateMetadata(ctx, updates)
	errs += updateErrs

	logger.InfoCtx(ctx, "[price-refresh] complete: %d updated, %d errors", updated, errs)
	return updated, errs
}
