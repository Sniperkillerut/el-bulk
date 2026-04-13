package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/cache"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

type ProductStore struct {
	*BaseStore[models.Product]
	facetCache *cache.TTLMap[models.Facets]
}

func NewProductStore(db *sqlx.DB) *ProductStore {
	return &ProductStore{
		BaseStore:  NewBaseStore[models.Product](db, "product"),
		facetCache: cache.NewTTLMap[models.Facets](5 * time.Minute),
	}
}

type ProductFilterParams struct {
	TCG         string
	Category    string
	Search      string
	StorageID   string
	Foil        string
	Treatment   string
	Condition   string
	Collection  string
	Rarity      string
	Language    string
	Color       string
	SetName     string
	InStock     bool
	SortBy      string
	SortDir     string
	FilterLogic string
	Page        int
	PageSize    int
	Offset      int
}

func (s *ProductStore) ListWithFilters(ctx context.Context, params ProductFilterParams) ([]models.Product, int, error) {
	start := time.Now()
	fromClause, where, args := s.buildFilters(params)

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) %s %s", fromClause, where)
	logger.TraceCtx(ctx, "[DB] Executing countQuery in ListWithFilters: %s | Args: %+v", countQuery, args)
	if err := s.DB.GetContext(ctx, &total, countQuery, args...); err != nil {
		logger.ErrorCtx(ctx, "[DB] ListWithFilters countQuery failed: %v", err)
		return nil, 0, err
	}

	orderBy := s.buildOrderBy(params.SortBy, params.SortDir, params.Search, len(args))

	// Use enriched view for listing
	viewFrom, where, args := s.buildFilters(params, "view_product_enriched")

	orderBy = s.buildOrderBy(params.SortBy, params.SortDir, params.Search, len(args))

	listQuery := fmt.Sprintf("SELECT p.* %s %s ORDER BY %s LIMIT $%d OFFSET $%d",
		viewFrom, where, orderBy, len(args)+1, len(args)+2)

	listArgs := append([]interface{}{}, args...)
	listArgs = append(listArgs, params.PageSize, params.Offset)

	var rows []struct {
		models.Product
		StoredInJSON   []byte `db:"stored_in_json"`
		CategoriesJSON []byte `db:"categories_json"`
		DeckCardsJSON  []byte `db:"deck_cards_json"`
	}

	if err := s.DB.Unsafe().SelectContext(ctx, &rows, listQuery, listArgs...); err != nil {
		logger.ErrorCtx(ctx, "[DB] ListWithFilters listQuery failed: %v", err)
		return nil, 0, err
	}
	logger.DebugCtx(ctx, "[DB] ListWithFilters (count+list) took %v", time.Since(start))

	products := make([]models.Product, len(rows))
	for i, r := range rows {
		products[i] = r.Product
		if r.StoredInJSON != nil {
			json.Unmarshal(r.StoredInJSON, &products[i].StoredIn)
		}
		if r.CategoriesJSON != nil {
			json.Unmarshal(r.CategoriesJSON, &products[i].Categories)
		}
		if r.DeckCardsJSON != nil {
			json.Unmarshal(r.DeckCardsJSON, &products[i].DeckCards)
		}
	}

	return products, total, nil
}

func (s *ProductStore) GetFacets(ctx context.Context, params ProductFilterParams, isAdmin bool) (models.Facets, error) {
	// Generate cache key from params
	cacheKey := fmt.Sprintf("facets:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v",
		params.TCG, params.Category, params.Search, params.StorageID, params.Foil, params.Treatment, params.Condition,
		params.Rarity, params.Language, params.Color, params.Collection, params.SetName, params.InStock, params.FilterLogic, isAdmin)

	if cached, ok := s.facetCache.Get(cacheKey); ok {
		logger.TraceCtx(ctx, "[CACHE] Facet hit for key: %s", cacheKey)
		return cached, nil
	}

	var result []byte
	start := time.Now()
	query := "SELECT fn_get_product_facets($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)"
	logger.TraceCtx(ctx, "[DB] Executing GetFacets: %s", query)
	err := s.DB.GetContext(ctx, &result, query,
		params.TCG, params.Category, params.Search, params.StorageID, params.Foil, params.Treatment, params.Condition,
		params.Rarity, params.Language, params.Color, params.Collection, params.SetName, params.InStock, params.FilterLogic, isAdmin)

	if err != nil {
		logger.ErrorCtx(ctx, "[DB] GetFacets failed: %v", err)
		return models.Facets{}, err
	}
	logger.DebugCtx(ctx, "[DB] GetFacets took %v", time.Since(start))

	var facets models.Facets
	if err := json.Unmarshal(result, &facets); err != nil {
		return models.Facets{}, err
	}

	// Cache the result for 2 minutes
	s.facetCache.Set(cacheKey, facets, 2*time.Minute)

	return facets, nil
}

