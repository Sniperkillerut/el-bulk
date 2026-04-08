package handlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
		WHERE ps.product_id IN (?) AND ps.quantity > 0 AND s.name != 'pending'
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
	subtotalCOP := 0.0
	orderItems := make([]map[string]interface{}, 0)
	for _, item := range input.Items {
		p, ok := productMap[item.ProductID]
		if !ok || item.Quantity <= 0 {
			continue
		}
		price := p.ComputePrice(s.USDToCOPRate, s.EURToCOPRate)
		subtotalCOP += price * float64(item.Quantity)

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

	shippingCOP := 0.0
	if !input.IsLocalPickup {
		shippingCOP = s.FlatShippingFeeCOP
	}
	taxCOP := 0.0 // Could be calculated here if needed
	totalCOP := subtotalCOP + shippingCOP + taxCOP

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
		"subtotal_cop":   subtotalCOP,
		"shipping_cop":   shippingCOP,
		"tax_cop":        taxCOP,
		"total_cop":      totalCOP,
		"is_local_pickup": input.IsLocalPickup,
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
		render.Error(w, "Failed to place order. Please try again.", http.StatusInternalServerError)

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
		Order:       order,
		Customer:    customer,
		Items:       detailItems,
		WhatsAppURL: h.generateWhatsAppURL(order, customer),
	})
}

func (h *OrderHandler) generateWhatsAppURL(order models.Order, customer models.Customer) string {
	phone := ""
	if customer.Phone != nil {
		p := *customer.Phone
		for _, char := range p {
			if char >= '0' && char <= '9' {
				phone += string(char)
			}
		}
	}
	if phone == "" {
		return ""
	}
	// Default to Colombia (57) if no country code provided and length is 10
	if len(phone) == 10 && !strings.HasPrefix(phone, "57") {
		phone = "57" + phone
	}

	msg := ""
	switch order.Status {
	case "ready_for_pickup":
		msg = fmt.Sprintf("¡Hola %s! 👋 Tu pedido %s en El Bulk ya está listo para ser reclamado en nuestra tienda ⚓. ¡Te esperamos!",
			customer.FirstName, order.OrderNumber)
	case "shipped":
		tracking := ""
		if order.TrackingNumber != nil {
			tracking = "con guía " + *order.TrackingNumber
		}
		msg = fmt.Sprintf("¡Hola %s! 📦 Tu pedido %s en El Bulk ya ha sido enviado %s. ¡Pronto estará contigo!",
			customer.FirstName, order.OrderNumber, tracking)
	default:
		msg = fmt.Sprintf("¡Hola %s! Referente a tu pedido %s en El Bulk...", customer.FirstName, order.OrderNumber)
	}

	return fmt.Sprintf("https://wa.me/%s?text=%s", phone, url.QueryEscape(msg))
}


