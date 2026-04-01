package handlers

import (
	"github.com/el-bulk/backend/utils/render"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/httputil"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/sqlutil"
)

type ProductHandler struct {
	DB *sqlx.DB
}

func NewProductHandler(db *sqlx.DB) *ProductHandler {
	return &ProductHandler{DB: db}
}

func (h *ProductHandler) populatePrices(products []models.Product) {
	s, err := loadSettings(h.DB)
	if err != nil {
		s = models.Settings{USDToCOPRate: 4200, EURToCOPRate: 4600}
	}
	for i := range products {
		products[i].Price = products[i].ComputePrice(s.USDToCOPRate, s.EURToCOPRate)
	}
}

func (h *ProductHandler) populateStorage(products []models.Product) {
	if len(products) == 0 {
		return
	}
	var pids []string
	for _, p := range products {
		pids = append(pids, p.ID)
	}

	query, args, err := sqlx.In(`
		SELECT ps.product_id, s.id as stored_in_id, s.name, ps.quantity 
		FROM product_storage ps 
		JOIN storage_location s ON ps.storage_id = s.id 
		WHERE ps.quantity > 0 AND ps.product_id IN (?)
	`, pids)
	if err != nil {
		return
	}

	query = h.DB.Rebind(query)
	var storageRows []struct {
		ProductID string `db:"product_id"`
		StorageID string `db:"stored_in_id"`
		Name      string `db:"name"`
		Quantity  int    `db:"quantity"`
	}
	if err := h.DB.Select(&storageRows, query, args...); err != nil {
		return
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
}

func (h *ProductHandler) populateCartCounts(products []models.Product) {
	if len(products) == 0 {
		return
	}
	var pids []string
	for _, p := range products {
		pids = append(pids, p.ID)
	}

	query, args, err := sqlx.In(`
		SELECT oi.product_id, COUNT(DISTINCT o.customer_id) as cart_count
		FROM "order" o
		JOIN order_item oi ON o.id = oi.order_id
		WHERE o.status = 'pending' AND oi.product_id IN (?)
		GROUP BY oi.product_id
	`, pids)
	if err != nil {
		return
	}

	query = h.DB.Rebind(query)
	var countRows []struct {
		ProductID string `db:"product_id"`
		CartCount int    `db:"cart_count"`
	}
	if err := h.DB.Select(&countRows, query, args...); err != nil {
		return
	}

	countMap := make(map[string]int)
	for _, r := range countRows {
		countMap[r.ProductID] = r.CartCount
	}

	for i := range products {
		products[i].CartCount = countMap[products[i].ID]
	}
}

func (h *ProductHandler) saveProductCategories(productID string, categoryIDs []string) {
	_, err := h.DB.Exec("DELETE FROM product_category WHERE product_id = $1", productID)
	if err != nil {
		logger.Error("Error deleting product_category: %v", err)
	}
	for _, cid := range categoryIDs {
		_, err := h.DB.Exec("INSERT INTO product_category (product_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", productID, cid)
		if err != nil {
			logger.Error("Error inserting product_category (product=%s, cat=%s): %v", productID, cid, err)
		}
	}
}

func (h *ProductHandler) saveDeckCards(productID string, cards []models.DeckCard) {
	_, err := h.DB.Exec("DELETE FROM deck_card WHERE product_id = $1", productID)
	if err != nil {
		logger.Error("Error clearing deck_card for %s: %v", productID, err)
	}

	if len(cards) == 0 {
		return
	}

	query := "INSERT INTO deck_card (product_id, name, set_code, collector_number, quantity, type_line, image_url, foil_treatment, card_treatment, rarity, art_variation) VALUES "
	values := make([]interface{}, 0, len(cards)*11)
	placeholders := make([]string, 0, len(cards))

	for i, c := range cards {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*11+1, i*11+2, i*11+3, i*11+4, i*11+5, i*11+6, i*11+7, i*11+8, i*11+9, i*11+10, i*11+11))
		values = append(values, productID, c.Name, c.SetCode, c.CollectorNumber, c.Quantity, c.TypeLine, c.ImageURL, c.FoilTreatment, c.CardTreatment, c.Rarity, c.ArtVariation)
	}

	query += strings.Join(placeholders, ", ")
	_, err = h.DB.Exec(query, values...)
	if err != nil {
		logger.Error("Error bulk inserting deck_card for %s: %v", productID, err)
	}
}

