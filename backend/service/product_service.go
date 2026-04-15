package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"context"

	"github.com/el-bulk/backend/external"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

type ProductService struct {
	Store    *store.ProductStore
	TCGStore *store.TCGStore
	Settings SettingsProvider
	Audit    Auditer
}

func NewProductService(s *store.ProductStore, tcg *store.TCGStore, settings SettingsProvider, audit Auditer) *ProductService {
	return &ProductService{
		Store:    s,
		TCGStore: tcg,
		Settings: settings,
		Audit:    audit,
	}
}

func (s *ProductService) List(ctx context.Context, params store.ProductFilterParams, isAdmin bool) (models.ProductListResponse, error) {
	logger.TraceCtx(ctx, "Entering ProductService.List | Admin: %v | Params: %+v", isAdmin, params)
	settings, err := s.Settings.GetSettings(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to get settings in ProductService.List: %v", err)
	} else {
		params.USDRate = settings.USDToCOPRate
		params.EURRate = settings.EURToCOPRate
		params.CKRate = settings.CKToCOPRate
	}

	products, total, err := s.Store.ListWithFilters(ctx, params)
	if err != nil {
		return models.ProductListResponse{}, err
	}

	// Settings already fetched above for sorting

	if err := s.EnrichProducts(ctx, products, settings, isAdmin); err != nil {
		return models.ProductListResponse{}, err
	}

	facets, err := s.Store.GetFacets(ctx, params, isAdmin)
	if err != nil {
		return models.ProductListResponse{}, err
	}

	return models.ProductListResponse{
		Products: products,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Facets:   facets,
	}, nil
}

func (s *ProductService) GetByID(ctx context.Context, id string, isAdmin bool) (*models.Product, error) {
	logger.TraceCtx(ctx, "Entering ProductService.GetByID | ID: %s | Admin: %v", id, isAdmin)
	product, err := s.Store.GetEnrichedByID(ctx, id)
	if err != nil {
		return nil, err
	}

	settings, err := s.Settings.GetSettings(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to get settings in ProductService.GetByID: %v", err)
	}

	products := []models.Product{*product}
	if err := s.CalculatePrices(products, settings); err != nil {
		return nil, err
	}
	if err := s.Store.PopulateCartCounts(ctx, products); err != nil {
		return nil, err
	}
	
	if isAdmin {
		if err := s.Store.PopulateStorage(ctx, products); err != nil {
			return nil, err
		}
	} else {
		// Filter categories for public view
		if err := s.Store.PopulateCategories(ctx, products); err != nil {
			return nil, err
		}
		filtered := []models.CustomCategory{}
		for _, c := range products[0].Categories {
			if c.ShowBadge {
				filtered = append(filtered, c)
			}
		}
		products[0].Categories = filtered
	}

	return &products[0], nil
}

func (s *ProductService) Create(ctx context.Context, input models.ProductInput) (*models.Product, error) {
	logger.TraceCtx(ctx, "Entering ProductService.Create | Name: %s", input.Name)

	// Conditional Default Pricing Source
	if input.PriceSource == "" {
		if input.TCG == "mtg" && (input.Category == "singles" || input.Category == "store_exclusives") {
			input.PriceSource = models.PriceSourceCardKingdom
		} else {
			input.PriceSource = models.PriceSourceManual
		}
	}

	product, err := s.Store.CreateProduct(ctx, input)
	if err != nil {
		logger.ErrorCtx(ctx, "Core product creation failed: %v", err)
		return nil, err
	}

	// Logging (Asynchronous/Non-blocking)
	s.Audit.LogAction(ctx, "CREATE_PRODUCT", "product", product.ID, models.JSONB{"input": input})

	if err := s.Store.SaveCategories(ctx, product.ID, input.CategoryIDs); err != nil {
		logger.ErrorCtx(ctx, "Failed to save categories for product %s: %v", product.ID, err)
	}
	if err := s.Store.SaveStorage(ctx, product.ID, input.StorageItems); err != nil {
		logger.ErrorCtx(ctx, "Failed to save storage for product %s: %v", product.ID, err)
	}
	if input.Category == "store_exclusives" {
		if err := s.Store.SaveDeckCards(ctx, product.ID, input.DeckCards); err != nil {
			logger.ErrorCtx(ctx, "Failed to save deck cards for product %s: %v", product.ID, err)
		}
		product.DeckCards = input.DeckCards
	}

	settings, _ := s.Settings.GetSettings(ctx)
	products := []models.Product{*product}
	
	// Enrichment is highly recommended but non-terminal for a successful POST
	if err := s.EnrichProducts(ctx, products, settings, true); err != nil {
		logger.WarnCtx(ctx, "Partial enrichment failure for product %s: %v", product.ID, err)
	}

	return &products[0], nil
}