func (s *ProductStore) PopulateStorage(ctx context.Context, products []models.Product) error {
	if len(products) == 0 {
		return nil
	}
	var pids []string
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

	for i := range products {
		if locs, ok := storageMap[products[i].ID]; ok {
			products[i].StoredIn = locs
		} else {
			products[i].StoredIn = []models.StorageLocation{}
		}
	}
	return nil
}

func (s *ProductStore) PopulateCategories(ctx context.Context, products []models.Product) error {
	if len(products) == 0 {
		return nil
	}
	var pids []string
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

	for i := range products {
		if cats, ok := catMap[products[i].ID]; ok {
			products[i].Categories = cats
		} else {
			products[i].Categories = []models.CustomCategory{}
		}
	}
	return nil
}

func (s *ProductStore) PopulateCartCounts(ctx context.Context, products []models.Product) error {
	if len(products) == 0 {
		return nil
	}
	var pids []string
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
	validIDs := []string{}
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
	// Explicitly rebind using DOLLAR for Postgres compatibility, 
	// as some custom driver names (like cloudsql-postgres) might not be auto-detected by sqlx.
	err = s.DB.SelectContext(ctx, &hotIDs, sqlx.Rebind(sqlx.DOLLAR, query), args...)
	return hotIDs, err
}

func (s *ProductStore) SaveCategories(ctx context.Context, productID string, categoryIDs []string) error {
	s.facetCache.Clear()
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
	return nil
}

func (s *ProductStore) SaveDeckCards(ctx context.Context, productID string, cards []models.DeckCard) error {
	s.facetCache.Clear()
	_, err := s.DB.ExecContext(ctx, "DELETE FROM deck_card WHERE product_id = $1", productID)
	if err != nil {
		return err
	}

	if len(cards) == 0 {
		return nil
	}

	query := "INSERT INTO deck_card (product_id, name, set_code, collector_number, quantity, type_line, image_url, foil_treatment, card_treatment, rarity, art_variation, scryfall_id) VALUES "
	values := make([]interface{}, 0, len(cards)*12)
	placeholders := make([]string, 0, len(cards))

	for i, c := range cards {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*12+1, i*12+2, i*12+3, i*12+4, i*12+5, i*12+6, i*12+7, i*12+8, i*12+9, i*12+10, i*12+11, i*12+12))
		values = append(values, productID, c.Name, c.SetCode, c.CollectorNumber, c.Quantity, c.TypeLine, c.ImageURL, c.FoilTreatment, c.CardTreatment, c.Rarity, c.ArtVariation, c.ScryfallID)
	}

	query += strings.Join(placeholders, ", ")
	logger.TraceCtx(ctx, "[DB] Executing SaveDeckCards for product %s: %s | Values: %d", productID, query, len(values))
	_, err = s.DB.ExecContext(ctx, query, values...)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] SaveDeckCards failed for product %s: %v", productID, err)
	}
	return err
}

func (s *ProductStore) SaveStorage(ctx context.Context, productID string, items []models.StorageLocation) error {
	s.facetCache.Clear()
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
	s.facetCache.Clear()
	query := `INSERT INTO product (
		name, tcg, category, set_name, set_code, condition,
		foil_treatment, card_treatment,
		price_reference, price_source, price_cop_override,
		stock, cost_basis_cop, image_url, description, collector_number, promo_type,
		language, color_identity, rarity, cmc, is_legendary, is_historic, is_land, is_basic_land, art_variation,
		oracle_text, artist, type_line, border_color, frame, full_art, textless, scryfall_id
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34
	) RETURNING *`

	err := s.DB.QueryRowxContext(ctx, query,
		input.Name, input.TCG, input.Category, input.SetName, input.SetCode, input.Condition,
		input.FoilTreatment, input.CardTreatment,
		input.PriceReference, input.PriceSource, input.PriceCOPOverride,
		input.Stock, input.CostBasis, input.ImageURL, input.Description, input.CollectorNumber, input.PromoType,
		input.Language, input.ColorIdentity, input.Rarity, input.CMC, input.IsLegendary, input.IsHistoric, input.IsLand, input.IsBasicLand, input.ArtVariation,
		input.OracleText, input.Artist, input.TypeLine, input.BorderColor, input.Frame, input.FullArt, input.Textless, input.ScryfallID,
	).StructScan(&product)

	logger.DebugCtx(ctx, "[DB] CreateProduct result: %+v | Error: %v", product.ID, err)

	return &product, err
}

