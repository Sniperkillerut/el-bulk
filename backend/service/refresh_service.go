package service

import (
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

func (s *RefreshService) RunPriceRefresh() (updated int, errs int) {
	rows, err := s.Store.ListRefreshableProducts()
	if err != nil {
		logger.Error("[price-refresh] failed to query products: %v", err)
		return 0, 1
	}

	if len(rows) == 0 {
		logger.Info("[price-refresh] no products with external price source, skipping")
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
		priceMap, err = external.BuildPriceMap()
		if err != nil {
			logger.Error("[price-refresh] scryfall bulk download failed: %v", err)
			return 0, len(mtgRows)
		}
	}

	updates, resolveErrs := store.BuildPriceUpdates(mtgRows, priceMap)
	errs += resolveErrs

	updated, updateErrs := s.Store.BulkUpdateMetadata(updates)
	errs += updateErrs

	logger.Info("[price-refresh] complete: %d updated, %d errors", updated, errs)
	return updated, errs
}
