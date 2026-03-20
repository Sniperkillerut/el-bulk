package handlers

import (
	"encoding/json"
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

// GET /api/products
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	tcg := q.Get("tcg")
	category := q.Get("category")
	search := q.Get("search")
	foil := q.Get("foil")
	treatment := q.Get("treatment")
	condition := q.Get("condition")
	featuredOnly := q.Get("featured") == "true"

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
	if featuredOnly {
		conditions = append(conditions, "featured = TRUE")
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
	if err := h.DB.Get(&total, "SELECT COUNT(*) FROM products "+where, args...); err != nil {
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	listQuery := "SELECT * FROM products " + where + " ORDER BY featured DESC, created_at DESC LIMIT $" + strconv.Itoa(idx) + " OFFSET $" + strconv.Itoa(idx+1)
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
	h.populatePrices([]models.Product{product})
	jsonOK(w, product)
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
		                      stock, image_url, description, featured)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING *
	`, input.Name, input.TCG, input.Category, input.SetName, input.SetCode, input.Condition,
		input.FoilTreatment, input.CardTreatment,
		input.PriceReference, input.PriceSource, input.PriceCOPOverride,
		input.Stock, input.ImageURL, input.Description, input.Featured,
	).StructScan(&product)

	if err != nil {
		jsonError(w, "Failed to create product: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.populatePrices([]models.Product{product})
	w.WriteHeader(http.StatusCreated)
	jsonOK(w, product)
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
		    stock=$12, image_url=$13, description=$14, featured=$15
		WHERE id=$16
		RETURNING *
	`, input.Name, input.TCG, input.Category, input.SetName, input.SetCode, input.Condition,
		input.FoilTreatment, input.CardTreatment,
		input.PriceReference, input.PriceSource, input.PriceCOPOverride,
		input.Stock, input.ImageURL, input.Description, input.Featured, id,
	).StructScan(&product)

	if err != nil {
		jsonError(w, "Product not found or update failed", http.StatusNotFound)
		return
	}

	h.populatePrices([]models.Product{product})
	jsonOK(w, product)
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
