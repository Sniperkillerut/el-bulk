package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/el-bulk/backend/utils/render"

	"github.com/go-chi/chi/v5"

	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/utils/httputil"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/sqlutil"
)

type OrderHandler struct {
	Service *service.OrderService
}

func NewOrderHandler(s *service.OrderService) *OrderHandler {
	return &OrderHandler{Service: s}
}



// POST /api/orders — public checkout
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering OrderHandler.Create")
	var input models.CreateOrderInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode order input: %v", err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if input.FirstName == "" || input.Phone == "" || input.PaymentMethod == "" {
		render.Error(w, "first_name, phone, and payment_method are required", http.StatusBadRequest)
		return
	}
	if len(input.Items) == 0 {
		render.Error(w, "At least one item is required", http.StatusBadRequest)
		return
	}

	var customerID string
	if ctxID := r.Context().Value(middleware.UserIDKey); ctxID != nil {
		customerID = ctxID.(string)
	}

	orderID, orderNumber, totalCOP, err := h.Service.CreateOrder(r.Context(), input, customerID)
	if err != nil {
		logger.ErrorCtx(r.Context(), "CreateOrder failed: %v", err)
		render.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.Success(w, map[string]interface{}{
		"order_number": orderNumber,
		"order_id":     orderID,
		"total_cop":    totalCOP,
		"status":       "pending",
	})
}

// GET /api/admin/orders — list orders with pagination and filters
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering OrderHandler.List | URL: %s", r.URL.String())
	q := r.URL.Query()
	status := q.Get("status")
	search := q.Get("search")

	page, pageSize, _ := httputil.GetPagination(r, 20, 100)

	builder := sqlutil.NewBuilder("")
	if status != "" {
		builder.AddCondition("o.status = ?", status)
	}
	if search != "" {
		sPattern := "%" + search + "%"
		builder.AddCondition("(o.order_number ILIKE ? OR o.customer_name ILIKE ? OR o.customer_phone ILIKE ? OR o.customer_email ILIKE ?)", sPattern)
	}

	whereClause, args := builder.Build()

	orders, total, err := h.Service.ListOrders(r.Context(), whereClause, args, page, pageSize)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Order list error: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	render.Success(w, models.OrderListResponse{
		Orders:   orders,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GET /api/admin/orders/{id} — full order detail with product info
func (h *OrderHandler) GetDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.TraceCtx(r.Context(), "Entering OrderHandler.GetDetail | ID: %s", id)

	detail, err := h.Service.GetOrderDetail(r.Context(), id, true)
	if err != nil {
		logger.ErrorCtx(r.Context(), "GetOrderDetail failed for %s: %v", id, err)
		render.Error(w, fmt.Sprintf("Order detail error: %v", err), http.StatusInternalServerError)
		return
	}

	render.Success(w, detail)
}




// PUT /api/admin/orders/{id} — update order (status, item quantities)
func (h *OrderHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.TraceCtx(r.Context(), "Entering OrderHandler.Update | ID: %s", id)

	var input models.UpdateOrderInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode update input for %s: %v", id, err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.Service.UpdateOrder(r.Context(), id, input); err != nil {
		logger.ErrorCtx(r.Context(), "UpdateOrder failed for %s: %v", id, err)
		render.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.GetDetail(w, r)
}

// POST /api/admin/orders/{id}/confirm — mark order confirmed and decrement stock
func (h *OrderHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.TraceCtx(r.Context(), "Entering OrderHandler.Confirm | ID: %s", id)

	var input models.ConfirmOrderInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode confirm input for %s: %v", id, err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.Service.ConfirmOrder(r.Context(), id, input.Decrements); err != nil {
		logger.ErrorCtx(r.Context(), "Confirm order failed: %v", err)
		status := http.StatusInternalServerError
		errMsg := "Failed to confirm order: " + err.Error()

		errStrLower := strings.ToLower(err.Error())
		if strings.Contains(errStrLower, "stock") {
			status = http.StatusBadRequest
		} else if strings.Contains(errStrLower, "already processed") {
			status = http.StatusBadRequest
			errMsg = "Order is already processed"
		}
		render.Error(w, errMsg, status)
		return
	}

	h.GetDetail(w, r)
}

// POST /api/admin/orders/{id}/restore — manually restore stock for a cancelled order
func (h *OrderHandler) RestoreStock(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.TraceCtx(r.Context(), "Entering OrderHandler.RestoreStock | ID: %s", id)

	var input struct {
		Increments []models.StockDecrement `json:"increments"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to decode restore input for %s: %v", id, err)
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.Service.RestoreStock(r.Context(), id, input.Increments); err != nil {
		logger.ErrorCtx(r.Context(), "Restore stock failed: %v", err)
		render.Error(w, "Failed to restore stock: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.GetDetail(w, r)
}

// GET /api/orders/me — list orders for the current customer
func (h *OrderHandler) ListMe(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering OrderHandler.ListMe")
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		render.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	orders, err := h.Service.ListMe(r.Context(), userID)
	if err != nil {
		logger.ErrorCtx(r.Context(), "User order list error for %s: %v", userID, err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	render.Success(w, orders)
}

// GET /api/orders/me/{id} — get a single order for the current user
func (h *OrderHandler) GetMeDetail(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering OrderHandler.GetMeDetail")
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		render.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	detail, err := h.Service.GetOrderDetail(r.Context(), id, false)
	if err != nil {
		render.Error(w, "Order not found or database error", http.StatusNotFound)
		return
	}

	// Verify ownership
	if detail.Order.CustomerID != userID {
		render.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	render.Success(w, detail)
}

// POST /api/orders/me/{id}/cancel — cancel a pending order for the current user
func (h *OrderHandler) CancelMe(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering OrderHandler.CancelMe")
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		render.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")

	if err := h.Service.CancelMe(r.Context(), id, userID); err != nil {
		logger.ErrorCtx(r.Context(), "User order cancel error for %s (userID: %s): %v", id, userID, err)
		render.Error(w, "Order cannot be cancelled. It may not exist, belong to you, or is already being processed.", http.StatusBadRequest)
		return
	}

	h.GetMeDetail(w, r)
}
