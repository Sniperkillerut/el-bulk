package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/cache"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

type SearchResponse struct {
	Products []models.Product
	Total    int
}

type ProductStore struct {
	*BaseStore[models.Product]
	facetCache  *cache.TTLMap[models.Facets]
	searchCache *cache.TTLMap[SearchResponse]
}

func NewProductStore(db *sqlx.DB) *ProductStore {
	return &ProductStore{
		BaseStore:   NewBaseStore[models.Product](db, "product"),
		facetCache:  cache.NewTTLMap[models.Facets](5 * time.Minute),
		searchCache: cache.NewTTLMap[SearchResponse](2 * time.Minute),
	}
}

type ProductFilterParams struct {
	TCG            string
	Category       string
	Search         string
	StorageID      string
	Foil           string
	Treatment      string
	Condition      string
	Collection     string
	Rarity         string
	Language       string
	Color          string
	SetName        string
	InStock        bool
	SortBy         string
	SortDir        string
	OnlyDuplicates bool
	FilterLogic    string
	Page           int
	PageSize       int
	Offset         int

	IDs []string

	// MTG Metadata Filters
	IsLegendary    string
	IsLand         string
	IsHistoric     string
	IsPrepared     string
	LandType       string
	Format         string
	FrameEffects   string
	CardTypes      string
	FullArt        string
	Textless       string

	// Exchange rates for on-the-fly price sorting
	USDRate float64
	EURRate float64
	CKRate  float64
}

func (s *ProductStore) ListWithFilters(ctx context.Context, params ProductFilterParams) ([]models.Product, int, error) {
	if s.DB == nil {
		return nil, 0, fmt.Errorf("database connection is not initialized")
	}

	// Try cache for common pages (head of the catalog)
	cacheKey := s.GetSearchCacheKey(params)
	if params.Page <= 5 {
		if cached, found := s.searchCache.Get(cacheKey); found {
			return cached.Products, cached.Total, nil
		}
	}

	start := time.Now()
	fromClause, where, args := s.BuildFilters(ctx, params)

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) %s %s", fromClause, where)
	if err := s.DB.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	orderBy := s.BuildOrderBy(params, len(args))
	viewFrom, where, args := s.BuildFilters(ctx, params)

	listQuery := fmt.Sprintf("SELECT p.* %s %s ORDER BY %s LIMIT $%d OFFSET $%d",
		viewFrom, where, orderBy, len(args)+1, len(args)+2)

	listArgs := append([]interface{}{}, args...)
	listArgs = append(listArgs, params.PageSize, params.Offset)

	products, _, err := s.SelectEnriched(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}

	// Cache common pages
	if params.Page <= 5 {
		s.searchCache.Set(cacheKey, SearchResponse{Products: products, Total: total})
	}

	logger.DebugCtx(ctx, "[DB] ListWithFilters (count+list) took %v", time.Since(start))
	return products, total, nil
}

func (s *ProductStore) GetSearchCacheKey(p ProductFilterParams) string {
	// Simple string key for performance, including all parameters that affect the query
	return fmt.Sprintf("search:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s:%v:%v:%s:%s:%d:%d:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s",
		p.TCG, p.Category, p.Search, p.StorageID, p.Foil, p.Treatment, p.Condition, p.Collection, p.Rarity, p.Language,
		p.Color, p.SetName, p.InStock, p.OnlyDuplicates, p.SortBy, p.SortDir, p.Page, p.PageSize,
		p.IsLegendary, p.IsLand, p.IsHistoric, p.IsPrepared, p.LandType, p.Format, p.FrameEffects, p.CardTypes,
		p.FullArt, p.Textless, p.FilterLogic)
}

func (s *ProductStore) InvalidateCaches() {
	s.facetCache.Flush()
	s.searchCache.Flush()
}

func (s *ProductStore) Update(ctx context.Context, id string, updates map[string]interface{}) (*models.Product, error) {
	p, err := s.BaseStore.Update(ctx, id, updates)
	if err == nil {
		s.InvalidateCaches()
	}
	return p, err
}

func (s *ProductStore) Delete(ctx context.Context, id string) error {
	err := s.BaseStore.Delete(ctx, id)
	if err == nil {
		s.InvalidateCaches()
	}
	return err
}

func (s *ProductStore) SelectEnriched(ctx context.Context, query string, args ...interface{}) ([]models.Product, int, error) {
	var products []models.Product
	if err := s.DB.Unsafe().SelectContext(ctx, &products, query, args...); err != nil {
		return nil, 0, err
	}

	if len(products) > 0 {
		var wg sync.WaitGroup
		wg.Add(4)

		go func() {
			defer wg.Done()
			s.PopulateStorage(ctx, products)
		}()
		go func() {
			defer wg.Done()
			s.PopulateCategories(ctx, products)
		}()
		go func() {
			defer wg.Done()
			s.PopulateCartCounts(ctx, products)
		}()
		go func() {
			defer wg.Done()
			s.PopulateDeckCards(ctx, products)
		}()

		wg.Wait()
	}

	return products, len(products), nil
}

