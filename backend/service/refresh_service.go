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

func (s *RefreshService) RunPriceRefresh(ctx context.Context) (updated int, errs int) {
	rows, err := s.Store.ListRefreshableProducts(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "[price-refresh] failed to query products: %v", err)
		return 0, 1
	}

	if len(rows) == 0 {
		logger.InfoCtx(ctx, "[price-refresh] no products with external price source, skipping")
		return 0, 0
	}

	// Separate MTG products (use Scryfall) from non-MTG (not yet supported)
	var mtgRows []store.RefreshRow
	for _, r := range rows {
		if r.TCG == "mtg" {
			mtgRows = append(mtgRows, r)
		}
	}

	var priceMap map[external.PriceKey]external.CardMetadata
	if len(mtgRows) > 0 {
		priceMap, err = external.BuildPriceMap(ctx)
		if err != nil {
			logger.ErrorCtx(ctx, "[price-refresh] scryfall bulk download failed: %v", err)
			return 0, len(mtgRows)
		}
	}

	updates, resolveErrs := store.BuildPriceUpdates(mtgRows, priceMap)
	errs += resolveErrs

	updated, updateErrs := s.Store.BulkUpdateMetadata(ctx, updates)
	errs += updateErrs

	logger.InfoCtx(ctx, "[price-refresh] complete: %d updated, %d errors", updated, errs)
	return updated, errs
}
