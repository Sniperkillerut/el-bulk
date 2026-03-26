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

	tx, err := h.DB.Beginx()
	if err != nil {
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// 1. Upsert customer by phone
	var customerID string
	email := nullStr(input.Email)
	idNumber := nullStr(input.IDNumber)
	address := nullStr(input.Address)

	err = tx.QueryRow(`
		INSERT INTO customers (first_name, last_name, email, phone, id_number, address)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING
		RETURNING id
	`, input.FirstName, input.LastName, email, input.Phone, idNumber, address).Scan(&customerID)

	if err != nil {
		// Customer might already exist, find by phone
		err = tx.QueryRow(`SELECT id FROM customers WHERE phone = $1`, input.Phone).Scan(&customerID)
		if err != nil {
			jsonError(w, "Failed to create/find customer: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Update customer info
		tx.Exec(`UPDATE customers SET first_name=$1, last_name=$2, email=$3, id_number=$4, address=$5 WHERE id=$6`,
			input.FirstName, input.LastName, email, idNumber, address, customerID)
	}

	// 2. Load exchange rates for price computation
	s, err := loadSettings(h.DB)
	if err != nil {
		s = models.Settings{USDToCOPRate: 4200, EURToCOPRate: 4600}
	}

	// 3. Fetch products and compute prices (snapshot at order time)
	var totalCOP float64
	type itemWithPrice struct {
		models.CreateOrderItem
		product models.Product
		price   float64
	}
	var enrichedItems []itemWithPrice

	// 3. Enrich items and calculate total in bulk
	productIDs := make([]string, 0)
	for _, item := range input.Items {
		if item.Quantity > 0 {
			productIDs = append(productIDs, item.ProductID)
		}
	}

	productMap := make(map[string]models.Product)
	if len(productIDs) > 0 {
		query, args, err := sqlx.In(`SELECT * FROM products WHERE id IN (?)`, productIDs)
		if err != nil {
			jsonError(w, "Query preparation error", http.StatusInternalServerError)
			return
		}
		var products []models.Product
		if err := tx.Select(&products, tx.Rebind(query), args...); err != nil {
			jsonError(w, "Failed to fetch products: "+err.Error(), http.StatusInternalServerError)
			return
		}
		for _, p := range products {
			productMap[p.ID] = p
		}
	}

	for _, item := range input.Items {
		if item.Quantity <= 0 {
			continue
		}
		product, ok := productMap[item.ProductID]
		if !ok {
			jsonError(w, fmt.Sprintf("Product %s not found", item.ProductID), http.StatusBadRequest)
			return
		}
		price := product.ComputePrice(s.USDToCOPRate, s.EURToCOPRate)
		totalCOP += price * float64(item.Quantity)
		enrichedItems = append(enrichedItems, itemWithPrice{
			CreateOrderItem: item,
			product:         product,
			price:           price,
		})
	}

	// 4. Create order
	orderNumber := generateOrderNumber()
	notes := nullStr(input.Notes)
	var order models.Order
	err = tx.QueryRowx(`
		INSERT INTO orders (order_number, customer_id, status, payment_method, total_cop, notes)
		VALUES ($1, $2, 'pending', $3, $4, $5)
		RETURNING *
	`, orderNumber, customerID, input.PaymentMethod, totalCOP, notes).StructScan(&order)
	if err != nil {
		jsonError(w, "Failed to create order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Create order items with snapshotted prices and current storage info
	// Bulk fetch all storage locations for snapshots
	storageMap := make(map[string][]models.StorageLocation)
	if len(productIDs) > 0 {
		query, args, err := sqlx.In(`
			SELECT ps.product_id as product_id_temp, ps.stored_in_id, s.name, ps.quantity 
			FROM product_stored_in ps 
			JOIN stored_in s ON ps.stored_in_id = s.id 
			WHERE ps.product_id IN (?) AND ps.quantity > 0
		`, productIDs)
		if err == nil {
			type storageWithPID struct {
				models.StorageLocation
				ProductID string `db:"product_id_temp"`
			}
			var rows []storageWithPID
			if err := tx.Select(&rows, tx.Rebind(query), args...); err == nil {
				for _, row := range rows {
					storageMap[row.ProductID] = append(storageMap[row.ProductID], row.StorageLocation)
				}
			}
		}
	}

	for _, ei := range enrichedItems {
		// Get current storage snapshot
		var storageJSON *string
		if storageRows, ok := storageMap[ei.ProductID]; ok {
			b, _ := json.Marshal(storageRows)
			str := string(b)
			storageJSON = &str
		}

		_, err = tx.Exec(`
			INSERT INTO order_items (order_id, product_id, product_name, product_set, 
				foil_treatment, card_treatment, condition, unit_price_cop, quantity, stored_in_snapshot)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, order.ID, ei.product.ID, ei.product.Name, ei.product.SetName,
			strPtr(string(ei.product.FoilTreatment)), strPtr(string(ei.product.CardTreatment)),
			ei.product.Condition, ei.price, ei.Quantity, storageJSON)

		if err != nil {
			jsonError(w, "Failed to create order item: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		jsonError(w, "Failed to finalize order", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonOK(w, map[string]interface{}{
		"order_number": order.OrderNumber,
		"order_id":     order.ID,
		"total_cop":    order.TotalCOP,
		"status":       order.Status,
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
		conditions = append(conditions, "(o.order_number ILIKE $"+strconv.Itoa(idx)+" OR c.first_name ILIKE $"+strconv.Itoa(idx)+" OR c.last_name ILIKE $"+strconv.Itoa(idx)+" OR c.phone ILIKE $"+strconv.Itoa(idx)+")")
		args = append(args, "%"+search+"%")
		idx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count
	var total int
	countQ := `SELECT COUNT(*) FROM orders o JOIN customers c ON o.customer_id = c.id ` + where
	if err := h.DB.Get(&total, countQ, args...); err != nil {
		logger.Error("Order count error: %v", err)
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Fetch
	listQ := fmt.Sprintf(`
		SELECT o.*, 
			c.first_name || ' ' || c.last_name as customer_name,
			COALESCE((SELECT COUNT(*) FROM order_items WHERE order_id = o.id), 0) as item_count
		FROM orders o
		JOIN customers c ON o.customer_id = c.id
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
	if err := h.DB.Get(&order, `SELECT * FROM orders WHERE id = $1`, id); err != nil {
		if err.Error() == "sql: no rows in result set" {
			jsonError(w, "Order not found", http.StatusNotFound)
		} else {
			jsonError(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Fetch customer
	var customer models.Customer
	if err := h.DB.Get(&customer, `SELECT * FROM customers WHERE id = $1`, order.CustomerID); err != nil {
		jsonError(w, "Customer not found", http.StatusInternalServerError)
		return
	}

	// Fetch order items with explicit columns (stored_in_snapshot is JSONB, cast to text)
	var items []models.OrderItem
	if err := h.DB.Select(&items, `
		SELECT id, order_id, product_id, product_name, product_set, 
		       foil_treatment, card_treatment, condition, 
		       unit_price_cop, quantity, 
		       stored_in_snapshot::text as stored_in_snapshot
		FROM order_items WHERE order_id = $1 ORDER BY product_name
	`, id); err != nil {
		logger.Error("Error loading order items for %s: %v", id, err)
		jsonError(w, "Failed to load order items", http.StatusInternalServerError)
		return
	}

	// Enrichment: Fetch all related products and storage locations in bulk
	productIDs := make([]string, 0)
	idMap := make(map[string]bool)
	for _, item := range items {
		if item.ProductID != nil {
			if !idMap[*item.ProductID] {
				productIDs = append(productIDs, *item.ProductID)
				idMap[*item.ProductID] = true
			}
		}
	}

	productMap := make(map[string]models.Product)
	storageMap := make(map[string][]models.StorageLocation)

	if len(productIDs) > 0 {
		// Bulk fetch products
		query, args, err := sqlx.In(`SELECT * FROM products WHERE id IN (?)`, productIDs)
		if err == nil {
			var products []models.Product
			if err := h.DB.Select(&products, h.DB.Rebind(query), args...); err == nil {
				for _, p := range products {
					productMap[p.ID] = p
				}
			}
		}

		// Bulk fetch storage locations
		query, args, err = sqlx.In(`
			SELECT ps.product_id as product_id_temp, ps.stored_in_id, s.name, ps.quantity
			FROM product_stored_in ps
			JOIN stored_in s ON ps.stored_in_id = s.id
			WHERE ps.product_id IN (?) AND ps.quantity > 0
		`, productIDs)
		if err == nil {
			// We need a temporary struct to capture product_id from the join
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
	}

	// Assemble detail items
	var detailItems []models.OrderItemDetail
	for _, item := range items {
		detail := models.OrderItemDetail{
			OrderItem: item,
			StoredIn:  []models.StorageLocation{},
		}

		if item.ProductID != nil {
			if p, ok := productMap[*item.ProductID]; ok {
				detail.Stock = p.Stock
				detail.ImageURL = p.ImageURL
			}
			if locs, ok := storageMap[*item.ProductID]; ok {
				detail.StoredIn = locs
			}
		}
		detailItems = append(detailItems, detail)
	}
	if detailItems == nil {
		detailItems = []models.OrderItemDetail{}
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
		_, err = tx.Exec(`UPDATE orders SET status = $1 WHERE id = $2`, *input.Status, id)
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
		err = tx.Get(&stock, `SELECT p.stock FROM products p JOIN order_items oi ON p.id = oi.product_id WHERE oi.id = $1`, item.ID)
		if err == nil && item.Quantity > stock {
			jsonError(w, fmt.Sprintf("Quantity %d exceeds available stock %d", item.Quantity, stock), http.StatusBadRequest)
			return
		}

		_, err = tx.Exec(`UPDATE order_items SET quantity = $1 WHERE id = $2 AND order_id = $3`,
			item.Quantity, item.ID, id)
		if err != nil {
			jsonError(w, "Failed to update item quantity", http.StatusInternalServerError)
			return
		}
	}

	// Recalculate total
	var newTotal float64
	err = tx.Get(&newTotal, `SELECT COALESCE(SUM(unit_price_cop * quantity), 0) FROM order_items WHERE order_id = $1`, id)
	if err == nil {
		tx.Exec(`UPDATE orders SET total_cop = $1 WHERE id = $2`, newTotal, id)
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

	// Verify order exists and is not already completed
	var order models.Order
	if err := h.DB.Get(&order, `SELECT * FROM orders WHERE id = $1`, id); err != nil {
		jsonError(w, "Order not found", http.StatusNotFound)
		return
	}
	if order.Status == "completed" {
		jsonError(w, "Order is already completed", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Decrement stock from specified storage locations
	for _, dec := range input.Decrements {
		if dec.Quantity <= 0 {
			continue
		}

		// Verify sufficient stock at that location
		var available int
		err := tx.Get(&available, `
			SELECT COALESCE(quantity, 0) FROM product_stored_in 
			WHERE product_id = $1 AND stored_in_id = $2
		`, dec.ProductID, dec.StoredInID)
		if err != nil || available < dec.Quantity {
			jsonError(w, fmt.Sprintf("Insufficient stock for product %s at location %s (have %d, need %d)",
				dec.ProductID, dec.StoredInID, available, dec.Quantity), http.StatusBadRequest)
			return
		}

		_, err = tx.Exec(`
			UPDATE product_stored_in 
			SET quantity = quantity - $1 
			WHERE product_id = $2 AND stored_in_id = $3
		`, dec.Quantity, dec.ProductID, dec.StoredInID)
		if err != nil {
			jsonError(w, "Failed to decrement stock: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Mark order completed
	_, err = tx.Exec(`UPDATE orders SET status = 'completed', completed_at = now() WHERE id = $1`, id)
	if err != nil {
		jsonError(w, "Failed to complete order", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		jsonError(w, "Failed to finalize completion", http.StatusInternalServerError)
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