func (s *ProductStore) GetFacets(ctx context.Context, params ProductFilterParams, isAdmin bool) (models.Facets, error) {
	if s.DB == nil {
		return models.Facets{}, fmt.Errorf("database connection is not initialized")
	}
	// Generate cache key from params - must include ALL fields that affect the query
	cacheKey := fmt.Sprintf("facets:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v",
		params.TCG, params.Category, params.Search, params.StorageID, params.Foil, params.Treatment, params.Condition,
		params.Collection, params.Rarity, params.Language, params.Color, params.SetName, params.InStock, params.OnlyDuplicates,
		params.FilterLogic, isAdmin, params.IsLegendary, params.IsLand, params.IsHistoric, params.IsPrepared,
		params.LandType, params.Format, params.FrameEffects, params.CardTypes, params.FullArt, params.Textless)

	if cached, ok := s.facetCache.Get(cacheKey); ok {
		logger.TraceCtx(ctx, "[CACHE] Facet hit for key: %s", cacheKey)
		return cached, nil
	}

	var result []byte
	start := time.Now()
	query := "SELECT fn_get_product_facets($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)"
	logger.TraceCtx(ctx, "[DB] Executing GetFacets: %s", query)
	err := s.DB.GetContext(ctx, &result, query,
		params.TCG, params.Category, params.Search, params.StorageID, params.Foil, params.Treatment, params.Condition,
		params.Rarity, params.Language, params.Color, params.Collection, params.SetName, params.InStock, params.FilterLogic, isAdmin,
		params.IsLegendary, params.IsLand, params.IsHistoric, params.IsPrepared, params.Format, params.FrameEffects, params.CardTypes)

	if err != nil {
		logger.ErrorCtx(ctx, "[DB] GetFacets failed: %v", err)
		return models.Facets{}, err
	}
	logger.DebugCtx(ctx, "[DB] GetFacets took %v", time.Since(start))

	var facets models.Facets
	if err := json.Unmarshal(result, &facets); err != nil {
		return models.Facets{}, err
	}

	// Cache the result (shared across all users with same filter set)
	s.facetCache.Set(cacheKey, facets, 5*time.Minute)

	return facets, nil
}

func (s *ProductStore) PopulateStorage(ctx context.Context, products []models.Product) error {
	if len(products) == 0 {
		return nil
	}
	pids := make([]string, 0, len(products))
	for _, p := range products {
		if p.ID != "" {
			pids = append(pids, p.ID)
		}
	}

	if len(pids) == 0 {
		return nil
	}

	query, args, err := sqlx.In(`
		SELECT ps.product_id, s.id as stored_in_id, s.name, ps.quantity 
		FROM product_storage ps 
		JOIN storage_location s ON ps.storage_id = s.id 
		WHERE ps.quantity > 0 AND ps.product_id IN (?)
	`, pids)
	if err != nil {
		return err
	}

	query = s.DB.Rebind(query)
	var storageRows []struct {
		ProductID string `db:"product_id"`
		StorageID string `db:"stored_in_id"`
		Name      string `db:"name"`
		Quantity  int    `db:"quantity"`
	}
	if err := s.DB.SelectContext(ctx, &storageRows, query, args...); err != nil {
		return err
	}

	storageMap := make(map[string][]models.StorageLocation)
	for _, r := range storageRows {
		storageMap[r.ProductID] = append(storageMap[r.ProductID], models.StorageLocation{
			StorageID: r.StorageID,
			Name:      r.Name,
			Quantity:  r.Quantity,
		})
	}

	var emptyLocs []models.StorageLocation
	for i := range products {
		if locs, ok := storageMap[products[i].ID]; ok {
			products[i].StoredIn = locs
		} else {
			products[i].StoredIn = emptyLocs
		}
	}
	return nil
}

func (s *ProductStore) PopulateCategories(ctx context.Context, products []models.Product) error {
	if len(products) == 0 {
		return nil
	}
	pids := make([]string, 0, len(products))
	for _, p := range products {
		if p.ID != "" {
			pids = append(pids, p.ID)
		}
	}

	if len(pids) == 0 {
		return nil
	}

	sql := `
		SELECT pc.product_id, c.id, c.name, c.slug, c.show_badge, c.is_active, c.searchable, c.bg_color, c.text_color, c.icon
		FROM product_category pc
		JOIN custom_category c ON pc.category_id = c.id
		WHERE pc.product_id IN (?)
	`
	sql += " ORDER BY c.name "

	query, args, err := sqlx.In(sql, pids)
	if err != nil {
		return err
	}

	query = s.DB.Rebind(query)
	var catRows []struct {
		ProductID  string  `db:"product_id"`
		ID         string  `db:"id"`
		Name       string  `db:"name"`
		Slug       string  `db:"slug"`
		ShowBadge  bool    `db:"show_badge"`
		IsActive   bool    `db:"is_active"`
		Searchable bool    `db:"searchable"`
		BgColor    *string `db:"bg_color"`
		TextColor  *string `db:"text_color"`
		Icon       *string `db:"icon"`
	}
	if err := s.DB.SelectContext(ctx, &catRows, query, args...); err != nil {
		return err
	}

	catMap := make(map[string][]models.CustomCategory)
	for _, r := range catRows {
		catMap[r.ProductID] = append(catMap[r.ProductID], models.CustomCategory{
			ID:         r.ID,
			Name:       r.Name,
			Slug:       r.Slug,
			ShowBadge:  r.ShowBadge,
			IsActive:   r.IsActive,
			Searchable: r.Searchable,
			BgColor:    r.BgColor,
			TextColor:  r.TextColor,
			Icon:       r.Icon,
		})
	}

	var emptyCats []models.CustomCategory
	for i := range products {
		if cats, ok := catMap[products[i].ID]; ok {
			products[i].Categories = cats
		} else {
			products[i].Categories = emptyCats
		}
	}
	return nil
}