func (s *ProductStore) UpdateProduct(ctx context.Context, id string, input models.ProductInput) (*models.Product, error) {
	if input.PriceSource == "" {
		input.PriceSource = models.PriceSourceManual
	}

	var product models.Product
	s.facetCache.Clear()
	query := `UPDATE product SET
		name=$1, tcg=$2, category=$3, set_name=$4, set_code=$5, condition=$6,
		foil_treatment=$7, card_treatment=$8,
		price_reference=$9, price_source=$10, price_cop_override=$11,
		stock=$12, cost_basis_cop=$13, image_url=$14, description=$15, collector_number=$16, promo_type=$17,
		language=$18, color_identity=$19, rarity=$20, cmc=$21, is_legendary=$22, is_historic=$23, is_land=$24, is_basic_land=$25, art_variation=$26,
		oracle_text=$27, artist=$28, type_line=$29, border_color=$30, frame=$31, full_art=$32, textless=$33, scryfall_id=$34
	WHERE id=$35 RETURNING *`

	err := s.DB.QueryRowxContext(ctx, query,
		input.Name, input.TCG, input.Category, input.SetName, input.SetCode, input.Condition,
		input.FoilTreatment, input.CardTreatment,
		input.PriceReference, input.PriceSource, input.PriceCOPOverride,
		input.Stock, input.CostBasis, input.ImageURL, input.Description, input.CollectorNumber, input.PromoType,
		input.Language, input.ColorIdentity, input.Rarity, input.CMC, input.IsLegendary, input.IsHistoric, input.IsLand, input.IsBasicLand, input.ArtVariation,
		input.OracleText, input.Artist, input.TypeLine, input.BorderColor, input.Frame, input.FullArt, input.Textless, input.ScryfallID,
		id,
	).StructScan(&product)

	logger.DebugCtx(ctx, "[DB] UpdateProduct result: %+v | Error: %v", product.ID, err)
	return &product, err
}

func (s *ProductStore) GetEnrichedByID(ctx context.Context, id string) (*models.Product, error) {
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
	s.facetCache.Clear()
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
	return result, nil
}

