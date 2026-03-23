package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/el-bulk/backend/models"
)

type ProductHandler struct {
	DB *sqlx.DB
}

func NewProductHandler(db *sqlx.DB) *ProductHandler {
	return &ProductHandler{DB: db}
}

// populatePrices fetches current exchange rates and sets the computed COP Price
// on each product in the slice.
func (h *ProductHandler) populatePrices(products []models.Product) {
	s, err := loadSettings(h.DB)
	if err != nil {
		// Fall back to safe defaults if settings are unavailable
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
		FROM product_stored_in ps 
		JOIN stored_in s ON ps.stored_in_id = s.id 
		WHERE ps.quantity > 0 AND ps.product_id IN (?)
	`, pids)
	if err != nil {
		return
	}
	
	query = h.DB.Rebind(query)
	var storageRows []struct {
		ProductID  string `db:"product_id"`
		StoredInID string `db:"stored_in_id"`
		Name       string `db:"name"`
		Quantity   int    `db:"quantity"`
	}
	if err := h.DB.Select(&storageRows, query, args...); err != nil {
		return
	}

	storageMap := make(map[string][]models.StorageLocation)
	for _, r := range storageRows {
		storageMap[r.ProductID] = append(storageMap[r.ProductID], models.StorageLocation{
			StoredInID: r.StoredInID,
			Name:       r.Name,
			Quantity:   r.Quantity,
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

func (h *ProductHandler) saveProductCategories(productID string, categoryIDs []string) {
	log.Printf("saveProductCategories called for Product: %s with categories: %v\n", productID, categoryIDs)
	_, err := h.DB.Exec("DELETE FROM product_categories WHERE product_id = $1", productID)
	if err != nil {
		log.Printf("Error deleting product_categories: %v\n", err)
	}
	for _, cid := range categoryIDs {
		_, err := h.DB.Exec("INSERT INTO product_categories (product_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", productID, cid)
		if err != nil {
			log.Printf("Error inserting product_categories (product=%s, cat=%s): %v\n", productID, cid, err)
		}
	}
}

func (h *ProductHandler) populateCategories(products []models.Product) {
	if len(products) == 0 {
		return
	}
	var pids []string
	for _, p := range products {
		pids = append(pids, p.ID)
	}

	query, args, err := sqlx.In(`
		SELECT pc.product_id, c.id, c.name, c.slug
		FROM product_categories pc
		JOIN custom_categories c ON pc.category_id = c.id
		WHERE pc.product_id IN (?)
		ORDER BY c.name
	`, pids)
	if err != nil {
		log.Printf("Error creating IN query for populateCategories: %v\n", err)
		return
	}
	
	query = h.DB.Rebind(query)
	var catRows []struct {
		ProductID string    `db:"product_id"`
		ID        string    `db:"id"`
		Name      string    `db:"name"`
		Slug      string    `db:"slug"`
	}
	if err := h.DB.Select(&catRows, query, args...); err != nil {
		log.Printf("Error selecting categories: %v\n", err)
		return
	}

	catMap := make(map[string][]models.CustomCategory)
	for _, r := range catRows {
		catMap[r.ProductID] = append(catMap[r.ProductID], models.CustomCategory{
			ID:   r.ID,
			Name: r.Name,
			Slug: r.Slug,
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

// GET /api/products
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	tcg := q.Get("tcg")
	category := q.Get("category")
	search := q.Get("search")
	storageID := q.Get("storage_id")
	foil := q.Get("foil")
	treatment := q.Get("treatment")
	condition := q.Get("condition")
	collection := q.Get("collection")

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var conditions []string
	var args []interface{}
	idx := 1

	fromClause := "FROM products"
	if storageID != "" {
		fromClause = "FROM products JOIN product_stored_in ps ON products.id = ps.product_id"
		conditions = append(conditions, "ps.stored_in_id = $"+strconv.Itoa(idx), "ps.quantity > 0")
		args = append(args, storageID)
		idx++
	}

	if tcg != "" {
		conditions = append(conditions, "tcg = $"+strconv.Itoa(idx))
		args = append(args, strings.ToLower(tcg))
		idx++
	}
	if category != "" {
		conditions = append(conditions, "category = $"+strconv.Itoa(idx))
		args = append(args, strings.ToLower(category))
		idx++
	}
	if foil != "" {
		conditions = append(conditions, "foil_treatment = $"+strconv.Itoa(idx))
		args = append(args, foil)
		idx++
	}
	if treatment != "" {
		conditions = append(conditions, "card_treatment = $"+strconv.Itoa(idx))
		args = append(args, treatment)
		idx++
	}
	if condition != "" {
		conditions = append(conditions, "condition = $"+strconv.Itoa(idx))
		args = append(args, strings.ToUpper(condition))
		idx++
	}
	if collection != "" {
		fromClause += " JOIN product_categories pc_col ON products.id = pc_col.product_id JOIN custom_categories c_col ON pc_col.category_id = c_col.id"
		conditions = append(conditions, "c_col.slug = $"+strconv.Itoa(idx))
		args = append(args, collection)
		idx++
	}
	if search != "" {
		conditions = append(conditions, "to_tsvector('english', name || ' ' || COALESCE(set_name, '')) @@ plainto_tsquery('english', $"+strconv.Itoa(idx)+")")
		args = append(args, search)
		idx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	if err := h.DB.Get(&total, "SELECT COUNT(*) "+fromClause+" "+where, args...); err != nil {
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	listQuery := "SELECT products.* " + fromClause + " " + where + " ORDER BY products.created_at DESC LIMIT $" + strconv.Itoa(idx) + " OFFSET $" + strconv.Itoa(idx+1)
	args = append(args, pageSize, offset)

	var products []models.Product
	if err := h.DB.Select(&products, listQuery, args...); err != nil {
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	if products == nil {
		products = []models.Product{}
	}

	h.populatePrices(products)
	h.populateStorage(products)
	h.populateCategories(products)

	jsonOK(w, models.ProductListResponse{
		Products: products,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GET /api/products/:id
func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var product models.Product
	err := h.DB.Get(&product, "SELECT * FROM products WHERE id = $1", id)
	if err != nil {
		jsonError(w, "Product not found", http.StatusNotFound)
		return
	}
	
	products := []models.Product{product}
	h.populatePrices(products)
	h.populateStorage(products)
	h.populateCategories(products)
	jsonOK(w, products[0])
}

// GET /api/tcgs
func (h *ProductHandler) ListTCGs(w http.ResponseWriter, r *http.Request) {
	var tcgs []string
	err := h.DB.Select(&tcgs, "SELECT DISTINCT tcg FROM products ORDER BY tcg")
	if err != nil {
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string][]string{"tcgs": tcgs})
}

// POST /api/admin/products
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.ProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Name == "" || input.TCG == "" || input.Category == "" {
		jsonError(w, "name, tcg, and category are required", http.StatusBadRequest)
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
		INSERT INTO products (name, tcg, category, set_name, set_code, condition,
		                      foil_treatment, card_treatment,
		                      price_reference, price_source, price_cop_override,
		                      stock, image_url, description)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING *
	`, input.Name, input.TCG, input.Category, input.SetName, input.SetCode, input.Condition,
		input.FoilTreatment, input.CardTreatment,
		input.PriceReference, input.PriceSource, input.PriceCOPOverride,
		input.Stock, input.ImageURL, input.Description,
	).StructScan(&product)

	if err != nil {
		jsonError(w, "Failed to create product: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.saveProductCategories(product.ID, input.CategoryIDs)

	products := []models.Product{product}
	h.populatePrices(products)
	h.populateStorage(products)
	h.populateCategories(products)
	w.WriteHeader(http.StatusCreated)
	jsonOK(w, products[0])
}

// PUT /api/admin/products/:id
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var input models.ProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.PriceSource == "" {
		input.PriceSource = models.PriceSourceManual
	}

	var product models.Product
	err := h.DB.QueryRowx(`
		UPDATE products
		SET name=$1, tcg=$2, category=$3, set_name=$4, set_code=$5, condition=$6,
		    foil_treatment=$7, card_treatment=$8,
		    price_reference=$9, price_source=$10, price_cop_override=$11,
		    stock=$12, image_url=$13, description=$14
		WHERE id=$15
		RETURNING *
	`, input.Name, input.TCG, input.Category, input.SetName, input.SetCode, input.Condition,
		input.FoilTreatment, input.CardTreatment,
		input.PriceReference, input.PriceSource, input.PriceCOPOverride,
		input.Stock, input.ImageURL, input.Description, id,
	).StructScan(&product)

	if err != nil {
		jsonError(w, "Product not found or update failed", http.StatusNotFound)
		return
	}

	h.saveProductCategories(product.ID, input.CategoryIDs)

	products := []models.Product{product}
	h.populatePrices(products)
	h.populateStorage(products)
	h.populateCategories(products)
	jsonOK(w, products[0])
}

// DELETE /api/admin/products/:id
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := h.DB.Exec("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		jsonError(w, "Product not found", http.StatusNotFound)
		return
	}

	jsonOK(w, map[string]string{"message": "Product deleted"})
}