func (s *ProductStore) PopulateDeckCards(ctx context.Context, products []models.Product) error {
	if len(products) == 0 {
		return nil
	}
	pids := make([]string, 0, len(products))
	for _, p := range products {
		if p.ID != "" {
			pids = append(pids, p.ID)
		}
	}

	if len(pids) == 0 {
		return nil
	}

	query, args, err := sqlx.In(`
		SELECT product_id, id, name, set_code, collector_number, quantity, type_line, image_url, foil_treatment, card_treatment, rarity, art_variation, scryfall_id, frame_effects
		FROM deck_card 
		WHERE product_id IN (?)
	`, pids)
	if err != nil {
		return err
	}

	query = s.DB.Rebind(query)
	var deckRows []struct {
		models.DeckCard
		ProductID string `db:"product_id"`
	}
	if err := s.DB.SelectContext(ctx, &deckRows, query, args...); err != nil {
		return err
	}

	deckMap := make(map[string][]models.DeckCard)
	for _, r := range deckRows {
		deckMap[r.ProductID] = append(deckMap[r.ProductID], r.DeckCard)
	}

	for i := range products {
		if cards, ok := deckMap[products[i].ID]; ok {
			products[i].DeckCards = cards
		} else {
			products[i].DeckCards = []models.DeckCard{}
		}
	}
	return nil
}

func (s *ProductStore) PopulateCartCounts(ctx context.Context, products []models.Product) error {
	if len(products) == 0 {
		return nil
	}
	pids := make([]string, 0, len(products))
	for _, p := range products {
		if p.ID != "" {
			pids = append(pids, p.ID)
		}
	}

	if len(pids) == 0 {
		return nil
	}

	query, args, err := sqlx.In(`
		SELECT oi.product_id, COUNT(DISTINCT o.customer_id) as cart_count
		FROM "order" o
		JOIN order_item oi ON o.id = oi.order_id
		WHERE o.status = 'pending' AND oi.product_id IN (?)
		GROUP BY oi.product_id
	`, pids)
	if err != nil {
		return err
	}

	query = s.DB.Rebind(query)
	var countRows []struct {
		ProductID string `db:"product_id"`
		CartCount int    `db:"cart_count"`
	}
	if err := s.DB.SelectContext(ctx, &countRows, query, args...); err != nil {
		return err
	}

	countMap := make(map[string]int)
	for _, r := range countRows {
		countMap[r.ProductID] = r.CartCount
	}

	for i := range products {
		products[i].CartCount = countMap[products[i].ID]
	}
	return nil
}

func (s *ProductStore) GetHotProductIDs(ctx context.Context, hotDays, hotSales int, candidateIDs []string) ([]string, error) {
	if len(candidateIDs) == 0 {
		return []string{}, nil
	}

	// Filter out empty IDs to prevent syntax errors
	validIDs := make([]string, 0, len(candidateIDs))
	for _, id := range candidateIDs {
		if id != "" {
			validIDs = append(validIDs, id)
		}
	}

	if len(validIDs) == 0 {
		return []string{}, nil
	}

	query, args, err := sqlx.In(fmt.Sprintf(`
		SELECT product_id
		FROM order_item oi
		JOIN "order" o ON oi.order_id = o.id
		WHERE o.created_at > (now() - interval '%d days')
		  AND product_id IN (?)
		GROUP BY product_id
		HAVING SUM(quantity) >= %d
	`, hotDays, hotSales), validIDs)

	if err != nil {
		return nil, err
	}

	var hotIDs []string
	err = s.DB.SelectContext(ctx, &hotIDs, s.DB.Rebind(query), args...)
	return hotIDs, err
}

func (s *ProductStore) SaveCategories(ctx context.Context, productID string, categoryIDs []string) error {
	// s.facetCache.Clear() - cache expires naturally via TTL
	_, err := s.DB.ExecContext(ctx, "DELETE FROM product_category WHERE product_id = $1", productID)
	if err != nil {
		return err
	}
	for _, cid := range categoryIDs {
		_, err := s.DB.ExecContext(ctx, "INSERT INTO product_category (product_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", productID, cid)
		if err != nil {
			logger.ErrorCtx(ctx, "Error inserting product_category (product=%s, cat=%s): %v", productID, cid, err)
		}
	}
	if err == nil {
		s.InvalidateCaches()
	}
	return nil
}

func (s *ProductStore) SaveDeckCards(ctx context.Context, productID string, cards []models.DeckCard) error {
	// s.facetCache.Clear() - cache expires naturally via TTL
	_, err := s.DB.ExecContext(ctx, "DELETE FROM deck_card WHERE product_id = $1", productID)
	if err != nil {
		return err
	}

	if len(cards) == 0 {
		return nil
	}

	query := "INSERT INTO deck_card (product_id, name, set_code, collector_number, quantity, type_line, image_url, foil_treatment, card_treatment, rarity, art_variation, scryfall_id, frame_effects) VALUES "
	values := make([]interface{}, 0, len(cards)*13)
	placeholders := make([]string, 0, len(cards))

	for i, c := range cards {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*13+1, i*13+2, i*13+3, i*13+4, i*13+5, i*13+6, i*13+7, i*13+8, i*13+9, i*13+10, i*13+11, i*13+12, i*13+13))
		values = append(values, productID, c.Name, c.SetCode, c.CollectorNumber, c.Quantity, c.TypeLine, c.ImageURL, c.FoilTreatment, c.CardTreatment, c.Rarity, c.ArtVariation, c.ScryfallID, c.FrameEffects)
	}

	query += strings.Join(placeholders, ", ")
	logger.TraceCtx(ctx, "[DB] Executing SaveDeckCards for product %s: %s | Values: %d", productID, query, len(values))
	_, err = s.DB.ExecContext(ctx, query, values...)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] SaveDeckCards failed for product %s: %v", productID, err)
	}
	if err == nil {
		s.InvalidateCaches()
	}
	return err
}