func (s *ProductService) Update(ctx context.Context, id string, input models.ProductInput) (*models.Product, error) {
	logger.TraceCtx(ctx, "Entering ProductService.Update | ID: %s", id)
	
	oldProduct, _ := s.Store.GetEnrichedByID(ctx, id)
	
	product, err := s.Store.UpdateProduct(ctx, id, input)
	if err != nil {
		logger.ErrorCtx(ctx, "Core product update failed for %s: %v", id, err)
		return nil, err
	}

	// Logging (Asynchronous/Non-blocking)
	s.Audit.LogAction(ctx, "UPDATE_PRODUCT", "product", id, models.JSONB{
		"before": oldProduct,
		"after":  input,
	})

	if err := s.Store.SaveCategories(ctx, product.ID, input.CategoryIDs); err != nil {
		logger.ErrorCtx(ctx, "Failed to save categories for product %s: %v", product.ID, err)
	}
	if err := s.Store.SaveStorage(ctx, product.ID, input.StorageItems); err != nil {
		logger.ErrorCtx(ctx, "Failed to save storage for product %s: %v", product.ID, err)
	}
	if input.Category == "store_exclusives" {
		if err := s.Store.SaveDeckCards(ctx, product.ID, input.DeckCards); err != nil {
			logger.ErrorCtx(ctx, "Failed to save deck cards for product %s: %v", product.ID, err)
		}
		product.DeckCards = input.DeckCards
	}

	settings, _ := s.Settings.GetSettings(ctx)
	products := []models.Product{*product}
	
	// Enrichment is highly recommended but non-terminal for a successful PUT
	if err := s.EnrichProducts(ctx, products, settings, true); err != nil {
		logger.WarnCtx(ctx, "Partial enrichment failure during update for product %s: %v", product.ID, err)
	}

	return &products[0], nil
}

func (s *ProductService) Delete(ctx context.Context, id string) error {
	logger.TraceCtx(ctx, "Entering ProductService.Delete | ID: %s", id)
	
	oldProduct, _ := s.Store.GetEnrichedByID(ctx, id)
	
	err := s.Store.Delete(ctx, id)
	if err == nil {
		s.Audit.LogAction(ctx, "DELETE_PRODUCT", "product", id, models.JSONB{"deleted": oldProduct})
	}
	return err
}

