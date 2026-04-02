package handlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/el-bulk/backend/utils/render"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/crypto"
	"github.com/el-bulk/backend/utils/httputil"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/sqlutil"
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
		render.Error(w, "No valid items selected", http.StatusBadRequest)
		return
	}

	productMap := make(map[string]models.Product)
	query, args, err := sqlx.In(`SELECT * FROM product WHERE id IN (?)`, productIDs)
	if err != nil {
		render.Error(w, "Internal query error", http.StatusInternalServerError)
		return
	}
	var products []models.Product
	if err := h.DB.Select(&products, h.DB.Rebind(query), args...); err != nil {
		render.Error(w, "Failed to fetch products", http.StatusInternalServerError)
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
	totalCOP := 0.0
	orderItems := make([]map[string]interface{}, 0)
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

	if len(orderItems) == 0 {
		render.Error(w, "No valid items selected (they might be missing from inventory, please clear the cart and try again)", http.StatusBadRequest)
		return
	}

	var customerIDStr string
	if ctxID := r.Context().Value(middleware.UserIDKey); ctxID != nil {
		customerIDStr = ctxID.(string)
	}

	// Encrypt guest PII if provided
	encPhone, _ := crypto.Encrypt(input.Phone)
	encIDNumber, _ := crypto.Encrypt(input.IDNumber)
	encAddress, _ := crypto.Encrypt(input.Address)

	customerJSON, _ := json.Marshal(map[string]interface{}{
		"id":         customerIDStr,
		"first_name": input.FirstName,
		"last_name":  input.LastName,
		"email":      input.Email,
		"phone":      encPhone,
		"id_number":  encIDNumber,
		"address":    encAddress,
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
		render.Error(w, "Failed to place order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.Success(w, map[string]interface{}{
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

	page, pageSize, offset := httputil.GetPagination(r, 20, 100)

	builder := sqlutil.NewBuilder("FROM view_order_list o")
	if status != "" {
		builder.AddCondition("o.status = ?", status)
	}
	if search != "" {
		sPattern := "%" + search + "%"
		builder.AddCondition("(o.order_number ILIKE ? OR o.customer_name ILIKE ? OR o.customer_phone ILIKE ? OR o.customer_email ILIKE ?)", sPattern)
	}

	whereClause, args := builder.Build()

	// Count using the view
	var total int
	countQ := `SELECT COUNT(*) ` + whereClause
	if err := h.DB.Get(&total, countQ, args...); err != nil {
		logger.Error("Order count error: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Fetch
	listQ := fmt.Sprintf(`
		SELECT * 
		%s
		ORDER BY o.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, len(args)+1, len(args)+2)
	args = append(args, pageSize, offset)

	var orders []models.OrderWithCustomer
	if err := h.DB.Select(&orders, listQ, args...); err != nil {
		logger.Error("Order list error: %v", err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if orders == nil {
		orders = []models.OrderWithCustomer{}
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

	// Fetch order
	var order models.Order
	if err := h.DB.Get(&order, `SELECT * FROM "order" WHERE id = $1`, id); err != nil {
		if err.Error() == "sql: no rows in result set" {
			render.Error(w, "Order not found", http.StatusNotFound)
		} else {
			render.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Fetch customer
	var customer models.Customer
	if err := h.DB.Get(&customer, `SELECT * FROM customer WHERE id = $1`, order.CustomerID); err != nil {
		render.Error(w, "Customer not found", http.StatusInternalServerError)
		return
	}

	// Decrypt sensitive fields
	customer.Phone = crypto.DecryptSafe(customer.Phone)
	customer.IDNumber = crypto.DecryptSafe(customer.IDNumber)
	customer.Address = crypto.DecryptSafe(customer.Address)

	// Fetch order items with enrichment from view
	var rows []struct {
		models.OrderItem
		ImageURL     *string `db:"image_url"`
		Stock        int     `db:"stock"`
		StoredInJSON []byte  `db:"stored_in"`
	}
	if err := h.DB.Select(&rows, `SELECT * FROM view_order_item_enriched WHERE order_id = $1 ORDER BY product_name`, id); err != nil {
		logger.Error("Error loading order items for %s: %v", id, err)
		render.Error(w, "Failed to load order items", http.StatusInternalServerError)
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

	render.Success(w, models.OrderDetail{
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
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Update status if provided
	if input.Status != nil {
		_, err = tx.Exec(`UPDATE "order" SET status = $1 WHERE id = $2`, *input.Status, id)
		if err != nil {
			render.Error(w, "Failed to update order status: "+err.Error(), http.StatusInternalServerError)
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
			render.Error(w, fmt.Sprintf("Quantity %d exceeds available stock %d", item.Quantity, stock), http.StatusBadRequest)
			return
		}

		_, err = tx.Exec(`UPDATE order_item SET quantity = $1 WHERE id = $2 AND order_id = $3`,
			item.Quantity, item.ID, id)
		if err != nil {
			render.Error(w, "Failed to update item quantity", http.StatusInternalServerError)
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
		render.Error(w, "Failed to save changes", http.StatusInternalServerError)
		return
	}

	h.GetDetail(w, r)
}

// POST /api/admin/orders/{id}/complete — mark order complete and decrement stock
func (h *OrderHandler) Complete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var input models.CompleteOrderInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Prepare decrements for SP
	decrements := input.Decrements
	if decrements == nil {
		decrements = []models.StockDecrement{}
	}
	jsonData, err := json.Marshal(decrements)
	if err != nil {
		render.Error(w, "Failed to encode decrements", http.StatusInternalServerError)
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
		render.Error(w, errMsg, status)
		return
	}

	h.GetDetail(w, r)
}

// GET /api/orders/me — list orders for the current customer
func (h *OrderHandler) ListMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		render.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	var orders []struct {
		models.Order
		ItemCount int `db:"item_count" json:"item_count"`
	}

	err := h.DB.Select(&orders, `
		SELECT o.*, (SELECT SUM(quantity) FROM order_item WHERE order_id = o.id) as item_count
		FROM "order" o 
		WHERE o.customer_id = $1 
		ORDER BY o.created_at DESC
	`, userID)

	if err != nil {
		logger.Error("User order list error for %s: %v", userID, err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if orders == nil {
		orders = make([]struct {
			models.Order
			ItemCount int `db:"item_count" json:"item_count"`
		}, 0)
	}

	render.Success(w, orders)
}

// GET /api/orders/me/{id} — get a single order for the current user
func (h *OrderHandler) GetMeDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		render.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")

	// Fetch order - ensuring it belongs to this user
	var order models.Order
	if err := h.DB.Get(&order, `SELECT * FROM "order" WHERE id = $1 AND customer_id = $2`, id, userID); err != nil {
		if err.Error() == "sql: no rows in result set" {
			render.Error(w, "Order not found", http.StatusNotFound)
		} else {
			logger.Error("User order detail error for %s (userID: %s): %v", id, userID, err)
			render.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Fetch customer (just for completeness, user already has their own info)
	var customer models.Customer
	if err := h.DB.Get(&customer, `SELECT * FROM customer WHERE id = $1`, order.CustomerID); err != nil {
		render.Error(w, "Customer not found", http.StatusInternalServerError)
		return
	}

	// Decrypt sensitive fields
	customer.Phone = crypto.DecryptSafe(customer.Phone)
	customer.IDNumber = crypto.DecryptSafe(customer.IDNumber)
	customer.Address = crypto.DecryptSafe(customer.Address)

	// Fetch order items with enrichment from view
	var rows []struct {
		models.OrderItem
		ImageURL     *string `db:"image_url"`
		Stock        int     `db:"stock"`
		StoredInJSON []byte  `db:"stored_in"`
	}
	if err := h.DB.Select(&rows, `SELECT * FROM view_order_item_enriched WHERE order_id = $1 ORDER BY product_name`, id); err != nil {
		logger.Error("Error loading order items for %s: %v", id, err)
		render.Error(w, "Failed to load order items", http.StatusInternalServerError)
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
		// StoredIn is internal, but we populate it for models consistency even if frontend might ignore it
		if r.StoredInJSON != nil {
			json.Unmarshal(r.StoredInJSON, &detailItems[i].StoredIn)
		}
	}

	render.Success(w, models.OrderDetail{
		Order:    order,
		Customer: customer,
		Items:    detailItems,
	})
}
