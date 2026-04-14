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
	logger.TraceCtx(r.Context(), "Entering ProductHandler.List | URL: %s", r.URL.String())
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

	resp, err := h.Service.List(r.Context(), params, isAdmin)
	if err != nil {
		logger.ErrorCtx(r.Context(), "List products failed: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	resp.QueryTimeMS = time.Since(start).Milliseconds()
	render.Success(w, resp)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.TraceCtx(r.Context(), "Entering ProductHandler.GetByID | ID: %s", id)
	if id == "" {
		render.Error(w, "Product ID is required", http.StatusBadRequest)
		return
	}
	isAdmin, _ := r.Context().Value(middleware.IsAdminKey).(bool)

	product, err := h.Service.GetByID(r.Context(), id, isAdmin)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Get product %s failed: %v", id, err)
		render.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	render.Success(w, product)
}

func (h *ProductHandler) ListTCGs(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active_only") == "true"
	tcgs, err := h.Service.GetTCGs(r.Context(), activeOnly)
	if err != nil {
		logger.ErrorCtx(r.Context(), "List TCGs failed: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	render.Success(w, map[string]interface{}{"tcgs": tcgs})
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering ProductHandler.Create")
	var input models.ProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode product input: %v", err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Name == "" || input.TCG == "" || input.Category == "" {
		render.Error(w, "name, tcg, and category are required", http.StatusBadRequest)
		return
	}

	product, err := h.Service.Create(r.Context(), input)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Create product failed: %v", err)
		render.Error(w, "Failed to create product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.Success(w, product)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.TraceCtx(r.Context(), "Entering ProductHandler.Update | ID: %s", id)
	if id == "" {
		render.Error(w, "Product ID is required", http.StatusBadRequest)
		return
	}
	var input models.ProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode update input for %s: %v", id, err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	product, err := h.Service.Update(r.Context(), id, input)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Update product %s failed: %v", id, err)
		render.Error(w, "Product not found or update failed", http.StatusNotFound)
		return
	}

	render.Success(w, product)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.TraceCtx(r.Context(), "Entering ProductHandler.Delete | ID: %s", id)
	if id == "" {
		render.Error(w, "Product ID is required", http.StatusBadRequest)
		return
	}
	if err := h.Service.Delete(r.Context(), id); err != nil {
		logger.ErrorCtx(r.Context(), "Delete product %s failed: %v", id, err)
		render.Error(w, "Delete failed", http.StatusInternalServerError)
		return
	}
	render.Success(w, map[string]string{"message": "Product deleted"})
}

func (h *ProductHandler) BulkCreate(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering ProductHandler.BulkCreate")
	var req struct {
		Products    []models.ProductInput `json:"products"`
		CategoryIDs []string              `json:"category_ids,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode bulk create input: %v", err)
		render.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	count, err := h.Service.BulkCreate(r.Context(), req.Products, req.CategoryIDs)
	if err != nil {
		logger.ErrorCtx(r.Context(), "BulkCreate failed: %v", err)
		render.Error(w, "Bulk import failed", http.StatusInternalServerError)
		return
	}

	render.Success(w, map[string]interface{}{
		"message": fmt.Sprintf("Successfully imported %d products", count),
		"count":   count,
	})
}

func (h *ProductHandler) BulkSearch(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering ProductHandler.BulkSearch")
	var req models.BulkSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode bulk search input: %v", err)
		render.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	results, err := h.Service.BulkSearch(r.Context(), req.List)
	if err != nil {
		logger.ErrorCtx(r.Context(), "BulkSearch failed: %v", err)
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

	products, err := h.Service.GetLowStock(r.Context(), threshold)
	if err != nil {
		logger.ErrorCtx(r.Context(), "GetLowStock failed: %v", err)
		render.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}

	render.Success(w, products)
}

func (h *ProductHandler) GetStorage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	items, err := h.Service.GetStorage(r.Context(), id)
	if err != nil {
		logger.ErrorCtx(r.Context(), "GetStorage %s failed: %v", id, err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	render.Success(w, items)
}

func (h *ProductHandler) UpdateStorage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.TraceCtx(r.Context(), "Entering ProductHandler.UpdateStorage | ID: %s", id)
	var updates []models.ProductStorage
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode storage update for %s: %v", id, err)
		render.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	items, err := h.Service.UpdateStorage(r.Context(), id, updates)
	if err != nil {
		logger.ErrorCtx(r.Context(), "UpdateStorage %s failed: %v", id, err)
		render.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}

	render.Success(w, items)
}

func (h *ProductHandler) BulkUpdateSource(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering ProductHandler.BulkUpdateSource")
	var req struct {
		IDs    []string           `json:"ids"`
		Source models.PriceSource `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode bulk update source input: %v", err)
		render.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if len(req.IDs) == 0 || req.Source == "" {
		render.Error(w, "ids and source are required", http.StatusBadRequest)
		return
	}

	count, err := h.Service.BulkUpdateSource(r.Context(), req.IDs, req.Source)
	if err != nil {
		logger.ErrorCtx(r.Context(), "BulkUpdateSource failed: %v", err)
		render.Error(w, "Bulk update failed", http.StatusInternalServerError)
		return
	}

	render.Success(w, map[string]interface{}{
		"message": fmt.Sprintf("Successfully updated %d products to source %s", count, req.Source),
		"count":   count,
	})
}
