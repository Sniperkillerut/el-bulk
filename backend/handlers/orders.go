package handlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
)

type OrderHandler struct {
	DB *sqlx.DB
}

func NewOrderHandler(db *sqlx.DB) *OrderHandler {
	return &OrderHandler{DB: db}
}

// generateOrderNumber creates a unique order number like EB-20260324-A1B2
func generateOrderNumber() string {
	now := time.Now()
	b := make([]byte, 2)
	rand.Read(b)
	return fmt.Sprintf("EB-%s-%04X", now.Format("20060102"), int(b[0])<<8|int(b[1]))
}

// POST /api/orders — public checkout
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.CreateOrderInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if input.FirstName == "" || input.Phone == "" || input.PaymentMethod == "" {
		jsonError(w, "first_name, phone, and payment_method are required", http.StatusBadRequest)
		return
	}
	if len(input.Items) == 0 {
		jsonError(w, "At least one item is required", http.StatusBadRequest)
		return
	}

	// 1. Load exchange rates for price computation
	s, err := loadSettings(h.DB)
	if err != nil {
		s = models.Settings{USDToCOPRate: 4200, EURToCOPRate: 4600}
	}

	// 2. Fetch products and compute prices (snapshot at order time)
	productIDs := make([]string, 0)
	for _, item := range input.Items {
		if item.Quantity > 0 {
			productIDs = append(productIDs, item.ProductID)
		}
	}

	if len(productIDs) == 0 {
		jsonError(w, "No valid items selected", http.StatusBadRequest)
		return
	}

	productMap := make(map[string]models.Product)
	query, args, err := sqlx.In(`SELECT * FROM product WHERE id IN (?)`, productIDs)
	if err != nil {
		jsonError(w, "Internal query error", http.StatusInternalServerError)
		return
	}
	var products []models.Product
	if err := h.DB.Select(&products, h.DB.Rebind(query), args...); err != nil {
		jsonError(w, "Failed to fetch products", http.StatusInternalServerError)
		return
	}
	for _, p := range products {
		productMap[p.ID] = p
	}

	// 3. Fetch storage locations for snapshots
	storageMap := make(map[string][]models.StorageLocation)
	query, args, err = sqlx.In(`
		SELECT ps.product_id as product_id_temp, ps.storage_id, s.name, ps.quantity 
		FROM product_storage ps 
		JOIN storage_location s ON ps.storage_id = s.id 
		WHERE ps.product_id IN (?) AND ps.quantity > 0
	`, productIDs)
	if err == nil {
		type storageWithPID struct {
			models.StorageLocation
			ProductID string `db:"product_id_temp"`
		}
		var rows []storageWithPID
		if err := h.DB.Select(&rows, h.DB.Rebind(query), args...); err == nil {
			for _, row := range rows {
				storageMap[row.ProductID] = append(storageMap[row.ProductID], row.StorageLocation)
			}
		}
	}

	// 4. Prepare data for Stored Procedure
	var totalCOP float64
	var orderItems []map[string]interface{}
	for _, item := range input.Items {
		p, ok := productMap[item.ProductID]
		if !ok || item.Quantity <= 0 {
			continue
		}
		price := p.ComputePrice(s.USDToCOPRate, s.EURToCOPRate)
		totalCOP += price * float64(item.Quantity)

		orderItems = append(orderItems, map[string]interface{}{
			"product_id":         p.ID,
			"product_name":       p.Name,
			"product_set":        p.SetName,
			"foil_treatment":     p.FoilTreatment,
			"card_treatment":     p.CardTreatment,
			"condition":          p.Condition,
			"unit_price_cop":     price,
			"quantity":           item.Quantity,
			"stored_in_snapshot": storageMap[p.ID],
		})
	}

	var customerIDStr string
	if ctxID := r.Context().Value(middleware.UserIDKey); ctxID != nil {
		customerIDStr = ctxID.(string)
	}

	customerJSON, _ := json.Marshal(map[string]interface{}{
		"id":         customerIDStr,
		"first_name": input.FirstName,
		"last_name":  input.LastName,
		"email":      input.Email,
		"phone":      input.Phone,
		"id_number":  input.IDNumber,
		"address":    input.Address,
	})
	itemsJSON, _ := json.Marshal(orderItems)
	metaJSON, _ := json.Marshal(map[string]interface{}{
		"order_number":   generateOrderNumber(),
		"payment_method": input.PaymentMethod,
		"total_cop":      totalCOP,
		"notes":          input.Notes,
	})

	// 5. Execute Stored Procedure
	var result struct {
		OrderID     string `db:"order_id"`
		OrderNumber string `db:"order_number"`
	}
	err = h.DB.Get(&result, "SELECT order_id, order_number FROM fn_place_order($1, $2, $3)",
		string(customerJSON), string(itemsJSON), string(metaJSON))

	if err != nil {
		logger.Error("Place order SP failed: %v", err)
		jsonError(w, "Failed to place order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonOK(w, map[string]interface{}{
		"order_number": result.OrderNumber,
		"order_id":     result.OrderID,
		"total_cop":    totalCOP,
		"status":       "pending",
	})
}