// GET /api/admin/products/:id/storage
func (h *ProductHandler) GetStorage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var items []models.StorageLocation
	err := h.DB.Select(&items, `
		SELECT s.id as stored_in_id, s.name, COALESCE(ps.quantity, 0) as quantity
		FROM stored_in s
		LEFT JOIN product_stored_in ps ON s.id = ps.stored_in_id AND ps.product_id = $1
		ORDER BY s.name
	`, id)
	if err != nil {
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}
	if items == nil {
		items = []models.StorageLocation{}
	}
	jsonOK(w, items)
}

// PUT /api/admin/products/:id/storage
func (h *ProductHandler) UpdateStorage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var updates []models.ProductStorage
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		jsonError(w, "Invalid input", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	// 1. Clear existing
	_, err = tx.Exec(`DELETE FROM product_stored_in WHERE product_id = $1`, id)
	if err != nil {
		tx.Rollback()
		jsonError(w, "Failed to clear existing storage", http.StatusInternalServerError)
		return
	}

	// 2. Insert active
	for _, u := range updates {
		if u.Quantity > 0 {
			_, err = tx.Exec(`
				INSERT INTO product_stored_in (product_id, stored_in_id, quantity)
				VALUES ($1, $2, $3)
			`, id, u.StoredInID, u.Quantity)
			if err != nil {
				tx.Rollback()
				jsonError(w, "Failed to update storage map", http.StatusInternalServerError)
				return
			}
		}
	}
	tx.Commit()

	h.GetStorage(w, r)
}

// Helpers
func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.ErrorResponse{Error: msg})
}
