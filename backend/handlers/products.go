package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	maxPageSize := 1000
	if isAdmin {
		maxPageSize = 5000
	}
	page, pageSize, offset := httputil.GetPagination(r, 20, maxPageSize)

	params := store.ProductFilterParams{
		TCG:            q.Get("tcg"),
		Category:       q.Get("category"),
		Search:         q.Get("search"),
		StorageID:      q.Get("storage_id"),
		Foil:           q.Get("foil"),
		Treatment:      q.Get("treatment"),
		Condition:      q.Get("condition"),
		Collection:     q.Get("collection"),
		Rarity:         q.Get("rarity"),
		Language:       q.Get("language"),
		Color:          q.Get("color"),
		SetName:        q.Get("set_name"),
		InStock:        q.Get("in_stock") == "true",
		SortBy:         q.Get("sort_by"),
		SortDir:        q.Get("sort_dir"),
		OnlyDuplicates: q.Get("only_duplicates") == "true",
		FilterLogic:    q.Get("logic"),
		IsLegendary:    q.Get("is_legendary"),
		IsLand:         q.Get("is_land"),
		IsHistoric:     q.Get("is_historic"),
		FullArt:        q.Get("full_art"),
		Textless:       q.Get("textless"),
		IsBasicLand:    q.Get("is_basic_land"),
		IsCreature:     q.Get("is_creature"),
		IsSorcery:      q.Get("is_sorcery"),
		IsInstant:      q.Get("is_instant"),
		IsArtifact:     q.Get("is_artifact"),
		IsEnchantment:  q.Get("is_enchantment"),
		IsPlaneswalker: q.Get("is_planeswalker"),
		IsNonBasicLand: q.Get("is_non_basic_land"),
		Format:         q.Get("format"),
		Page:           page,
		PageSize:       pageSize,
		Offset:         offset,
	}

	if q.Get("ids") != "" {
		params.IDs = strings.Split(q.Get("ids"), ",")
	}

	resp, err := h.Service.List(r.Context(), params, isAdmin)
	if err != nil {
		logger.ErrorCtx(r.Context(), "List products failed: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	resp.QueryTimeMS = time.Since(start).Milliseconds()

	if !isAdmin {
		// Public catalog caching: 60s at the edge to absorb traffic spikes, but keep stock/prices relatively fresh
		w.Header().Set("Cache-Control", "public, s-maxage=60, stale-while-revalidate=120")
	} else {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	}

	render.Success(w, resp)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.TraceCtx(r.Context(), "Entering ProductHandler.GetByID | ID: %s", id)
	if id == "" {
		render.Error(w, "Product ID is required", http.StatusBadRequest)
		return
	}
	if err := httputil.ValidateUUID(id); err != nil {
		render.Error(w, "Invalid Product ID format", http.StatusBadRequest)
		return
	}
	isAdmin, _ := r.Context().Value(middleware.IsAdminKey).(bool)

	product, err := h.Service.GetByID(r.Context(), id, isAdmin)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Get product %s failed: %v", id, err)
		render.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	if !isAdmin {
		// Product detail caching: 60s at the edge
		w.Header().Set("Cache-Control", "public, s-maxage=60, stale-while-revalidate=120")
	} else {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
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
	// TCGs rarely change, safe to cache for an hour at the edge
	w.Header().Set("Cache-Control", "public, s-maxage=3600, stale-while-revalidate=7200")
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

	if err := input.Validate(); err != nil {
		render.Error(w, err.Error(), http.StatusBadRequest)
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
	if err := httputil.ValidateUUID(id); err != nil {
		render.Error(w, "Invalid Product ID format", http.StatusBadRequest)
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
	if err := httputil.ValidateUUID(id); err != nil {
		render.Error(w, "Invalid Product ID format", http.StatusBadRequest)
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

	var validProducts []models.ProductInput
	for i, p := range req.Products {
		if err := p.Validate(); err != nil {
			logger.WarnCtx(r.Context(), "Validation failed for product at index %d: %v. Skipping.", i, err)
			continue
		}
		validProducts = append(validProducts, p)
	}

	if len(validProducts) == 0 && len(req.Products) > 0 {
		render.Error(w, "No valid products provided", http.StatusBadRequest)
		return
	}

	count, err := h.Service.BulkCreate(r.Context(), validProducts, req.CategoryIDs)
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
	if err := httputil.ValidateUUID(id); err != nil {
		render.Error(w, "Invalid Product ID format", http.StatusBadRequest)
		return
	}
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
	if err := httputil.ValidateUUID(id); err != nil {
		render.Error(w, "Invalid Product ID format", http.StatusBadRequest)
		return
	}
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

	// Check if flusher is supported for real-time progress
	flusher, canFlush := w.(http.Flusher)
	if canFlush {
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)

		onProgress := func(current, total int) {
			progress := map[string]interface{}{
				"type":    "progress",
				"current": current,
				"total":   total,
			}
			data, _ := json.Marshal(progress)
			fmt.Fprintf(w, "%s\n", data)
			flusher.Flush()
		}

		count, err := h.Service.BulkUpdateSource(r.Context(), req.IDs, req.Source, onProgress)
		if err != nil {
			logger.ErrorCtx(r.Context(), "BulkUpdateSource streaming failed: %v", err)
			progress := map[string]interface{}{
				"type":  "error",
				"error": err.Error(),
			}
			data, _ := json.Marshal(progress)
			fmt.Fprintf(w, "%s\n", data)
			return
		}

		final := map[string]interface{}{
			"type":    "complete",
			"count":   count,
			"message": fmt.Sprintf("Successfully updated %d products", count),
		}
		data, _ := json.Marshal(final)
		fmt.Fprintf(w, "%s\n", data)
		return
	}

	// Fallback for non-streaming clients
	count, err := h.Service.BulkUpdateSource(r.Context(), req.IDs, req.Source, nil)
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

func (h *ProductHandler) BulkMoveStorage(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering ProductHandler.BulkMoveStorage")
	var req models.BulkMoveStorageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode bulk move storage input: %v", err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.Service.BulkMoveStorage(r.Context(), req); err != nil {
		logger.ErrorCtx(r.Context(), "BulkMoveStorage failed: %v", err)
		render.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.Success(w, map[string]string{"message": "Relocation complete"})
}