func (s *ProductStore) SaveStorage(ctx context.Context, productID string, items []models.StorageLocation) error {
	// s.facetCache.Clear() - cache expires naturally via TTL
	_, err := s.DB.ExecContext(ctx, "DELETE FROM product_storage WHERE product_id = $1", productID)
	if err != nil {
		return err
	}

	var validItems []models.StorageLocation
	for _, item := range items {
		if item.Quantity > 0 {
			validItems = append(validItems, item)
		}
	}

	if len(validItems) == 0 {
		return nil
	}

	query := "INSERT INTO product_storage (product_id, storage_id, quantity) VALUES "
	values := make([]interface{}, 0, len(validItems)*3)
	placeholders := make([]string, 0, len(validItems))

	for i, item := range validItems {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
		values = append(values, productID, item.StorageID, item.Quantity)
	}

	query += strings.Join(placeholders, ", ")
	logger.TraceCtx(ctx, "[DB] Executing SaveStorage for product %s: %s | Values: %d", productID, query, len(values))
	_, err = s.DB.ExecContext(ctx, query, values...)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] SaveStorage failed for product %s: %v", productID, err)
	}
	if err == nil {
		s.InvalidateCaches()
	}
	return err
}

func (s *ProductStore) CreateProduct(ctx context.Context, input models.ProductInput) (*models.Product, error) {
	if input.FoilTreatment == "" {
		input.FoilTreatment = models.FoilNonFoil
	}
	if input.CardTreatment == "" {
		input.CardTreatment = models.TreatmentNormal
	}
	if input.PriceSource == "" {
		input.PriceSource = models.PriceSourceManual
	}

	var product models.Product
	// s.facetCache.Clear() - cache expires naturally via TTL
	query := `INSERT INTO product (
		name, tcg, category, set_name, set_code, condition,
		foil_treatment, card_treatment,
		price_reference, price_source, price_cop_override,
		stock, cost_basis_cop, image_url, description, collector_number, promo_type,
		language, color_identity, rarity, cmc, is_legendary, is_historic, is_land, is_basic_land, art_variation,
		oracle_text, artist, type_line, border_color, frame, full_art, textless, scryfall_id, frame_effects
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35
	) RETURNING *`

	err := s.DB.QueryRowxContext(ctx, query,
		input.Name, input.TCG, input.Category, input.SetName, input.SetCode, input.Condition,
		input.FoilTreatment, input.CardTreatment,
		input.PriceReference, input.PriceSource, input.PriceCOPOverride,
		input.Stock, input.CostBasis, input.ImageURL, input.Description, input.CollectorNumber, input.PromoType,
		input.Language, input.ColorIdentity, input.Rarity, input.CMC, input.IsLegendary, input.IsHistoric, input.IsLand, input.IsBasicLand, input.ArtVariation,
		input.OracleText, input.Artist, input.TypeLine, input.BorderColor, input.Frame, input.FullArt, input.Textless, input.ScryfallID, input.FrameEffects,
	).StructScan(&product)

	logger.DebugCtx(ctx, "[DB] CreateProduct result: %+v | Error: %v", product.ID, err)

	if err == nil {
		s.InvalidateCaches()
	}

	return &product, err
}

func (s *ProductStore) UpdateProduct(ctx context.Context, id string, input models.ProductInput) (*models.Product, error) {
	if input.PriceSource == "" {
		input.PriceSource = models.PriceSourceManual
	}

	var product models.Product
	// s.facetCache.Clear() - cache expires naturally via TTL
	query := `UPDATE product SET
		name=$1, tcg=$2, category=$3, set_name=$4, set_code=$5, condition=$6,
		foil_treatment=$7, card_treatment=$8,
		price_reference=$9, price_source=$10, price_cop_override=$11,
		stock=$12, cost_basis_cop=$13, image_url=$14, description=$15, collector_number=$16, promo_type=$17,
		language=$18, color_identity=$19, rarity=$20, cmc=$21, is_legendary=$22, is_historic=$23, is_land=$24, is_basic_land=$25, art_variation=$26,
		oracle_text=$27, artist=$28, type_line=$29, border_color=$30, frame=$31, full_art=$32, textless=$33, scryfall_id=$34, frame_effects=$35
	WHERE id=$36 RETURNING *`

	err := s.DB.QueryRowxContext(ctx, query,
		input.Name, input.TCG, input.Category, input.SetName, input.SetCode, input.Condition,
		input.FoilTreatment, input.CardTreatment,
		input.PriceReference, input.PriceSource, input.PriceCOPOverride,
		input.Stock, input.CostBasis, input.ImageURL, input.Description, input.CollectorNumber, input.PromoType,
		input.Language, input.ColorIdentity, input.Rarity, input.CMC, input.IsLegendary, input.IsHistoric, input.IsLand, input.IsBasicLand, input.ArtVariation,
		input.OracleText, input.Artist, input.TypeLine, input.BorderColor, input.Frame, input.FullArt, input.Textless, input.ScryfallID, input.FrameEffects,
		id,
	).StructScan(&product)

	logger.DebugCtx(ctx, "[DB] UpdateProduct result: %+v | Error: %v", product.ID, err)
	if err == nil {
		s.InvalidateCaches()
	}
	return &product, err
}

