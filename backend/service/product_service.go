package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"context"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/el-bulk/backend/utils/logger"
)

type ProductService struct {
	Store    *store.ProductStore
	TCGStore *store.TCGStore
	Settings *SettingsService
	Audit    *AuditService
}

func NewProductService(s *store.ProductStore, tcg *store.TCGStore, settings *SettingsService, audit *AuditService) *ProductService {
	return &ProductService{
		Store:    s,
		TCGStore: tcg,
		Settings: settings,
		Audit:    audit,
	}
}

func (s *ProductService) List(params store.ProductFilterParams, isAdmin bool) (models.ProductListResponse, error) {
	logger.Trace("Entering ProductService.List | Admin: %v | Params: %+v", isAdmin, params)
	products, total, err := s.Store.ListWithFilters(params)
	if err != nil {
		return models.ProductListResponse{}, err
	}

	settings, err := s.Settings.GetSettings()
	if err != nil {
		logger.Error("Failed to get settings in ProductService.List: %v", err)
	}

	if err := s.EnrichProducts(products, settings, isAdmin); err != nil {
		return models.ProductListResponse{}, err
	}

	facets, err := s.Store.GetFacets(params, isAdmin)
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

func (s *ProductService) GetByID(id string, isAdmin bool) (*models.Product, error) {
	logger.Trace("Entering ProductService.GetByID | ID: %s | Admin: %v", id, isAdmin)
	product, err := s.Store.GetEnrichedByID(id)
	if err != nil {
		return nil, err
	}

	settings, err := s.Settings.GetSettings()
	if err != nil {
		logger.Error("Failed to get settings in ProductService.GetByID: %v", err)
	}

	products := []models.Product{*product}
	if err := s.CalculatePrices(products, settings); err != nil {
		return nil, err
	}
	if err := s.Store.PopulateCartCounts(products); err != nil {
		return nil, err
	}
	
	if isAdmin {
		if err := s.Store.PopulateStorage(products); err != nil {
			return nil, err
		}
	} else {
		// Filter categories for public view
		if err := s.Store.PopulateCategories(products); err != nil {
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
	logger.Trace("Entering ProductService.Create | Name: %s", input.Name)
	product, err := s.Store.CreateProduct(input)
	if err != nil {
		return nil, err
	}

	// Logging
	s.Audit.LogAction(ctx, "CREATE_PRODUCT", "product", product.ID, models.JSONB{"input": input})

	if err := s.Store.SaveCategories(product.ID, input.CategoryIDs); err != nil {
		logger.Error("Failed to save categories for product %s: %v", product.ID, err)
	}
	if err := s.Store.SaveStorage(product.ID, input.StorageItems); err != nil {
		logger.Error("Failed to save storage for product %s: %v", product.ID, err)
	}
	if input.Category == "store_exclusives" {
		if err := s.Store.SaveDeckCards(product.ID, input.DeckCards); err != nil {
			logger.Error("Failed to save deck cards for product %s: %v", product.ID, err)
		}
		product.DeckCards = input.DeckCards
	}

	settings, _ := s.Settings.GetSettings()
	products := []models.Product{*product}
	if err := s.EnrichProducts(products, settings, true); err != nil {
		return nil, err
	}

	return &products[0], nil
}

func (s *ProductService) Update(ctx context.Context, id string, input models.ProductInput) (*models.Product, error) {
	logger.Trace("Entering ProductService.Update | ID: %s", id)
	
	oldProduct, _ := s.Store.GetEnrichedByID(id)
	
	product, err := s.Store.UpdateProduct(id, input)
	if err != nil {
		return nil, err
	}

	// Logging
	s.Audit.LogAction(ctx, "UPDATE_PRODUCT", "product", id, models.JSONB{
		"before": oldProduct,
		"after":  input,
	})

	if err := s.Store.SaveCategories(product.ID, input.CategoryIDs); err != nil {
		logger.Error("Failed to save categories for product %s: %v", product.ID, err)
	}
	if err := s.Store.SaveStorage(product.ID, input.StorageItems); err != nil {
		logger.Error("Failed to save storage for product %s: %v", product.ID, err)
	}
	if input.Category == "store_exclusives" {
		if err := s.Store.SaveDeckCards(product.ID, input.DeckCards); err != nil {
			logger.Error("Failed to save deck cards for product %s: %v", product.ID, err)
		}
		product.DeckCards = input.DeckCards
	}

	settings, _ := s.Settings.GetSettings()
	products := []models.Product{*product}
	if err := s.EnrichProducts(products, settings, true); err != nil {
		return nil, err
	}

	return &products[0], nil
}

func (s *ProductService) Delete(ctx context.Context, id string) error {
	logger.Trace("Entering ProductService.Delete | ID: %s", id)
	
	oldProduct, _ := s.Store.GetEnrichedByID(id)
	
	err := s.Store.Delete(id)
	if err == nil {
		s.Audit.LogAction(ctx, "DELETE_PRODUCT", "product", id, models.JSONB{"deleted": oldProduct})
	}
	return err
}

func (s *ProductService) EnrichProducts(products []models.Product, settings models.Settings, isAdmin bool) error {
	if len(products) == 0 {
		return nil
	}

	if err := s.CalculatePrices(products, settings); err != nil {
		return err
	}
	if err := s.Store.PopulateStorage(products); err != nil {
		return err
	}
	if err := s.Store.PopulateCategories(products); err != nil {
		return err
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

	if err := s.Store.PopulateCartCounts(products); err != nil {
		return err
	}
	if err := s.IdentifyHotNew(products, settings); err != nil {
		return err
	}

	return nil
}

func (s *ProductService) CalculatePrices(products []models.Product, settings models.Settings) error {
	for i := range products {
		products[i].Price = products[i].ComputePrice(settings.USDToCOPRate, settings.EURToCOPRate)
	}
	return nil
}

func (s *ProductService) IdentifyHotNew(products []models.Product, settings models.Settings) error {
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

	hotIDs, err := s.Store.GetHotProductIDs(hotDays, hotSales, pids)
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

func (s *ProductService) BulkCreate(rawProducts []models.ProductInput, batchCategoryIDs []string) (int, error) {
	logger.Trace("Entering ProductService.BulkCreate | Count: %d", len(rawProducts))
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

	ids, err := s.Store.BulkUpsert(string(jsonData))
	if err != nil {
		return 0, err
	}

	return len(ids), nil
}

func (s *ProductService) BulkSearch(list string) ([]models.DeckMatch, error) {
	logger.Trace("Entering ProductService.BulkSearch | ListLength: %d", len(list))
	settings, _ := s.Settings.GetSettings()
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
			_ = s.Store.DB.Select(&matches, sql, cleanName, setHint, cnHint)
		}

		if len(matches) == 0 && setHint != "" {
			sql := `SELECT * FROM product WHERE (LOWER(name) = LOWER($1) OR name ILIKE $1) 
			        AND (LOWER(set_code) = LOWER($2) OR LOWER(set_name) = LOWER($2))
					AND stock > 0 ORDER BY stock DESC LIMIT 5`
			_ = s.Store.DB.Select(&matches, sql, cleanName, setHint)
		}

		if len(matches) == 0 {
			sql := `SELECT * FROM product WHERE (LOWER(name) = LOWER($1) OR name ILIKE $1) 
					AND stock > 0 ORDER BY stock DESC LIMIT 5`
			_ = s.Store.DB.Select(&matches, sql, cleanName)
		}

		if err := s.EnrichProducts(matches, settings, false); err != nil {
			logger.Error("Enrichment failed in BulkSearch: %v", err)
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
	logger.Debug("BulkSearch matched %d/%d lines", func() int {
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

func (s *ProductService) GetLowStock(threshold int) ([]models.Product, error) {
	var products []models.Product
	err := s.Store.DB.Select(&products, "SELECT * FROM product WHERE stock <= $1 AND stock > 0 ORDER BY stock ASC LIMIT 100", threshold)
	if err != nil {
		return nil, err
	}

	settings, _ := s.Settings.GetSettings()
	if err := s.EnrichProducts(products, settings, true); err != nil {
		return nil, err
	}

	return products, nil
}

func (s *ProductService) GetTCGs(activeOnly bool) ([]models.TCG, error) {
	return s.TCGStore.ListWithCount(activeOnly)
}

func (s *ProductService) GetStorage(id string) ([]models.StorageLocation, error) {
	var items []models.StorageLocation
	err := s.Store.DB.Select(&items, `
		SELECT s.id as storage_id, s.name, COALESCE(ps.quantity, 0) as quantity
		FROM storage_location s
		LEFT JOIN product_storage ps ON s.id = ps.storage_id AND ps.product_id = $1
		ORDER BY s.name
	`, id)
	return items, err
}

func (s *ProductService) UpdateStorage(id string, updates []models.ProductStorage) ([]models.StorageLocation, error) {
	tx, err := s.Store.DB.Beginx()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`DELETE FROM product_storage WHERE product_id = $1`, id)
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
		_, err = tx.Exec(query, values...)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	
	return s.GetStorage(id)
}