func (h *ProductHandler) saveProductStorage(productID string, items []models.StorageLocation) {
	_, err := h.DB.Exec("DELETE FROM product_storage WHERE product_id = $1", productID)
	if err != nil {
		logger.Error("Error clearing product_storage for %s: %v", productID, err)
	}

	var validItems []models.StorageLocation
	for _, item := range items {
		if item.Quantity > 0 {
			validItems = append(validItems, item)
		}
	}

	if len(validItems) == 0 {
		return
	}

	query := "INSERT INTO product_storage (product_id, storage_id, quantity) VALUES "
	values := make([]interface{}, 0, len(validItems)*3)
	placeholders := make([]string, 0, len(validItems))

	for i, item := range validItems {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
		values = append(values, productID, item.StorageID, item.Quantity)
	}

	query += strings.Join(placeholders, ", ")
	_, err = h.DB.Exec(query, values...)
	if err != nil {
		logger.Error("Error bulk inserting product_storage for %s: %v", productID, err)
	}
}

func (h *ProductHandler) populateCategories(products []models.Product, isAdmin bool) {
	if len(products) == 0 {
		return
	}
	var pids []string
	for _, p := range products {
		pids = append(pids, p.ID)
	}

	sql := `
		SELECT pc.product_id, c.id, c.name, c.slug, c.show_badge, c.is_active, c.searchable
		FROM product_category pc
		JOIN custom_category c ON pc.category_id = c.id
		WHERE pc.product_id IN (?)
	`
	if !isAdmin {
		sql += " AND c.show_badge = true "
	}
	sql += " ORDER BY c.name "

	query, args, err := sqlx.In(sql, pids)
	if err != nil {
		logger.Error("Error creating IN query for populateCategories: %v", err)
		return
	}

	query = h.DB.Rebind(query)
	var catRows []struct {
		ProductID  string `db:"product_id"`
		ID         string `db:"id"`
		Name       string `db:"name"`
		Slug       string `db:"slug"`
		ShowBadge  bool   `db:"show_badge"`
		IsActive   bool   `db:"is_active"`
		Searchable bool   `db:"searchable"`
	}
	if err := h.DB.Select(&catRows, query, args...); err != nil {
		logger.Error("Error selecting categories: %v", err)
		return
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
		})
	}

	for i := range products {
		if cats, ok := catMap[products[i].ID]; ok {
			products[i].Categories = cats
		} else {
			products[i].Categories = []models.CustomCategory{}
		}
	}
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	q := r.URL.Query()

	tcg := q.Get("tcg")
	category := q.Get("category")
	search := q.Get("search")
	storageID := q.Get("storage_id")
	foil := q.Get("foil")
	treatment := q.Get("treatment")
	condition := q.Get("condition")
	collection := q.Get("collection")
	rarity := q.Get("rarity")
	language := q.Get("language")
	color := q.Get("color")
	setName := q.Get("set_name")
	inStock := q.Get("in_stock") == "true"
	sortBy := q.Get("sort_by")
	sortDir := q.Get("sort_dir")
	filterLogic := q.Get("logic")
	if filterLogic == "" {
		filterLogic = "or"
	}

	isAdmin := strings.Contains(r.URL.Path, "/admin/")
	maxPageSize := 100
	if isAdmin {
		maxPageSize = 5000
	}
	page, pageSize, offset := httputil.GetPagination(r, 20, maxPageSize)

	fromClause, conditions, args := h.buildFilters(tcg, category, search, storageID, foil, treatment, condition, collection, rarity, language, color, setName, inStock, filterLogic, isAdmin)

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	if err := h.DB.Get(&total, "SELECT COUNT(*) "+fromClause+" "+where, args...); err != nil {
		logger.Error("Error counting products: %v.", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	orderBy := h.buildOrderBy(sortBy, sortDir, search, len(args))

	viewFrom := strings.Replace(fromClause, "FROM product p", "FROM view_product_enriched p", 1)
	listQuery := `SELECT p.* ` + viewFrom + " " + where + " ORDER BY " + orderBy + " LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)

	listArgs := append([]interface{}{}, args...)
	listArgs = append(listArgs, pageSize, offset)

	var rows []struct {
		models.Product
		StoredInJSON   []byte `db:"stored_in_json"`
		CategoriesJSON []byte `db:"categories_json"`
		DeckCardsJSON  []byte `db:"deck_cards_json"`
	}

	if err := h.DB.Select(&rows, listQuery, listArgs...); err != nil {
		logger.Error("Error selecting products: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	products := make([]models.Product, len(rows))
	for i, r := range rows {
		products[i] = r.Product
		if r.StoredInJSON != nil {
			json.Unmarshal(r.StoredInJSON, &products[i].StoredIn)
		}
		if r.CategoriesJSON != nil {
			json.Unmarshal(r.CategoriesJSON, &products[i].Categories)
			if !isAdmin {
				filtered := []models.CustomCategory{}
				for _, c := range products[i].Categories {
					if c.ShowBadge {
						filtered = append(filtered, c)
					}
				}
				products[i].Categories = filtered
			}
		}
		if r.DeckCardsJSON != nil {
			json.Unmarshal(r.DeckCardsJSON, &products[i].DeckCards)
		}
	}

	h.populatePrices(products)
	h.populateStorage(products)
	h.populateCategories(products, isAdmin)
	h.populateCartCounts(products)

	render.Success(w, models.ProductListResponse{
		Products:    products,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		Facets:      h.getFacets(tcg, category, search, storageID, foil, treatment, condition, rarity, language, color, collection, setName, inStock, filterLogic, isAdmin),
		QueryTimeMS: time.Since(start).Milliseconds(),
	})
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	isAdmin := strings.Contains(r.URL.Path, "/admin/")

	if !isAdmin {
		var active bool
		err := h.DB.Get(&active, "SELECT COALESCE(t.is_active, true) FROM product p LEFT JOIN tcg t ON p.tcg = t.id WHERE p.id = $1", id)
		if err != nil || !active {
			render.Error(w, "Product not found or unavailable", http.StatusNotFound)
			return
		}
	}

	var jsonResult []byte
	err := h.DB.Get(&jsonResult, "SELECT fn_get_product_detail($1)", id)
	if err != nil {
		logger.Error("Error calling fn_get_product_detail: %v", err)
		render.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	var product models.Product
	if err := json.Unmarshal(jsonResult, &product); err != nil {
		logger.Error("Error unmarshaling product detail: %v", err)
		render.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.populatePrices([]models.Product{product})
	h.populateCartCounts([]models.Product{product})
	render.Success(w, product)
}

func (h *ProductHandler) ListTCGs(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active_only") == "true"
	var tcgs []models.TCG
	query := "SELECT * FROM tcg ORDER BY name"
	if activeOnly {
		query = "SELECT * FROM tcg WHERE is_active = true ORDER BY name"
	}

	err := h.DB.Select(&tcgs, query)
	if err != nil {
		logger.Error("Error listing TCGs: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if tcgs == nil {
		tcgs = []models.TCG{}
	}
	render.Success(w, map[string]interface{}{"tcgs": tcgs})
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.ProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Name == "" || input.TCG == "" || input.Category == "" {
		render.Error(w, "name, tcg, and category are required", http.StatusBadRequest)
		return
	}

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
	err := h.DB.QueryRowx(`
		INSERT INTO product (name, tcg, category, set_name, set_code, condition,
		                      foil_treatment, card_treatment,
		                      price_reference, price_source, price_cop_override,
		                      stock, image_url, description, collector_number, promo_type,
		                      language, color_identity, rarity, cmc, is_legendary, is_historic, is_land, is_basic_land, art_variation,
		                      oracle_text, artist, type_line, border_color, frame, full_art, textless)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32)
		RETURNING *
	`, input.Name, input.TCG, input.Category, input.SetName, input.SetCode, input.Condition,
		input.FoilTreatment, input.CardTreatment,
		input.PriceReference, input.PriceSource, input.PriceCOPOverride,
		input.Stock, input.ImageURL, input.Description, input.CollectorNumber, input.PromoType,
		input.Language, input.ColorIdentity, input.Rarity, input.CMC, input.IsLegendary, input.IsHistoric, input.IsLand, input.IsBasicLand, input.ArtVariation,
		input.OracleText, input.Artist, input.TypeLine, input.BorderColor, input.Frame, input.FullArt, input.Textless,
	).StructScan(&product)

	if err != nil {
		render.Error(w, "Failed to create product: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.saveProductCategories(product.ID, input.CategoryIDs)
	h.saveProductStorage(product.ID, input.StorageItems)
	if input.Category == "store_exclusives" {
		h.saveDeckCards(product.ID, input.DeckCards)
		product.DeckCards = input.DeckCards
	}

	products := []models.Product{product}
	h.populatePrices(products)
	h.populateStorage(products)
	h.populateCategories(products, true)
	w.WriteHeader(http.StatusCreated)
	render.Success(w, products[0])
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input models.ProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.PriceSource == "" {
		input.PriceSource = models.PriceSourceManual
	}

	var product models.Product
	err := h.DB.QueryRowx(`
		UPDATE product
		SET name=$1, tcg=$2, category=$3, set_name=$4, set_code=$5, condition=$6,
		    foil_treatment=$7, card_treatment=$8,
		    price_reference=$9, price_source=$10, price_cop_override=$11,
		    stock=$12, image_url=$13, description=$14, collector_number=$15, promo_type=$16,
		    language=$17, color_identity=$18, rarity=$19, cmc=$20, is_legendary=$21, is_historic=$22, is_land=$23, is_basic_land=$24, art_variation=$25,
		    oracle_text=$26, artist=$27, type_line=$28, border_color=$29, frame=$30, full_art=$31, textless=$32
		WHERE id=$33
		RETURNING *
	`, input.Name, input.TCG, input.Category, input.SetName, input.SetCode, input.Condition,
		input.FoilTreatment, input.CardTreatment,
		input.PriceReference, input.PriceSource, input.PriceCOPOverride,
		input.Stock, input.ImageURL, input.Description, input.CollectorNumber, input.PromoType,
		input.Language, input.ColorIdentity, input.Rarity, input.CMC, input.IsLegendary, input.IsHistoric, input.IsLand, input.IsBasicLand, input.ArtVariation,
		input.OracleText, input.Artist, input.TypeLine, input.BorderColor, input.Frame, input.FullArt, input.Textless,
		id,
	).StructScan(&product)

	if err != nil {
		logger.Error("Update product %s failed: %v", id, err)
		render.Error(w, "Product not found or update failed", http.StatusNotFound)
		return
	}

	h.saveProductCategories(product.ID, input.CategoryIDs)
	h.saveProductStorage(product.ID, input.StorageItems)
	if input.Category == "store_exclusives" {
		h.saveDeckCards(product.ID, input.DeckCards)
		product.DeckCards = input.DeckCards
	}

	products := []models.Product{product}
	h.populatePrices(products)
	h.populateStorage(products)
	h.populateCategories(products, true)
	render.Success(w, products[0])
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result, err := h.DB.Exec("DELETE FROM product WHERE id = $1", id)
	if err != nil {
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		render.Error(w, "Product not found", http.StatusNotFound)
		return
	}
	render.Success(w, map[string]string{"message": "Product deleted"})
}

func (h *ProductHandler) BulkCreate(w http.ResponseWriter, r *http.Request) {
	var inputs []models.ProductInput
	if err := json.NewDecoder(r.Body).Decode(&inputs); err != nil {
		render.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(inputs) == 0 {
		render.Success(w, map[string]interface{}{"message": "No products to import", "count": 0})
		return
	}

	jsonData, err := json.Marshal(inputs)
	if err != nil {
		render.Error(w, "Failed to encode products for database", http.StatusInternalServerError)
		return
	}

	var ids []struct {
		ID string `db:"upserted_id"`
	}
	err = h.DB.Select(&ids, "SELECT upserted_id FROM fn_bulk_upsert_product($1)", string(jsonData))
	if err != nil {
		logger.Error("Bulk upsert failed: %v", err)
		render.Error(w, "Database failure during bulk import: "+err.Error(), http.StatusInternalServerError)
		return
	}

	render.Success(w, map[string]interface{}{
		"message": fmt.Sprintf("Successfully imported %d products", len(ids)),
		"count":   len(ids),
	})
}

func (h *ProductHandler) GetStorage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var items []models.StorageLocation
	err := h.DB.Select(&items, `
		SELECT s.id as storage_id, s.name, COALESCE(ps.quantity, 0) as quantity
		FROM storage_location s
		LEFT JOIN product_storage ps ON s.id = ps.storage_id AND ps.product_id = $1
		ORDER BY s.name
	`, id)
	if err != nil {
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if items == nil {
		items = []models.StorageLocation{}
	}
	render.Success(w, items)
}

func (h *ProductHandler) UpdateStorage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var updates []models.ProductStorage
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		render.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`DELETE FROM product_storage WHERE product_id = $1`, id)
	if err != nil {
		tx.Rollback()
		render.Error(w, "Failed to clear existing storage", http.StatusInternalServerError)
		return
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
			render.Error(w, "Failed to update storage map", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		render.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}
	h.GetStorage(w, r)
}

func (h *ProductHandler) getFacets(tcg, category, search, storageID, foil, treatment, condition, rarity, language, color, collection, setName string, inStock bool, filterLogic string, isAdmin bool) models.Facets {
	var result []byte
	err := h.DB.Get(&result, "SELECT fn_get_product_facets($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)",
		tcg, category, search, storageID, foil, treatment, condition, rarity, language, color, collection, setName, inStock, filterLogic, isAdmin)

	if err != nil {
		logger.Error("Error calling fn_get_product_facets: %v", err)
		return models.Facets{}
	}

	var facets models.Facets
	if err := json.Unmarshal(result, &facets); err != nil {
		logger.Error("Error unmarshaling facets: %v", err)
		return models.Facets{}
	}
	return facets
}

func (h *ProductHandler) buildOrderBy(sortBy, sortDir, search string, argsLen int) string {
	dir := "DESC"
	if strings.EqualFold(sortDir, "asc") {
		dir = "ASC"
	}
	if sortBy == "" {
		if search != "" {
			return "similarity(p.name, $" + strconv.Itoa(argsLen-1) + ") DESC, p.created_at DESC"
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
	case "price":
		s, err := loadSettings(h.DB)
		if err != nil {
			s = models.Settings{USDToCOPRate: 4200, EURToCOPRate: 4600}
		}
		usdRate := strconv.FormatFloat(s.USDToCOPRate, 'f', 4, 64)
		eurRate := strconv.FormatFloat(s.EURToCOPRate, 'f', 4, 64)
		col = fmt.Sprintf(`COALESCE(p.price_cop_override,
			CASE p.price_source
				WHEN 'tcgplayer' THEN p.price_reference * %s
				WHEN 'cardmarket' THEN p.price_reference * %s
				ELSE 0
			END, 0)`, usdRate, eurRate)
	case "created_at":
		col = "p.created_at"
	case "updated_at":
		col = "p.updated_at"
	default:
		if search != "" {
			return "similarity(p.name, $" + strconv.Itoa(argsLen-1) + ") DESC, p.created_at DESC"
		}
		return "p.created_at DESC"
	}
	return col + " " + dir + ", p.created_at DESC"
}

func (h *ProductHandler) buildFilters(tcg, category, search, storageID, foil, treatment, condition, collection, rarity, language, color, setName string, inStock bool, filterLogic string, isAdmin bool) (string, []string, []interface{}) {
	fromClause := "FROM product p"
	builder := sqlutil.NewBuilder(fromClause)

	var mandatory []string
	var optional []string

	if !isAdmin {
		builder.BaseQuery += " LEFT JOIN tcg t ON p.tcg = t.id"
		mandatory = append(mandatory, "(t.is_active IS NULL OR t.is_active = true)")
	}

	if storageID != "" {
		builder.BaseQuery = "FROM product p JOIN product_storage ps ON p.id = ps.product_id"
		placeholder := fmt.Sprintf("$%d", len(builder.Args)+1)
		builder.Args = append(builder.Args, storageID)
		mandatory = append(mandatory, "ps.storage_id = "+placeholder)
		mandatory = append(mandatory, "ps.quantity > 0")
	}

	if tcg != "" {
		placeholder := fmt.Sprintf("$%d", len(builder.Args)+1)
		builder.Args = append(builder.Args, strings.ToLower(tcg))
		mandatory = append(mandatory, "p.tcg = "+placeholder)
	}
	if category != "" {
		placeholder := fmt.Sprintf("$%d", len(builder.Args)+1)
		builder.Args = append(builder.Args, strings.ToLower(category))
		mandatory = append(mandatory, "p.category = "+placeholder)
	}

	if inStock {
		mandatory = append(mandatory, "p.stock > 0")
	}
	if search != "" {
		searchTerms := strings.Fields(search)
		for _, term := range searchTerms {
			placeholderIdx := len(builder.Args) + 1
			for j := 0; j < 8; j++ {
				builder.Args = append(builder.Args, "%"+term+"%")
			}
			p1 := fmt.Sprintf("$%d", placeholderIdx)
			p2 := fmt.Sprintf("$%d", placeholderIdx+1)
			p3 := fmt.Sprintf("$%d", placeholderIdx+2)
			p4 := fmt.Sprintf("$%d", placeholderIdx+3)
			p5 := fmt.Sprintf("$%d", placeholderIdx+4)
			p6 := fmt.Sprintf("$%d", placeholderIdx+5)
			p7 := fmt.Sprintf("$%d", placeholderIdx+6)
			p8 := fmt.Sprintf("$%d", placeholderIdx+7)
			mandatory = append(mandatory, fmt.Sprintf("(p.name ILIKE %s OR p.set_name ILIKE %s OR p.set_code ILIKE %s OR p.artist ILIKE %s OR p.collector_number ILIKE %s OR p.oracle_text ILIKE %s OR p.type_line ILIKE %s OR p.promo_type ILIKE %s)", p1, p2, p3, p4, p5, p6, p7, p8))
		}
	}

	opLogic := " OR "
	if strings.ToLower(filterLogic) == "and" {
		opLogic = " AND "
	}

	if foil != "" {
		vals := strings.Split(foil, ",")
		var conds []string
		for _, v := range vals {
			placeholder := fmt.Sprintf("$%d", len(builder.Args)+1)
			conds = append(conds, "LOWER(p.foil_treatment) = "+placeholder)
			builder.Args = append(builder.Args, strings.ToLower(v))
		}
		optional = append(optional, "("+strings.Join(conds, opLogic)+")")
	}
	if treatment != "" {
		vals := strings.Split(treatment, ",")
		var conds []string
		for _, v := range vals {
			lv := strings.ToLower(v)
			placeholder := fmt.Sprintf("$%d", len(builder.Args)+1)
			cond := "LOWER(p.card_treatment) = " + placeholder
			switch lv {
			case "textless":
				cond = "(" + cond + " OR p.textless = true)"
			case "full_art":
				cond = "(" + cond + " OR p.full_art = true)"
			}
			conds = append(conds, cond)
			builder.Args = append(builder.Args, lv)
		}
		optional = append(optional, "("+strings.Join(conds, opLogic)+")")
	}
	if condition != "" {
		vals := strings.Split(condition, ",")
		var conds []string
		for _, v := range vals {
			placeholder := fmt.Sprintf("$%d", len(builder.Args)+1)
			conds = append(conds, "p.condition = "+placeholder)
			builder.Args = append(builder.Args, strings.ToUpper(v))
		}
		optional = append(optional, "("+strings.Join(conds, opLogic)+")")
	}
	if collection != "" {
		vals := strings.Split(collection, ",")
		var conds []string
		for _, v := range vals {
			placeholder := fmt.Sprintf("$%d", len(builder.Args)+1)
			conds = append(conds, "EXISTS (SELECT 1 FROM product_category pc_col JOIN custom_category c_col ON pc_col.category_id = c_col.id WHERE pc_col.product_id = p.id AND c_col.slug = "+placeholder+")")
			builder.Args = append(builder.Args, strings.ToLower(v))
		}
		optional = append(optional, "("+strings.Join(conds, opLogic)+")")
	}
	if rarity != "" {
		vals := strings.Split(rarity, ",")
		var conds []string
		for _, v := range vals {
			placeholder := fmt.Sprintf("$%d", len(builder.Args)+1)
			conds = append(conds, "LOWER(p.rarity) = "+placeholder)
			builder.Args = append(builder.Args, strings.ToLower(v))
		}
		optional = append(optional, "("+strings.Join(conds, opLogic)+")")
	}
	if language != "" {
		vals := strings.Split(language, ",")
		var conds []string
		for _, v := range vals {
			placeholder := fmt.Sprintf("$%d", len(builder.Args)+1)
			conds = append(conds, "LOWER(p.language) = "+placeholder)
			builder.Args = append(builder.Args, strings.ToLower(v))
		}
		optional = append(optional, "("+strings.Join(conds, opLogic)+")")
	}
	if color != "" {
		vals := strings.Split(color, ",")
		var conds []string
		for _, v := range vals {
			placeholder := fmt.Sprintf("$%d", len(builder.Args)+1)
			conds = append(conds, "p.color_identity ILIKE "+placeholder)
			builder.Args = append(builder.Args, "%"+strings.ToUpper(v)+"%")
		}
		optional = append(optional, "("+strings.Join(conds, opLogic)+")")
	}
	if setName != "" {
		vals := strings.Split(setName, ",")
		var conds []string
		for _, v := range vals {
			placeholder := fmt.Sprintf("$%d", len(builder.Args)+1)
			conds = append(conds, "p.set_name = "+placeholder)
			builder.Args = append(builder.Args, v)
		}
		optional = append(optional, "("+strings.Join(conds, opLogic)+")")
	}

	finalConditions := mandatory
	if len(optional) > 0 {
		finalConditions = append(finalConditions, "("+strings.Join(optional, opLogic)+")")
	}

	return builder.BaseQuery, finalConditions, builder.Args
}