func (s *ProductStore) buildFilters(params ProductFilterParams, baseFrom ...string) (string, string, []interface{}) {
	table := "product"
	if len(baseFrom) > 0 {
		table = baseFrom[0]
	}

	fromClause := fmt.Sprintf("FROM %s p", table)
	args := []interface{}{}

	var mandatory []string
	var optional []string

	fromClause += " LEFT JOIN tcg t ON p.tcg = t.id"
	mandatory = append(mandatory, "(t.is_active IS NULL OR t.is_active = true)")

	if params.StorageID != "" {
		fromClause = fmt.Sprintf("FROM %s p JOIN product_storage ps ON p.id = ps.product_id", table)
		fromClause += " LEFT JOIN tcg t ON p.tcg = t.id"
		placeholder := fmt.Sprintf("$%d", len(args)+1)
		args = append(args, params.StorageID)
		mandatory = append(mandatory, "ps.storage_id = "+placeholder)
		mandatory = append(mandatory, "ps.quantity > 0")
	}

	if params.TCG != "" {
		placeholder := fmt.Sprintf("$%d", len(args)+1)
		args = append(args, strings.ToLower(params.TCG))
		mandatory = append(mandatory, "p.tcg = "+placeholder)
	}
	if params.Category != "" {
		placeholder := fmt.Sprintf("$%d", len(args)+1)
		args = append(args, strings.ToLower(params.Category))
		mandatory = append(mandatory, "p.category = "+placeholder)
	}

	if params.InStock {
		mandatory = append(mandatory, "p.stock > 0")
	}
	if params.Search != "" {
		placeholderIdx := len(args) + 1
		args = append(args, params.Search)
		mandatory = append(mandatory, fmt.Sprintf("p.search_vector @@ websearch_to_tsquery('english', $%d)", placeholderIdx))
	}

	opLogic := " OR "
	if strings.ToLower(params.FilterLogic) == "and" {
		opLogic = " AND "
	}

	if params.Foil != "" {
		vals := strings.Split(params.Foil, ",")
		var conds []string
		for _, v := range vals {
			placeholder := fmt.Sprintf("$%d", len(args)+1)
			conds = append(conds, "LOWER(p.foil_treatment) = "+placeholder)
			args = append(args, strings.ToLower(v))
		}
		optional = append(optional, "("+strings.Join(conds, opLogic)+")")
	}
	if params.Treatment != "" {
		vals := strings.Split(params.Treatment, ",")
		var conds []string
		for _, v := range vals {
			lv := strings.ToLower(v)
			placeholder := fmt.Sprintf("$%d", len(args)+1)
			cond := "LOWER(p.card_treatment) = " + placeholder
			switch lv {
			case "textless":
				cond = "(" + cond + " OR p.textless = true)"
			case "full_art":
				cond = "(" + cond + " OR p.full_art = true)"
			}
			conds = append(conds, cond)
			args = append(args, lv)
		}
		optional = append(optional, "("+strings.Join(conds, opLogic)+")")
	}
	if params.Condition != "" {
		vals := strings.Split(params.Condition, ",")
		var conds []string
		for _, v := range vals {
			placeholder := fmt.Sprintf("$%d", len(args)+1)
			conds = append(conds, "p.condition = "+placeholder)
			args = append(args, strings.ToUpper(v))
		}
		optional = append(optional, "("+strings.Join(conds, opLogic)+")")
	}
	if params.Collection != "" {
		vals := strings.Split(params.Collection, ",")
		var conds []string
		for _, v := range vals {
			placeholder := fmt.Sprintf("$%d", len(args)+1)
			// Simpler subquery to avoid complex nesting collisions at character 194
			cond := fmt.Sprintf("p.id IN (SELECT pc_col.product_id FROM product_category pc_col JOIN custom_category c_col ON pc_col.category_id = c_col.id WHERE c_col.slug = %s)", placeholder)
			conds = append(conds, cond)
			args = append(args, strings.ToLower(v))
		}
		optional = append(optional, "("+strings.Join(conds, opLogic)+")")
	}
	if params.Rarity != "" {
		vals := strings.Split(params.Rarity, ",")
		var conds []string
		for _, v := range vals {
			placeholder := fmt.Sprintf("$%d", len(args)+1)
			conds = append(conds, "LOWER(p.rarity) = "+placeholder)
			args = append(args, strings.ToLower(v))
		}
		optional = append(optional, "("+strings.Join(conds, opLogic)+")")
	}
	if params.Language != "" {
		vals := strings.Split(params.Language, ",")
		var conds []string
		for _, v := range vals {
			placeholder := fmt.Sprintf("$%d", len(args)+1)
			conds = append(conds, "LOWER(p.language) = "+placeholder)
			args = append(args, strings.ToLower(v))
		}
		optional = append(optional, "("+strings.Join(conds, opLogic)+")")
	}
	if params.Color != "" {
		vals := strings.Split(params.Color, ",")
		var conds []string
		for _, v := range vals {
			placeholder := fmt.Sprintf("$%d", len(args)+1)
			conds = append(conds, "p.color_identity ILIKE "+placeholder)
			args = append(args, "%"+strings.ToUpper(v)+"%")
		}
		optional = append(optional, "("+strings.Join(conds, opLogic)+")")
	}
	if params.SetName != "" {
		vals := strings.Split(params.SetName, ",")
		var conds []string
		for _, v := range vals {
			placeholder := fmt.Sprintf("$%d", len(args)+1)
			conds = append(conds, "p.set_name = "+placeholder)
			args = append(args, v)
		}
		optional = append(optional, "("+strings.Join(conds, opLogic)+")")
	}

	finalConditions := mandatory
	if len(optional) > 0 {
		finalConditions = append(finalConditions, "("+strings.Join(optional, opLogic)+")")
	}

	whereClause := ""
	if len(finalConditions) > 0 {
		whereClause = "WHERE " + strings.Join(finalConditions, " AND ")
	}

	return fromClause, whereClause, args
}

func (s *ProductStore) buildOrderBy(sortBy, sortDir, search string, argsLen int) string {
	dir := "DESC"
	if strings.EqualFold(sortDir, "asc") {
		dir = "ASC"
	}
	if sortBy == "" {
		if search != "" {
			placeholderIdx := argsLen // The search term is usually the last arg before limit/offset
			return fmt.Sprintf("ts_rank(p.search_vector, websearch_to_tsquery('english', $%d)) DESC, p.created_at DESC", placeholderIdx)
		}
		return "p.created_at DESC"
	}

	var col string
	switch strings.ToLower(sortBy) {
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
	case "created_at":
		col = "p.created_at"
	case "updated_at":
		col = "p.updated_at"
	default:
		if search != "" {
			return fmt.Sprintf("ts_rank(p.search_vector, websearch_to_tsquery('english', $%d)) DESC, p.created_at DESC", argsLen)
		}
		return "p.created_at DESC"
	}
	return col + " " + dir + ", p.created_at DESC"
}
