package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/el-bulk/backend/utils/render"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/store"
	"github.com/el-bulk/backend/utils/httputil"
	"github.com/el-bulk/backend/utils/logger"
)

type ProductHandler struct {
	Service *service.ProductService
	DB      *sqlx.DB // Kept for settings loading, though service could handle it too
}

func NewProductHandler(s *service.ProductService, db *sqlx.DB) *ProductHandler {
	return &ProductHandler{Service: s, DB: db}
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	q := r.URL.Query()

	isAdmin, _ := r.Context().Value(middleware.IsAdminKey).(bool)
	maxPageSize := 100
	if isAdmin {
		maxPageSize = 5000
	}
	page, pageSize, offset := httputil.GetPagination(r, 20, maxPageSize)

	params := store.ProductFilterParams{
		TCG:         q.Get("tcg"),
		Category:    q.Get("category"),
		Search:      q.Get("search"),
		StorageID:   q.Get("storage_id"),
		Foil:        q.Get("foil"),
		Treatment:   q.Get("treatment"),
		Condition:   q.Get("condition"),
		Collection:  q.Get("collection"),
		Rarity:      q.Get("rarity"),
		Language:    q.Get("language"),
		Color:       q.Get("color"),
		SetName:     q.Get("set_name"),
		InStock:     q.Get("in_stock") == "true",
		SortBy:      q.Get("sort_by"),
		SortDir:     q.Get("sort_dir"),
		FilterLogic: q.Get("logic"),
		Page:        page,
		PageSize:    pageSize,
		Offset:      offset,
	}

	resp, err := h.Service.List(params, isAdmin)
	if err != nil {
		logger.Error("List products failed: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	resp.QueryTimeMS = time.Since(start).Milliseconds()
	render.Success(w, resp)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	isAdmin, _ := r.Context().Value(middleware.IsAdminKey).(bool)

	product, err := h.Service.GetByID(id, isAdmin)
	if err != nil {
		logger.Error("Get product %s failed: %v", id, err)
		render.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	render.Success(w, product)
}

func (h *ProductHandler) ListTCGs(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active_only") == "true"
	tcgs, err := h.Service.GetTCGs(activeOnly)
	if err != nil {
		logger.Error("List TCGs failed: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
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

	product, err := h.Service.Create(input)
	if err != nil {
		logger.Error("Create product failed: %v", err)
		render.Error(w, "Failed to create product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.Success(w, product)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input models.ProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	product, err := h.Service.Update(id, input)
	if err != nil {
		logger.Error("Update product %s failed: %v", id, err)
		render.Error(w, "Product not found or update failed", http.StatusNotFound)
		return
	}

	render.Success(w, product)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.Service.Delete(id); err != nil {
		logger.Error("Delete product %s failed: %v", id, err)
		render.Error(w, "Delete failed", http.StatusInternalServerError)
		return
	}
	render.Success(w, map[string]string{"message": "Product deleted"})
}

func (h *ProductHandler) BulkCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Products    []models.ProductInput `json:"products"`
		CategoryIDs []string              `json:"category_ids,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Try fallback if needed, but we'll stick to structured req
		render.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	count, err := h.Service.BulkCreate(req.Products, req.CategoryIDs)
	if err != nil {
		logger.Error("BulkCreate failed: %v", err)
		render.Error(w, "Bulk import failed", http.StatusInternalServerError)
		return
	}

	render.Success(w, map[string]interface{}{
		"message": fmt.Sprintf("Successfully imported %d products", count),
		"count":   count,
	})
}

func (h *ProductHandler) BulkSearch(w http.ResponseWriter, r *http.Request) {
	var req models.BulkSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	results, err := h.Service.BulkSearch(req.List)
	if err != nil {
		logger.Error("BulkSearch failed: %v", err)
		render.Error(w, "Bulk search failed", http.StatusInternalServerError)
		return
	}

	render.Success(w, models.BulkSearchResponse{Matches: results})
}

func (h *ProductHandler) GetLowStock(w http.ResponseWriter, r *http.Request) {
	thresholdStr := r.URL.Query().Get("threshold")
	threshold := 5
	if t, err := strconv.Atoi(thresholdStr); err == nil {
		threshold = t
	}

	products, err := h.Service.GetLowStock(threshold)
	if err != nil {
		logger.Error("GetLowStock failed: %v", err)
		render.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}

	render.Success(w, products)
}

func (h *ProductHandler) GetStorage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	items, err := h.Service.GetStorage(id)
	if err != nil {
		logger.Error("GetStorage %s failed: %v", id, err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
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

	items, err := h.Service.UpdateStorage(id, updates)
	if err != nil {
		logger.Error("UpdateStorage %s failed: %v", id, err)
		render.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}

	render.Success(w, items)
}