func (s *ProductStore) GetEnrichedByID(ctx context.Context, id string) (*models.Product, error) {
	if id == "" {
		return nil, fmt.Errorf("GetEnrichedByID: id must not be empty")
	}
	var jsonResult []byte
	query := "SELECT fn_get_product_detail($1)"
	logger.TraceCtx(ctx, "[DB] Executing GetEnrichedByID for %s: %s", id, query)
	err := s.DB.GetContext(ctx, &jsonResult, query, id)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] GetEnrichedByID failed for %s: %v", id, err)
		return nil, err
	}

	var product models.Product
	if err := json.Unmarshal(jsonResult, &product); err != nil {
		return nil, err
	}
	return &product, nil
}

func (s *ProductStore) BulkUpsert(ctx context.Context, jsonData string) ([]string, error) {
	var ids []struct {
		ID string `db:"upserted_id"`
	}
	// s.facetCache.Clear() - cache expires naturally via TTL
	query := "SELECT upserted_id FROM fn_bulk_upsert_product($1)"
	logger.TraceCtx(ctx, "[DB] Executing BulkUpsert: %s | DataLen: %d", query, len(jsonData))
	err := s.DB.SelectContext(ctx, &ids, query, jsonData)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] BulkUpsert failed: %v", err)
		return nil, err
	}

	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = id.ID
	}
	if err == nil {
		s.InvalidateCaches()
	}
	return result, nil
}