func (s *ProductService) EnrichProducts(ctx context.Context, products []models.Product, settings models.Settings, isAdmin bool) error {
	if len(products) == 0 {
		return nil
	}

	// 1. Critical Enrichment: Pricing (Impacts all views)
	if err := s.CalculatePrices(products, settings); err != nil {
		return fmt.Errorf("pricing calculation failed: %w", err)
	}

	// 2. Secondary Enrichment: Side-data population
	// We log errors but continue to ensure the basic resource remains accessible
	
	if err := s.Store.PopulateStorage(ctx, products); err != nil {
		logger.ErrorCtx(ctx, "Failed to populate storage info for %d products: %v", len(products), err)
	}
	
	if err := s.Store.PopulateCategories(ctx, products); err != nil {
		logger.ErrorCtx(ctx, "Failed to populate category info for %d products: %v", len(products), err)
	}
	
	if !isAdmin {
		// Filter categories that shouldn't be shown to public
		for i := range products {
			filtered := []models.CustomCategory{}
			for _, c := range products[i].Categories {
				if c.ShowBadge {
					filtered = append(filtered, c)
				}
			}
			products[i].Categories = filtered
		}
	}

	if err := s.Store.PopulateCartCounts(ctx, products); err != nil {
		logger.ErrorCtx(ctx, "Failed to populate cart counts for %d products: %v", len(products), err)
	}
	
	if err := s.IdentifyHotNew(ctx, products, settings); err != nil {
		logger.ErrorCtx(ctx, "Failed to identify hot/new status for %d products: %v", len(products), err)
	}

	return nil
}

func (s *ProductService) CalculatePrices(products []models.Product, settings models.Settings) error {
	for i := range products {
		products[i].Price = products[i].ComputePrice(settings.USDToCOPRate, settings.EURToCOPRate, settings.CKToCOPRate)
	}
	return nil
}

func (s *ProductService) IdentifyHotNew(ctx context.Context, products []models.Product, settings models.Settings) error {
	if len(products) == 0 {
		return nil
	}

	newDays := settings.NewDaysThreshold
	if newDays <= 0 {
		newDays = 10
	}
	newThreshold := time.Now().AddDate(0, 0, -newDays)

	for i := range products {
		if products[i].CreatedAt.After(newThreshold) {
			products[i].IsNew = true
		}
	}

	hotSales := settings.HotSalesThreshold
	if hotSales <= 0 {
		hotSales = 3
	}
	hotDays := settings.HotDaysThreshold
	if hotDays <= 0 {
		hotDays = 7
	}

	pids := make([]string, len(products))
	for i, p := range products {
		pids[i] = p.ID
	}

	hotIDs, err := s.Store.GetHotProductIDs(ctx, hotDays, hotSales, pids)
	if err != nil {
		return err
	}

	hotMap := make(map[string]bool)
	for _, id := range hotIDs {
		hotMap[id] = true
	}

	for i := range products {
		if hotMap[products[i].ID] {
			products[i].IsHot = true
		}
	}

	return nil
}

func (s *ProductService) BulkCreate(ctx context.Context, rawProducts []models.ProductInput, batchCategoryIDs []string) (int, error) {
	logger.TraceCtx(ctx, "Entering ProductService.BulkCreate | Count: %d", len(rawProducts))
	if len(rawProducts) == 0 {
		return 0, nil
	}

	if len(batchCategoryIDs) > 0 {
		for i := range rawProducts {
			if rawProducts[i].CategoryIDs == nil {
				rawProducts[i].CategoryIDs = []string{}
			}

			existing := make(map[string]bool)
			for _, id := range rawProducts[i].CategoryIDs {
				existing[id] = true
			}

			for _, batchID := range batchCategoryIDs {
				if !existing[batchID] {
					rawProducts[i].CategoryIDs = append(rawProducts[i].CategoryIDs, batchID)
				}
			}
		}
	}

	jsonData, err := json.Marshal(rawProducts)
	if err != nil {
		return 0, err
	}

	ids, err := s.Store.BulkUpsert(ctx, string(jsonData))
	if err != nil {
		return 0, err
	}

	return len(ids), nil
}