// GET /api/admin/orders — list orders with pagination and filters
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	status := q.Get("status")
	search := q.Get("search")

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

	if status != "" {
		conditions = append(conditions, "o.status = $"+strconv.Itoa(idx))
		args = append(args, status)
		idx++
	}
	if search != "" {
		conditions = append(conditions, "(o.order_number ILIKE $"+strconv.Itoa(idx)+" OR o.customer_name ILIKE $"+strconv.Itoa(idx)+" OR o.customer_phone ILIKE $"+strconv.Itoa(idx)+" OR o.customer_email ILIKE $"+strconv.Itoa(idx)+")")
		args = append(args, "%"+search+"%")
		idx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count using the view for consistency
	var total int
	countQ := `SELECT COUNT(*) FROM view_order_list o ` + where
	if err := h.DB.Get(&total, countQ, args...); err != nil {
		logger.Error("Order count error: %v", err)
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Fetch
	listQ := fmt.Sprintf(`
		SELECT * FROM view_order_list o
		%s
		ORDER BY o.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, idx, idx+1)
	args = append(args, pageSize, offset)

	var orders []models.OrderWithCustomer
	if err := h.DB.Select(&orders, listQ, args...); err != nil {
		logger.Error("Order list error: %v", err)
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}
	if orders == nil {
		orders = []models.OrderWithCustomer{}
	}

	jsonOK(w, models.OrderListResponse{
		Orders:   orders,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GET /api/admin/orders/{id} — full order detail with product info
func (h *OrderHandler) GetDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Fetch order
	var order models.Order
	if err := h.DB.Get(&order, `SELECT * FROM "order" WHERE id = $1`, id); err != nil {
		if err.Error() == "sql: no rows in result set" {
			jsonError(w, "Order not found", http.StatusNotFound)
		} else {
			jsonError(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Fetch customer
	var customer models.Customer
	if err := h.DB.Get(&customer, `SELECT * FROM customer WHERE id = $1`, order.CustomerID); err != nil {
		jsonError(w, "Customer not found", http.StatusInternalServerError)
		return
	}

	// Fetch order items with enrichment from view
	var rows []struct {
		models.OrderItem
		ImageURL     *string `db:"image_url"`
		Stock        int     `db:"stock"`
		StoredInJSON []byte  `db:"stored_in"`
	}
	if err := h.DB.Select(&rows, `SELECT * FROM view_order_item_enriched WHERE order_id = $1 ORDER BY product_name`, id); err != nil {
		logger.Error("Error loading order items for %s: %v", id, err)
		jsonError(w, "Failed to load order items", http.StatusInternalServerError)
		return
	}

	// Assemble detail items
	detailItems := make([]models.OrderItemDetail, len(rows))
	for i, r := range rows {
		detailItems[i] = models.OrderItemDetail{
			OrderItem: r.OrderItem,
			ImageURL:  r.ImageURL,
			Stock:     r.Stock,
			StoredIn:  []models.StorageLocation{},
		}
		if r.StoredInJSON != nil {
			json.Unmarshal(r.StoredInJSON, &detailItems[i].StoredIn)
		}
	}

	jsonOK(w, models.OrderDetail{
		Order:    order,
		Customer: customer,
		Items:    detailItems,
	})
}

// PUT /api/admin/orders/{id} — update order (status, item quantities)
func (h *OrderHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var input struct {
		Status *string `json:"status"`
		Items  []struct {
			ID       string `json:"id"`
			Quantity int    `json:"quantity"`
		} `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Update status if provided
	if input.Status != nil {
		_, err = tx.Exec(`UPDATE "order" SET status = $1 WHERE id = $2`, *input.Status, id)
		if err != nil {
			jsonError(w, "Failed to update order status: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Update item quantities (allow 0 but don't delete)
	for _, item := range input.Items {
		if item.Quantity < 0 {
			item.Quantity = 0
		}

		// Verify stock limit
		var stock int
		err = tx.Get(&stock, `SELECT p.stock FROM product p JOIN order_item oi ON p.id = oi.product_id WHERE oi.id = $1`, item.ID)
		if err == nil && item.Quantity > stock {
			jsonError(w, fmt.Sprintf("Quantity %d exceeds available stock %d", item.Quantity, stock), http.StatusBadRequest)
			return
		}

		_, err = tx.Exec(`UPDATE order_item SET quantity = $1 WHERE id = $2 AND order_id = $3`,
			item.Quantity, item.ID, id)
		if err != nil {
			jsonError(w, "Failed to update item quantity", http.StatusInternalServerError)
			return
		}
	}

	// Recalculate total
	var newTotal float64
	err = tx.Get(&newTotal, `SELECT COALESCE(SUM(unit_price_cop * quantity), 0) FROM order_item WHERE order_id = $1`, id)
	if err == nil {
		tx.Exec(`UPDATE "order" SET total_cop = $1 WHERE id = $2`, newTotal, id)
	}

	if err := tx.Commit(); err != nil {
		jsonError(w, "Failed to save changes", http.StatusInternalServerError)
		return
	}

	h.GetDetail(w, r)
}

// POST /api/admin/orders/{id}/complete — mark order complete and decrement stock
func (h *OrderHandler) Complete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var input models.CompleteOrderInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Prepare decrements for SP
	jsonData, err := json.Marshal(input.Decrements)
	if err != nil {
		jsonError(w, "Failed to encode decrements", http.StatusInternalServerError)
		return
	}

	// Execute Stored Procedure
	if _, err := h.DB.Exec("SELECT fn_complete_order($1, $2)", id, string(jsonData)); err != nil {
		logger.Error("Complete order SP failed: %v", err)
		status := http.StatusInternalServerError
		errMsg := "Failed to complete order: " + err.Error()

		errStrLower := strings.ToLower(err.Error())
		if strings.Contains(errStrLower, "stock") {
			status = http.StatusBadRequest
		} else if strings.Contains(errStrLower, "already completed") {
			status = http.StatusBadRequest
			errMsg = "Order is already completed"
		}
		// Assuming utils.ErrorResponse is a helper function similar to jsonError
		// If not, replace with jsonError(w, errMsg, status)
		jsonError(w, errMsg, status)
		return
	}

	h.GetDetail(w, r)
}

// Helpers
func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func strPtr(s string) *string {
	return &s
}