func (s *ProductStore) BuildFilters(ctx context.Context, params ProductFilterParams, baseFrom ...string) (string, string, []interface{}) {
	table := "product"
	if len(baseFrom) > 0 {
		table = baseFrom[0]
	}

	fromClause := fmt.Sprintf("FROM %s p", table)
	fromClause += " LEFT JOIN tcg t ON p.tcg = t.id"
	args := []interface{}{}

	var mandatoryConditions []string
	var facetConditions []string

	mandatoryConditions = append(mandatoryConditions, "(t.is_active IS NULL OR t.is_active = true)")

	if params.StorageID != "" {
		fromClause = fmt.Sprintf("FROM %s p JOIN product_storage ps ON p.id = ps.product_id", table)
		fromClause += " LEFT JOIN tcg t ON p.tcg = t.id"
		placeholder := fmt.Sprintf("$%d", len(args)+1)
		args = append(args, params.StorageID)
		mandatoryConditions = append(mandatoryConditions, "ps.storage_id = "+placeholder)
		mandatoryConditions = append(mandatoryConditions, "ps.quantity > 0")
	}

	if params.TCG != "" {
		placeholder := fmt.Sprintf("$%d", len(args)+1)
		args = append(args, strings.ToLower(params.TCG))
		mandatoryConditions = append(mandatoryConditions, "p.tcg = "+placeholder)
	}
	if params.Category != "" {
		placeholder := fmt.Sprintf("$%d", len(args)+1)
		args = append(args, strings.ToLower(params.Category))
		mandatoryConditions = append(mandatoryConditions, "p.category = "+placeholder)
	}

	if params.InStock {
		mandatoryConditions = append(mandatoryConditions, "p.stock > 0")
	}
	if params.OnlyDuplicates {
		mandatoryConditions = append(mandatoryConditions, "p.name IN (SELECT name FROM product GROUP BY name HAVING COUNT(*) > 1)")
	}

	if len(params.IDs) > 0 {
		var idPlaceholders []string
		for _, id := range params.IDs {
			idPlaceholders = append(idPlaceholders, fmt.Sprintf("$%d", len(args)+1))
			args = append(args, id)
		}
		mandatoryConditions = append(mandatoryConditions, fmt.Sprintf("p.id IN (%s)", strings.Join(idPlaceholders, ", ")))
	}

	if params.Search != "" {
		placeholderIdx := len(args) + 1
		args = append(args, params.Search)
		cond := fmt.Sprintf("(p.search_vector @@ websearch_to_tsquery('english', $%d) OR p.name ILIKE '%%' || $%d || '%%')", placeholderIdx, placeholderIdx)
		mandatoryConditions = append(mandatoryConditions, cond)
	}

	isAndMode := strings.ToLower(params.FilterLogic) == "and"

	// Facet Filters
	if params.Foil != "" {
		vals := strings.Split(params.Foil, ",")
		var conds []string
		for _, v := range vals {
			v = strings.TrimSpace(v)
			if v == "" { continue }
			placeholder := fmt.Sprintf("$%d", len(args)+1)
			conds = append(conds, "LOWER(p.foil_treatment) = "+placeholder)
			args = append(args, strings.ToLower(v))
		}
		facetConditions = append(facetConditions, "("+strings.Join(conds, " OR ")+")")
	}
	if params.Treatment != "" {
		vals := strings.Split(params.Treatment, ",")
		var conds []string
		for _, v := range vals {
			v = strings.TrimSpace(v)
			if v == "" { continue }
			lv := strings.ToLower(v)
			placeholder := fmt.Sprintf("$%d", len(args)+1)
			cond := "LOWER(p.card_treatment) = " + placeholder
			switch lv {
			case "full_art":
				cond = "(" + cond + " OR p.full_art = true)"
			case "textless":
				cond = "(" + cond + " OR p.textless = true)"
			}
			conds = append(conds, cond)
			args = append(args, lv)
		}
		facetConditions = append(facetConditions, "("+strings.Join(conds, " OR ")+")")
	}
	if params.Condition != "" {
		vals := strings.Split(params.Condition, ",")
		var conds []string
		for _, v := range vals {
			v = strings.TrimSpace(v)
			if v == "" { continue }
			placeholder := fmt.Sprintf("$%d", len(args)+1)
			conds = append(conds, "p.condition = "+placeholder)
			args = append(args, strings.ToUpper(v))
		}
		facetConditions = append(facetConditions, "("+strings.Join(conds, " OR ")+")")
	}
	if params.Rarity != "" {
		vals := strings.Split(params.Rarity, ",")
		var conds []string
		for _, v := range vals {
			v = strings.TrimSpace(v)
			if v == "" { continue }
			placeholder := fmt.Sprintf("$%d", len(args)+1)
			conds = append(conds, "LOWER(p.rarity) = "+placeholder)
			args = append(args, strings.ToLower(v))
		}
		facetConditions = append(facetConditions, "("+strings.Join(conds, " OR ")+")")
	}
	if params.Language != "" {
		vals := strings.Split(params.Language, ",")
		var conds []string
		for _, v := range vals {
			v = strings.TrimSpace(v)
			if v == "" { continue }
			placeholder := fmt.Sprintf("$%d", len(args)+1)
			conds = append(conds, "LOWER(p.language) = "+placeholder)
			args = append(args, strings.ToLower(v))
		}
		facetConditions = append(facetConditions, "("+strings.Join(conds, " OR ")+")")
	}
	if params.SetName != "" {
		vals := strings.Split(params.SetName, ",")
		var conds []string
		for _, v := range vals {
			v = strings.TrimSpace(v)
			if v == "" { continue }
			placeholder := fmt.Sprintf("$%d", len(args)+1)
			conds = append(conds, "p.set_name = "+placeholder)
			args = append(args, v)
		}
		facetConditions = append(facetConditions, "("+strings.Join(conds, " OR ")+")")
	}

	// MTG Metadata Filters (Properties section)
	var propConds []string
	if params.IsLegendary == "true" { propConds = append(propConds, "p.is_legendary = true") }
	if params.IsLegendary == "false" { propConds = append(propConds, "p.is_legendary = false") }
	if params.IsLand == "true" { propConds = append(propConds, "p.is_land = true") }
	if params.IsLand == "false" { propConds = append(propConds, "p.is_land = false") }
	if params.IsHistoric == "true" { propConds = append(propConds, "p.is_historic = true") }
	if params.IsHistoric == "false" { propConds = append(propConds, "p.is_historic = false") }
	if params.FullArt == "true" { propConds = append(propConds, "p.full_art = true") }
	if params.FullArt == "false" { propConds = append(propConds, "p.full_art = false") }
	if params.Textless == "true" { propConds = append(propConds, "p.textless = true") }
	if params.Textless == "false" { propConds = append(propConds, "p.textless = false") }
	if params.IsPrepared == "true" { propConds = append(propConds, "p.is_prepared = true") }
	if params.LandType == "basic" { propConds = append(propConds, "p.is_basic_land = true") }
	if params.LandType == "non-basic" { propConds = append(propConds, "(p.is_land = true AND p.is_basic_land = false)") }

	if len(propConds) > 0 {
		joinOp := " OR "
		if isAndMode {
			joinOp = " AND "
		}
		facetConditions = append(facetConditions, "("+strings.Join(propConds, joinOp)+")")
	}

	// Multi-value fields
	if params.Format != "" {
		vals := strings.Split(params.Format, ",")
		var conds []string
		for _, v := range vals {
			v = strings.TrimSpace(v)
			if v == "" { continue }
			placeholder := fmt.Sprintf("$%d", len(args)+1)
			args = append(args, v)
			conds = append(conds, fmt.Sprintf("p.legalities->>%s = 'legal'", placeholder))
		}
		joinOp := " OR "
		if isAndMode {
			joinOp = " AND "
		}
		facetConditions = append(facetConditions, "("+strings.Join(conds, joinOp)+")")
	}
	if params.Collection != "" {
		vals := strings.Split(params.Collection, ",")
		if isAndMode {
			for _, v := range vals {
				v = strings.TrimSpace(v)
				if v == "" { continue }
				placeholder := fmt.Sprintf("$%d", len(args)+1)
				cond := fmt.Sprintf("p.id IN (SELECT pc_col.product_id FROM product_category pc_col JOIN custom_category c_col ON pc_col.category_id = c_col.id WHERE c_col.slug = %s)", placeholder)
				facetConditions = append(facetConditions, cond)
				args = append(args, strings.ToLower(v))
			}
		} else {
			var conds []string
			for _, v := range vals {
				v = strings.TrimSpace(v)
				if v == "" { continue }
				placeholder := fmt.Sprintf("$%d", len(args)+1)
				cond := fmt.Sprintf("p.id IN (SELECT pc_col.product_id FROM product_category pc_col JOIN custom_category c_col ON pc_col.category_id = c_col.id WHERE c_col.slug = %s)", placeholder)
				conds = append(conds, cond)
				args = append(args, strings.ToLower(v))
			}
			facetConditions = append(facetConditions, "("+strings.Join(conds, " OR ")+")")
		}
	}
	if params.Color != "" {
		vals := strings.Split(params.Color, ",")
		var conds []string
		for _, v := range vals {
			v = strings.TrimSpace(v)
			if v == "" { continue }
			placeholder := fmt.Sprintf("$%d", len(args)+1)
			conds = append(conds, "p.color_identity ILIKE "+placeholder)
			args = append(args, "%"+strings.ToUpper(v)+"%")
		}
		joinOp := " OR "
		if isAndMode {
			joinOp = " AND "
		}
		facetConditions = append(facetConditions, "("+strings.Join(conds, joinOp)+")")
	}

	if params.FrameEffects != "" {
		vals := strings.Split(params.FrameEffects, ",")
		var conds []string
		for _, v := range vals {
			v = strings.TrimSpace(v)
			if v == "" { continue }
			placeholder := fmt.Sprintf("$%d", len(args)+1)
			conds = append(conds, "p.frame_effects ? "+placeholder)
			args = append(args, strings.ToLower(v))
		}
		joinOp := " OR "
		if isAndMode {
			joinOp = " AND "
		}
		facetConditions = append(facetConditions, "("+strings.Join(conds, joinOp)+")")
	}

	if params.CardTypes != "" {
		vals := strings.Split(params.CardTypes, ",")
		var conds []string
		for _, v := range vals {
			v = strings.TrimSpace(v)
			if v == "" { continue }
			placeholder := fmt.Sprintf("$%d", len(args)+1)
			conds = append(conds, "p.card_types ? "+placeholder)
			args = append(args, v)
		}
		joinOp := " OR "
		if isAndMode {
			joinOp = " AND "
		}
		facetConditions = append(facetConditions, "("+strings.Join(conds, joinOp)+")")
	}

	finalConditions := mandatoryConditions
	if len(facetConditions) > 0 {
		joinOp := " OR "
		if isAndMode {
			joinOp = " AND "
		}
		finalConditions = append(finalConditions, "("+strings.Join(facetConditions, joinOp)+")")
	}

	whereClause := ""
	if len(finalConditions) > 0 {
		whereClause = "WHERE " + strings.Join(finalConditions, " AND ")
	}

	logger.DebugCtx(ctx, "[DB] BuildFilters: logic=%s, isAnd=%v, where=%s", params.FilterLogic, isAndMode, whereClause)

	return fromClause, whereClause, args
}