func (s *ProductService) BulkSearch(ctx context.Context, list string) ([]models.DeckMatch, error) {
	logger.TraceCtx(ctx, "Entering ProductService.BulkSearch | ListLength: %d", len(list))
	settings, _ := s.Settings.GetSettings(ctx)
	lines := strings.Split(list, "\n")
	results := make([]models.DeckMatch, 0)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		parts := strings.SplitN(trimmed, " ", 2)
		qty := 1
		name := trimmed
		if len(parts) > 1 {
			qStr := strings.ToLower(strings.TrimSpace(parts[0]))
			qStr = strings.TrimSuffix(qStr, "x")
			if q, err := strconv.Atoi(qStr); err == nil {
				qty = q
				name = parts[1]
			}
		}

		cleanName := name
		setHint := ""
		cnHint := ""

		if idx := strings.Index(name, "("); idx != -1 {
			cleanName = strings.TrimSpace(name[:idx])
			endIdx := strings.Index(name, ")")
			if endIdx > idx {
				setHint = strings.TrimSpace(name[idx+1 : endIdx])
				after := strings.TrimSpace(name[endIdx+1:])
				if after != "" {
					cnHint = after
				}
			}
		} else {
			lastSpace := strings.LastIndex(name, " ")
			if lastSpace != -1 {
				potentialCN := strings.TrimSpace(name[lastSpace+1:])
				if _, err := strconv.Atoi(potentialCN); err == nil {
					cnHint = potentialCN
					cleanName = strings.TrimSpace(name[:lastSpace])
				}
			}
		}

		var matches []models.Product
		if setHint != "" && cnHint != "" {
			sql := `SELECT * FROM product WHERE (LOWER(name) = LOWER($1) OR name ILIKE $1) 
			        AND (LOWER(set_code) = LOWER($2) OR LOWER(set_name) = LOWER($2))
					AND collector_number = $3 AND stock > 0 ORDER BY stock DESC LIMIT 5`
			_ = s.Store.DB.SelectContext(ctx, &matches, sql, cleanName, setHint, cnHint)
		}

		if len(matches) == 0 && setHint != "" {
			sql := `SELECT * FROM product WHERE (LOWER(name) = LOWER($1) OR name ILIKE $1) 
			        AND (LOWER(set_code) = LOWER($2) OR LOWER(set_name) = LOWER($2))
					AND stock > 0 ORDER BY stock DESC LIMIT 5`
			_ = s.Store.DB.SelectContext(ctx, &matches, sql, cleanName, setHint)
		}

		if len(matches) == 0 {
			sql := `SELECT * FROM product WHERE (LOWER(name) = LOWER($1) OR name ILIKE $1) 
					AND stock > 0 ORDER BY stock DESC LIMIT 5`
			_ = s.Store.DB.SelectContext(ctx, &matches, sql, cleanName)
		}

		if err := s.EnrichProducts(ctx, matches, settings, false); err != nil {
			logger.ErrorCtx(ctx, "Enrichment failed in BulkSearch: %v", err)
		}

		results = append(results, models.DeckMatch{
			RawLine:      trimmed,
			Quantity:     qty,
			Matches:      matches,
			IsMatched:    len(matches) > 0,
			RequestedSet: setHint,
			RequestedCN:  cnHint,
		})
	}
	logger.DebugCtx(ctx, "BulkSearch matched %d/%d lines", func() int {
		matched := 0
		for _, r := range results {
			if r.IsMatched {
				matched++
			}
		}
		return matched
	}(), len(results))
	return results, nil
}

func (s *ProductService) GetLowStock(ctx context.Context, threshold int) ([]models.Product, error) {
	var products []models.Product
	err := s.Store.DB.SelectContext(ctx, &products, "SELECT * FROM product WHERE stock <= $1 AND stock > 0 ORDER BY stock ASC LIMIT 100", threshold)
	if err != nil {
		return nil, err
	}

	settings, _ := s.Settings.GetSettings(ctx)
	if err := s.EnrichProducts(ctx, products, settings, true); err != nil {
		return nil, err
	}

	return products, nil
}

func (s *ProductService) GetTCGs(ctx context.Context, activeOnly bool) ([]models.TCG, error) {
	return s.TCGStore.ListWithCount(ctx, activeOnly)
}