// PUT /api/admin/orders/{id} — update order (status, item quantities)
func (h *OrderHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var input struct {
		Status         *string `json:"status"`
		TrackingNumber *string `json:"tracking_number"`
		TrackingURL    *string `json:"tracking_url"`
		Items          []struct {
			ID       string `json:"id"`
			Quantity int    `json:"quantity"`
		} `json:"items"`
		AddedItems []struct {
			ProductID    string  `json:"product_id"`
			Quantity     int     `json:"quantity"`
			UnitPriceCOP float64 `json:"unit_price_cop"`
		} `json:"added_items"`
		DeletedIDs []string `json:"deleted_ids"`
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

	// Update status and tracking if provided
	if input.Status != nil || input.TrackingNumber != nil || input.TrackingURL != nil {
		if input.Status != nil && *input.Status == "pending" {
			var currentStatus string
			err = tx.Get(&currentStatus, `SELECT status FROM "order" WHERE id = $1`, id)
			if err == nil && (currentStatus == "confirmed" || currentStatus == "completed") {
				render.Error(w, "Cannot move a confirmed/completed order back to pending", http.StatusBadRequest)
				return
			}
		}

		_, err = tx.Exec(`
			UPDATE "order" 
			SET status = COALESCE($1, status),
			    tracking_number = COALESCE($2, tracking_number),
			    tracking_url = COALESCE($3, tracking_url)
			WHERE id = $4`, 
			input.Status, input.TrackingNumber, input.TrackingURL, id)
		if err != nil {
			render.Error(w, "Failed to update order info: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Update existing item quantities (allow 0 but don't delete)
	// Also manage pending storage adjustments
	for _, item := range input.Items {
		if item.Quantity < 0 {
			item.Quantity = 0
		}

		var current struct {
			Quantity  int    `db:"quantity"`
			ProductID string `db:"product_id"`
			Stock     int    `db:"stock"`
		}
		err = tx.Get(&current, `
			SELECT oi.quantity, oi.product_id, p.stock 
			FROM order_item oi 
			JOIN product p ON p.id = oi.product_id 
			WHERE oi.id = $1 AND oi.order_id = $2
		`, item.ID, id)
		if err != nil {
			continue // gracefully skip invalid items
		}

		delta := item.Quantity - current.Quantity
		if delta == 0 {
			continue
		}

		// If increasing, verify we have enough logical storefront stock to take from
		if delta > 0 && delta > current.Stock {
			render.Error(w, fmt.Sprintf("Cannot increase order by %d; only %d in stock", delta, current.Stock), http.StatusBadRequest)
			return
		}

		// 1. Update the order_item details
		_, err = tx.Exec(`UPDATE order_item SET quantity = $1 WHERE id = $2 AND order_id = $3`,
			item.Quantity, item.ID, id)
		if err != nil {
			render.Error(w, "Failed to update item quantity", http.StatusInternalServerError)
			return
		}

		// 2. Adjust the pending storage box to reflect the new assigned reserved quantity
		_, err = tx.Exec(`
			UPDATE product_storage 
			SET quantity = quantity + $1 
			WHERE product_id = $2 
			  AND storage_id = (SELECT id FROM storage_location WHERE name = 'pending' LIMIT 1)
		`, delta, current.ProductID)
		if err != nil {
			logger.Error("Failed to update pending storage holding: %v", err)
			render.Error(w, "Failed to update pending storage holding", http.StatusInternalServerError)
			return
		}
	}

	// Add new items
	for _, item := range input.AddedItems {
		if item.Quantity <= 0 {
			continue
		}

		// Fetch product details for denormalized insertion
		var product struct {
			Name          string  `db:"name"`
			SetName       *string `db:"set_name"`
			FoilTreatment string  `db:"foil_treatment"`
			CardTreatment string  `db:"card_treatment"`
			Condition     *string `db:"condition"`
			Stock         int     `db:"stock"`
		}
		err = tx.Get(&product, `SELECT name, set_name, foil_treatment, card_treatment, condition, stock FROM product WHERE id = $1`, item.ProductID)
		if err != nil {
			render.Error(w, "Product not found while adding to order", http.StatusBadRequest)
			return
		}

		if item.Quantity > product.Stock {
			render.Error(w, fmt.Sprintf("Added quantity %d exceeds available stock %d", item.Quantity, product.Stock), http.StatusBadRequest)
			return
		}

		// 1. Insert into order_item with denormalized product info
		_, err = tx.Exec(`
			INSERT INTO order_item (order_id, product_id, product_name, product_set, foil_treatment, card_treatment, condition, quantity, unit_price_cop)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, id, item.ProductID, product.Name, product.SetName, product.FoilTreatment, product.CardTreatment, product.Condition, item.Quantity, item.UnitPriceCOP)
		if err != nil {
			render.Error(w, "Failed to add new item to order: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 2. Occupy pending storage
		_, err = tx.Exec(`
			INSERT INTO product_storage (product_id, storage_id, quantity)
			VALUES ($1, (SELECT id FROM storage_location WHERE name = 'pending' LIMIT 1), $2)
			ON CONFLICT (product_id, storage_id) DO UPDATE SET quantity = product_storage.quantity + EXCLUDED.quantity
		`, item.ProductID, item.Quantity)
		if err != nil {
			render.Error(w, "Failed to update pending storage for added item", http.StatusInternalServerError)
			return
		}
	}

	// Delete items
	for _, itemID := range input.DeletedIDs {
		var current struct {
			Quantity  int    `db:"quantity"`
			ProductID string `db:"product_id"`
		}
		err = tx.Get(&current, `SELECT quantity, product_id FROM order_item WHERE id = $1 AND order_id = $2`, itemID, id)
		if err != nil {
			continue // skip items not found in this order
		}

		// 1. Remove from order_item
		_, err = tx.Exec(`DELETE FROM order_item WHERE id = $1 AND order_id = $2`, itemID, id)
		if err != nil {
			render.Error(w, "Failed to delete item from order", http.StatusInternalServerError)
			return
		}

		// 2. Release pending stock
		_, err = tx.Exec(`
			UPDATE product_storage 
			SET quantity = GREATEST(0, quantity - $1) 
			WHERE product_id = $2 
			  AND storage_id = (SELECT id FROM storage_location WHERE name = 'pending' LIMIT 1)
		`, current.Quantity, current.ProductID)
		if err != nil {
			render.Error(w, "Failed to release pending stock for deleted item", http.StatusInternalServerError)
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

// POST /api/admin/orders/{id}/confirm — mark order confirmed and decrement stock
func (h *OrderHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var input models.ConfirmOrderInput
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
	if _, err := h.DB.Exec("SELECT fn_confirm_order($1::uuid, $2::jsonb)", id, string(jsonData)); err != nil {
		logger.Error("Confirm order SP failed: %v", err)
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

	var input struct {
		Increments []models.StockDecrement `json:"increments"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		render.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	jsonData, err := json.Marshal(input.Increments)
	if err != nil {
		render.Error(w, "Failed to encode increments", http.StatusInternalServerError)
		return
	}

	if _, err := h.DB.Exec("SELECT fn_restore_order_stock($1::uuid, $2::jsonb)", id, string(jsonData)); err != nil {
		logger.Error("Restore order stock SP failed: %v", err)
		render.Error(w, "Failed to restore stock: "+err.Error(), http.StatusInternalServerError)
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

// POST /api/orders/me/{id}/cancel — cancel a pending order for the current user
func (h *OrderHandler) CancelMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		render.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")

	// Execute update ensuring ownership and correct status
	res, err := h.DB.Exec(`
		UPDATE "order" 
		SET status = 'cancelled' 
		WHERE id = $1 AND customer_id = $2 AND status = 'pending'
	`, id, userID)

	if err != nil {
		logger.Error("User order cancel error for %s (userID: %s): %v", id, userID, err)
		render.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		// Either order not found, not pending, or not owned by the user
		render.Error(w, "Order cannot be cancelled. It may not exist, belong to you, or is already being processed.", http.StatusBadRequest)
		return
	}

	// Return updated detail
	h.GetMeDetail(w, r)
}