func (s *ProductStore) BuildOrderBy(params ProductFilterParams, argsLen int) string {
	dir := "DESC"
	if strings.EqualFold(params.SortDir, "asc") {
		dir = "ASC"
	}
	if params.SortBy == "" {
		if params.Search != "" {
			placeholderIdx := argsLen
			return fmt.Sprintf("ts_rank(p.search_vector, websearch_to_tsquery('english', $%d)) DESC, p.id DESC", placeholderIdx)
		}
		return "p.id DESC"
	}

	var col string
	switch strings.ToLower(params.SortBy) {
	case "name":
		col = "p.name"
	case "tcg":
		col = "p.tcg"
	case "category":
		col = "p.category"
	case "set_name":
		col = "COALESCE(p.set_name, '')"
	case "condition":
		col = "COALESCE(p.condition, '')"
	case "stock":
		col = "p.stock"
	case "price":
		col = fmt.Sprintf(`COALESCE(p.price_cop_override,
			CASE p.price_source
				WHEN 'tcgplayer' THEN p.price_reference * %f
				WHEN 'cardmarket' THEN p.price_reference * %f
				WHEN 'cardkingdom' THEN p.price_reference * %f
				ELSE p.price_reference
			END)`, params.USDRate, params.EURRate, params.CKRate)
	case "cmc":
		col = "COALESCE(p.cmc, 0)"
	case "rarity":
		col = `CASE LOWER(COALESCE(p.rarity, ''))
			WHEN 'mythic' THEN 6
			WHEN 'special' THEN 5
			WHEN 'rare' THEN 4
			WHEN 'uncommon' THEN 3
			WHEN 'common' THEN 2
			WHEN 'bonus' THEN 1
			ELSE 0 END`
	case "id":
		col = "p.id"
	case "created_at":
		col = "p.created_at"
	case "updated_at":
		col = "p.updated_at"
	default:
		if params.Search != "" {
			return fmt.Sprintf("ts_rank(p.search_vector, websearch_to_tsquery('english', $%d)) DESC, p.id DESC", argsLen)
		}
		return "p.id DESC"
	}
	return col + " " + dir + ", p.id DESC"
}