func (s *ProductService) GetStorage(ctx context.Context, id string) ([]models.StorageLocation, error) {
	var items []models.StorageLocation
	err := s.Store.DB.SelectContext(ctx, &items, `
		SELECT s.id as storage_id, s.name, COALESCE(ps.quantity, 0) as quantity
		FROM storage_location s
		LEFT JOIN product_storage ps ON s.id = ps.storage_id AND ps.product_id = $1
		ORDER BY s.name
	`, id)
	return items, err
}

func (h *ProductService) UpdateStorage(ctx context.Context, id string, updates []models.ProductStorage) ([]models.StorageLocation, error) {
	tx, err := h.Store.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM product_storage WHERE product_id = $1`, id)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	var validUpdates []models.ProductStorage
	for _, u := range updates {
		if u.Quantity > 0 {
			validUpdates = append(validUpdates, u)
		}
	}

	if len(validUpdates) > 0 {
		query := "INSERT INTO product_storage (product_id, storage_id, quantity) VALUES "
		values := make([]interface{}, 0, len(validUpdates)*3)
		placeholders := make([]string, 0, len(validUpdates))

		for i, u := range validUpdates {
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
			values = append(values, id, u.StorageID, u.Quantity)
		}

		query += strings.Join(placeholders, ", ")
		_, err = tx.ExecContext(ctx, query, values...)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	
	return h.GetStorage(ctx, id)
}
func (s *ProductService) BulkUpdateSource(ctx context.Context, ids []string, source models.PriceSource) (int, error) {
	logger.TraceCtx(ctx, "Entering ProductService.BulkUpdateSource | Count: %d | Source: %s", len(ids), source)
	if len(ids) == 0 {
		return 0, nil
	}

	// 1. Update source in DB
	query, args, err := sqlx.In(`UPDATE product SET price_source = ?, updated_at = now() WHERE id IN (?)`, source, ids)
	if err != nil {
		return 0, err
	}
	res, err := s.Store.DB.ExecContext(ctx, s.Store.DB.Rebind(query), args...)
	if err != nil {
		return 0, err
	}
	count, _ := res.RowsAffected()

	// 2. Trigger price refresh for these specific products
	var rows []store.RefreshRow
	query, args, err = sqlx.In(`
		SELECT 
			p.id, p.tcg, p.name, p.set_name, p.set_code, p.collector_number,
			p.foil_treatment, p.card_treatment, p.price_source, p.scryfall_id,
			s.ck_name as ck_set_name
		FROM product p
		LEFT JOIN tcg_set s ON p.tcg = s.tcg AND p.set_code = s.code
		WHERE p.id IN (?)`, ids)
	if err == nil {
		_ = s.Store.DB.SelectContext(ctx, &rows, s.Store.DB.Rebind(query), args...)
		
		if len(rows) > 0 {
			// Resolve external prices
			scryIDs := make([]string, 0, len(rows))
			needsCK := false
			for _, row := range rows {
				if (row.PriceSource == "tcgplayer" || row.PriceSource == "cardmarket") && row.ScryfallID != "" {
					scryIDs = append(scryIDs, row.ScryfallID)
				}
				if row.PriceSource == "cardkingdom" {
					needsCK = true
				}
			}

			// Fetch external data
			scryBatch, _ := external.BatchLookupMTG(ctx, scryIDs)
			var ckPriceMap map[string]*float64
			if needsCK {
				ckPriceMap, _ = external.BuildCardKingdomPriceMap(ctx)
			}
			
			// Map results
			// For TCGPlayer/Cardmarket we use the scryBatch directly, for CK we use BuildPriceUpdates
			updates := make([]store.MetadataUpdate, 0, len(rows))
			
			// Separated CK rows for specialized matching
			var ckRows []store.RefreshRow
			
			for _, row := range rows {
				if row.PriceSource == "cardkingdom" {
					ckRows = append(ckRows, row)
					continue
				}

				// TCGPlayer / Cardmarket logic
				var price *float64
				var scryID string
				var oracleText string
				var typeLine string
				var imageURL string
				var legalities models.JSONB

				if meta, ok := scryBatch[row.ScryfallID]; ok {
					scryID = meta.ScryfallID
					oracleText = meta.OracleText
					typeLine = meta.TypeLine
					legalities = meta.Legalities
					imageURL = meta.ImageURL
					if row.PriceSource == "tcgplayer" {
						price = meta.TCGPlayerUSD
					} else {
						price = meta.CardmarketEUR
					}
				}

				if price != nil || scryID != "" {
					updates = append(updates, store.MetadataUpdate{
						ID:          row.ID,
						Price:       price,
						ScryfallID:  scryID,
						OracleText:  oracleText,
						TypeLine:    typeLine,
						ImageURL:    imageURL,
						Legalities:  legalities,
						PriceSource: row.PriceSource,
					})
				}
			}

			// Add CK updates — BuildPriceUpdates handles edition+variation matching using
			// the ck_set_name joined from tcg_set. Pass nil for the scry price map since
			// the scryBatch type (keyed by Scryfall ID) differs from the PriceKey map it expects.
			if len(ckRows) > 0 && ckPriceMap != nil {
				ckUpdates, _ := store.BuildPriceUpdates(ckRows, nil, ckPriceMap)
				updates = append(updates, ckUpdates...)
			}

			if len(updates) > 0 {
				rs := store.RefreshStore{DB: s.Store.DB}
				settings, _ := s.Settings.GetSettings(ctx)
				_, _ = rs.BulkUpdateMetadata(ctx, updates, settings.USDToCOPRate, settings.EURToCOPRate, settings.CKToCOPRate)
			}
		}
	}

	s.Audit.LogAction(ctx, "BULK_UPDATE_SOURCE", "product", "batch", models.JSONB{
		"ids":    ids,
		"source": source,
		"count":  count,
	})

	return int(count), nil
}

func (s *ProductService) EnrichCardLookupResults(ctx context.Context, results []*external.CardLookupResult) error {
	if len(results) == 0 {
		return nil
	}

	// 1. Resolve CK mapping if any MTG cards are present
	var mtgResults []*external.CardLookupResult
	for _, r := range results {
		if r.MTGMetadata.SetCode != nil {
			mtgResults = append(mtgResults, r)
		}
	}

	if len(mtgResults) == 0 {
		return nil
	}

	// Load CK Price Map
	ckPriceMap, err := external.BuildCardKingdomPriceMap(ctx)
	if err != nil {
		logger.WarnCtx(ctx, "Failed to load CK price map for enrichment: %v", err)
		return nil // Non-fatal, just no CK prices
	}

	// Load Set Mappings for ck_name
	var sets []models.TCGSet
	if s.TCGStore != nil {
		sets, err = s.TCGStore.ListSets(ctx, "mtg")
		if err != nil {
			logger.WarnCtx(ctx, "Failed to load MTG sets for CK enrichment: %v", err)
			// We can still try to match by default names
		}
	}

	ckNameMap := make(map[string]string)
	for _, set := range sets {
		if set.CKName != nil && *set.CKName != "" && set.Code != "" {
			ckNameMap[strings.ToLower(set.Code)] = *set.CKName
		}
	}

	// Match results
	for _, r := range mtgResults {
		setCode := strings.ToLower(*r.MTGMetadata.SetCode)

		// Prefer curated ck_name from DB; fall back to Scryfall set name
		ckEdition := ""
		if mapped, ok := ckNameMap[setCode]; ok {
			ckEdition = mapped
		} else if r.MTGMetadata.SetName != nil {
			ckEdition = external.NormalizeCKEdition(*r.MTGMetadata.SetName)
		}

		isFoil := r.MTGMetadata.FoilTreatment != models.FoilNonFoil
		variation := external.MapFoilTreatmentToCKVariation(r.MTGMetadata.FoilTreatment, r.MTGMetadata.CardTreatment)

		r.PriceCardKingdom = external.LookupCKPrice(r.Name, ckEdition, variation, isFoil, ckPriceMap)
	}

	return nil
}