func (s *ProductStore) GetRecommendations(ctx context.Context, cartIDs []string, settings models.Settings) ([]models.Product, error) {
	if len(cartIDs) == 0 {
		return []models.Product{}, nil
	}

	// We use a named query for better readability and to use "variables" instead of positional parameters.
	query := `
		WITH cart_metadata AS (
			SELECT 
				array_agg(DISTINCT c) FILTER (WHERE c IS NOT NULL) as colors,
				array_agg(DISTINCT set_code) FILTER (WHERE set_code IS NOT NULL) as sets
			FROM product, UNNEST(STRING_TO_ARRAY(color_identity, ',')) as c
			WHERE id = ANY(:cart_ids)
		)
		SELECT p.*
		FROM product p, cart_metadata cm
		WHERE p.id != ALL(:cart_ids)
		  AND p.stock >= 1
		  AND (
		  	-- Check price using source-specific rate from variables
		  	COALESCE(p.price_cop_override, 
		  	  CASE p.price_source
		  	    WHEN 'cardkingdom' THEN p.price_reference * :ck_rate
		  	    WHEN 'tcgplayer'   THEN p.price_reference * :usd_rate
		  	    WHEN 'cardmarket'   THEN p.price_reference * :eur_rate
		  	    ELSE p.price_reference
		  	  END
		  	) <= :max_price
		  )
		  AND (
		      -- Match color identity overlap (at least one shared color)
		      (p.color_identity IS NOT NULL AND cm.colors IS NOT NULL AND STRING_TO_ARRAY(p.color_identity, ',') && cm.colors)
		      OR
		      -- Match set
		      (p.set_code IS NOT NULL AND cm.sets IS NOT NULL AND p.set_code = ANY(cm.sets))
		      OR
		      -- If cart has colorless/no-set cards, fallback to top-stock singles
		      (cm.colors IS NULL AND cm.sets IS NULL)
		  )
		ORDER BY p.stock DESC, random()
		LIMIT 10
	`

	argMap := map[string]interface{}{
		"cart_ids":  cartIDs,
		"usd_rate":  settings.USDToCOPRate,
		"eur_rate":  settings.EURToCOPRate,
		"ck_rate":   settings.CKToCOPRate,
		"max_price": settings.SynergyMaxPriceCOP,
	}

	logger.TraceCtx(ctx, "[DB] GetRecommendations params: ids=%v, usd=%.2f, eur=%.2f, ck=%.2f, max=%.2f",
		cartIDs, settings.USDToCOPRate, settings.EURToCOPRate, settings.CKToCOPRate, settings.SynergyMaxPriceCOP)

	nQuery, nArgs, err := sqlx.Named(query, argMap)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] GetRecommendations Named bind failed: %v", err)
		return nil, fmt.Errorf("failed to bind named query: %w", err)
	}
	nQuery = s.DB.Rebind(nQuery)

	var products []models.Product
	if err := s.DB.Unsafe().SelectContext(ctx, &products, nQuery, nArgs...); err != nil {
		logger.ErrorCtx(ctx, "[DB] GetRecommendations query failed: %v", err)
		return nil, err
	}

	if len(products) > 0 {
		s.PopulateStorage(ctx, products)
		s.PopulateCategories(ctx, products)
		s.PopulateCartCounts(ctx, products)
		s.PopulateDeckCards(ctx, products)
	}

	return products, nil
}

func (s *ProductStore) GetByNames(ctx context.Context, names []string) ([]models.Product, error) {
	if len(names) == 0 {
		return []models.Product{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT p.* 
		FROM product p
		WHERE p.name IN (?) AND p.stock >= 1
		LIMIT 20
	`, names)
	if err != nil {
		return nil, err
	}
	query = s.DB.Rebind(query)

	var products []models.Product
	if err := s.DB.Unsafe().SelectContext(ctx, &products, query, args...); err != nil {
		return nil, err
	}

	if len(products) > 0 {
		s.PopulateStorage(ctx, products)
		s.PopulateCategories(ctx, products)
		s.PopulateCartCounts(ctx, products)
		s.PopulateDeckCards(ctx, products)
	}

	return products, nil
}

func (s *ProductStore) GetByIDs(ctx context.Context, ids []string) ([]models.Product, error) {
	if len(ids) == 0 {
		return []models.Product{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT p.* 
		FROM product p
		WHERE p.id IN (?)
	`, ids)
	if err != nil {
		return nil, err
	}
	query = s.DB.Rebind(query)

	var products []models.Product
	if err := s.DB.Unsafe().SelectContext(ctx, &products, query, args...); err != nil {
		return nil, err
	}

	if len(products) > 0 {
		s.PopulateStorage(ctx, products)
		s.PopulateCategories(ctx, products)
		s.PopulateCartCounts(ctx, products)
		s.PopulateDeckCards(ctx, products)
	}

	return products, nil
}

func (s *ProductStore) BulkMoveStorage(ctx context.Context, req models.BulkMoveStorageRequest) error {
	return s.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		for _, move := range req.Moves {
			if move.Quantity <= 0 {
				continue
			}

			// 1. Decrement from source
			_, err := tx.ExecContext(ctx, `
				UPDATE product_storage 
				SET quantity = quantity - $1 
				WHERE product_id = $2 AND storage_id = $3
			`, move.Quantity, move.ProductID, move.FromStorageID)
			if err != nil {
				return fmt.Errorf("failed to decrement source: %w", err)
			}

			// 2. Cleanup source if 0 (ensures clean DB)
			_, err = tx.ExecContext(ctx, `
				DELETE FROM product_storage 
				WHERE product_id = $1 AND storage_id = $2 AND quantity = 0
			`, move.ProductID, move.FromStorageID)
			if err != nil {
				return fmt.Errorf("failed to cleanup source: %w", err)
			}

			// 3. Increment target (Upsert)
			_, err = tx.ExecContext(ctx, `
				INSERT INTO product_storage (product_id, storage_id, quantity)
				VALUES ($1, $2, $3)
				ON CONFLICT (product_id, storage_id) 
				DO UPDATE SET quantity = product_storage.quantity + EXCLUDED.quantity
			`, move.ProductID, req.TargetStorageID, move.Quantity)
			if err != nil {
				return fmt.Errorf("failed to increment target: %w", err)
			}
		}
		return nil
	})
}
