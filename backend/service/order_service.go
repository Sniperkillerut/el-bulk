package service

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"context"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/el-bulk/backend/utils/crypto"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

type OrderService struct {
	Store         *store.OrderStore
	ProductStore  *store.ProductStore
	CustomerStore *store.CustomerStore
	Settings      SettingsProvider
	Audit         Auditer
}

func NewOrderService(s *store.OrderStore, ps *store.ProductStore, cs *store.CustomerStore, settings SettingsProvider, audit Auditer) *OrderService {
	return &OrderService{
		Store:         s,
		ProductStore:  ps,
		CustomerStore: cs,
		Settings:      settings,
		Audit:         audit,
	}
}

// CreateOrder handles the public checkout flow
func (s *OrderService) CreateOrder(ctx context.Context, input models.CreateOrderInput, customerID string) (string, string, float64, error) {
	logger.TraceCtx(ctx, "Entering OrderService.CreateOrder | CustomerID: %s | Items: %d", customerID, len(input.Items))
	// 1. Load settings for rates and shipping
	settings, err := s.Settings.GetSettings(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to get settings in OrderService.CreateOrder: %v", err)
	}

	// 2. Fetch products and compute prices
	productIDs := make([]string, 0)
	for _, item := range input.Items {
		if item.Quantity > 0 {
			productIDs = append(productIDs, item.ProductID)
		}
	}

	if len(productIDs) == 0 {
		return "", "", 0, fmt.Errorf("no valid items selected")
	}

	// Fetch products directly for now
	var products []models.Product
	query, args, err := sqlx.In(`SELECT * FROM product WHERE id IN (?)`, productIDs)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to build product query: %w", err)
	}
	if err := s.Store.DB.SelectContext(ctx, &products, s.Store.DB.Rebind(query), args...); err != nil {
		return "", "", 0, fmt.Errorf("failed to fetch products: %w", err)
	}

	productMap := make(map[string]models.Product)
	for _, p := range products {
		productMap[p.ID] = p
	}

	// 3. Fetch storage locations for snapshots (excluding 'pending')
	storageMap := make(map[string][]models.StorageLocation)
	type storageWithPID struct {
		models.StorageLocation
		ProductID string `db:"product_id_temp"`
	}
	var rows []storageWithPID
	qStorage, aStorage, err := sqlx.In(`
		SELECT ps.product_id as product_id_temp, ps.storage_id, sl.name, ps.quantity 
		FROM product_storage ps 
		JOIN storage_location sl ON ps.storage_id = sl.id 
		WHERE ps.product_id IN (?) AND ps.quantity > 0 AND sl.name != 'pending'
	`, productIDs)
	if err == nil {
		err = s.Store.DB.SelectContext(ctx, &rows, s.Store.DB.Rebind(qStorage), aStorage...)
		if err == nil {
			for _, row := range rows {
				storageMap[row.ProductID] = append(storageMap[row.ProductID], row.StorageLocation)
			}
		}
	}

	// 4. Prepare data for PlaceOrder
	subtotalCOP := 0.0
	orderItems := make([]map[string]interface{}, 0)
	for _, item := range input.Items {
		p, ok := productMap[item.ProductID]
		if !ok || item.Quantity <= 0 {
			continue
		}
		price := p.ComputePrice(settings.USDToCOPRate, settings.EURToCOPRate, settings.CKToCOPRate)
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
		return "", "", 0, fmt.Errorf("no valid items selected")
	}

	shippingCOP := 0.0
	if !input.IsLocalPickup {
		if input.IsPriority {
			shippingCOP = settings.PriorityShippingFeeCOP
			if shippingCOP == 0 {
				// Fallback if not set specifically, e.g. 1.5x standard
				shippingCOP = settings.FlatShippingFeeCOP * 1.5
			}
		} else {
			shippingCOP = settings.FlatShippingFeeCOP
		}
	}
	totalCOP := subtotalCOP + shippingCOP

	// 5. Encrypt PII
	encPhone, _ := crypto.Encrypt(input.Phone)
	encIDNumber, _ := crypto.Encrypt(input.IDNumber)
	encAddress, _ := crypto.Encrypt(input.Address)

	customerJSON, _ := json.Marshal(map[string]interface{}{
		"id":         customerID,
		"first_name": input.FirstName,
		"last_name":  input.LastName,
		"email":      input.Email,
		"phone":      encPhone,
		"id_number":  encIDNumber,
		"address":    encAddress,
	})
	itemsJSON, _ := json.Marshal(orderItems)
	metaJSON, _ := json.Marshal(map[string]interface{}{
		"order_number":    s.GenerateOrderNumber(),
		"payment_method":  input.PaymentMethod,
		"subtotal_cop":    subtotalCOP,
		"shipping_cop":    shippingCOP,
		"tax_cop":         0.0,
		"total_cop":       totalCOP,
		"is_local_pickup": input.IsLocalPickup,
		"notes":           input.Notes,
	})

	// 6. Execute Store method
	orderID, orderNumber, err := s.Store.PlaceOrder(ctx, string(customerJSON), string(itemsJSON), string(metaJSON))
	if err == nil {
		logger.DebugCtx(ctx, "Order placed successfully: %s (%s) for total: %.2f COP", orderNumber, orderID, totalCOP)
	} else {
		logger.ErrorCtx(ctx, "Order placement failed: %v", err)
	}
	return orderID, orderNumber, totalCOP, err
}

func (s *OrderService) GetOrderDetail(ctx context.Context, orderID string, isAdmin bool) (*models.OrderDetail, error) {
	logger.TraceCtx(ctx, "Entering OrderService.GetOrderDetail | OrderID: %s", orderID)
	order, err := s.Store.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	customer, err := s.CustomerStore.GetByID(ctx, order.CustomerID)
	if err != nil {
		return nil, err
	}

	// Decrypt PII
	customer.Phone = crypto.DecryptSafe(customer.Phone)
	customer.IDNumber = crypto.DecryptSafe(customer.IDNumber)
	customer.Address = crypto.DecryptSafe(customer.Address)

	items, err := s.Store.GetEnrichedItems(ctx, orderID)
	if err != nil {
		return nil, err
	}

	whatsappURL := s.GenerateWhatsAppURL(*order, *customer)

	res := &models.OrderDetail{
		Order:       *order,
		Customer:    *customer,
		Items:       items,
		WhatsAppURL: whatsappURL,
	}
	res.Redact(isAdmin)

	return res, nil
}

func (s *OrderService) UpdateOrder(ctx context.Context, orderID string, input models.UpdateOrderInput) error {
	logger.TraceCtx(ctx, "Entering OrderService.UpdateOrder | OrderID: %s | NewStatus: %v", orderID, input.Status)

	tx, err := s.Store.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	oldOrder, _ := s.Store.GetByID(ctx, orderID)

	// 0. Check if inventory modification is allowed
	var currentStatus string
	err = tx.GetContext(ctx, &currentStatus, `SELECT status FROM "order" WHERE id = $1`, orderID)
	if err != nil {
		return fmt.Errorf("failed to get current order status: %w", err)
	}

	if currentStatus == "completed" || currentStatus == "cancelled" {
		return fmt.Errorf("cannot modify an order in terminal state: %s", currentStatus)
	}

	inventoryChanged := len(input.Items) > 0 || len(input.AddedItems) > 0 || len(input.DeletedIDs) > 0
	if inventoryChanged && currentStatus != "pending" {
		return fmt.Errorf("cannot modify items of an order in status: %s", currentStatus)
	}

	metadataChanged := input.PaymentMethod != nil || input.ShippingCOP != nil
	if metadataChanged && currentStatus != "pending" && currentStatus != "confirmed" {
		return fmt.Errorf("cannot modify payment/shipping for order in status: %s", currentStatus)
	}

	// 1. Update status, tracking, payment, and shipping if provided
	if input.Status != nil || input.TrackingNumber != nil || input.TrackingURL != nil || input.PaymentMethod != nil || input.ShippingCOP != nil {
		if input.Status != nil {
			newStatus := *input.Status
			// Restriction: Cannot skip confirmation
			isPostConfirmation := newStatus == "ready_for_pickup" || newStatus == "shipped" || newStatus == "completed"
			if isPostConfirmation && currentStatus == "pending" {
				return fmt.Errorf("order must be confirmed before moving to status: %s", newStatus)
			}

			// Existing restriction: No back to pending
			if newStatus == "pending" && (currentStatus == "confirmed" || currentStatus == "shipped" || currentStatus == "completed") {
				return fmt.Errorf("cannot move a confirmed/shipped/completed order back to pending")
			}
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE "order" 
			SET status = COALESCE($1, status),
			    tracking_number = COALESCE($2, tracking_number),
			    tracking_url = COALESCE($3, tracking_url),
				payment_method = COALESCE($4, payment_method),
				shipping_cop = COALESCE($5, shipping_cop)
			WHERE id = $6`,
			input.Status, input.TrackingNumber, input.TrackingURL, input.PaymentMethod, input.ShippingCOP, orderID)
		if err != nil {
			return fmt.Errorf("failed to update order info: %w", err)
		}
	}

	// 2. Update existing item quantities
	if len(input.Items) > 0 {
		itemIDs := make([]string, 0, len(input.Items))
		for _, item := range input.Items {
			itemIDs = append(itemIDs, item.ID)
		}

		query, args, err := sqlx.In(`
			SELECT oi.id, oi.quantity, oi.product_id, p.stock
			FROM order_item oi 
			JOIN product p ON p.id = oi.product_id 
			WHERE oi.order_id = ? AND oi.id IN (?)
		`, orderID, itemIDs)
		if err != nil {
			return fmt.Errorf("failed to build query for existing items: %w", err)
		}

		type CurrentItem struct {
			ID        string `db:"id"`
			Quantity  int    `db:"quantity"`
			ProductID string `db:"product_id"`
			Stock     int    `db:"stock"`
		}

		var currentItems []CurrentItem
		err = tx.SelectContext(ctx, &currentItems, s.Store.DB.Rebind(query), args...)
		if err != nil {
			return fmt.Errorf("failed to fetch existing items: %w", err)
		}

		currentMap := make(map[string]CurrentItem)
		for _, item := range currentItems {
			currentMap[item.ID] = item
		}

		for _, item := range input.Items {
			qty := item.Quantity
			if qty < 0 {
				qty = 0
			}

			current, ok := currentMap[item.ID]
			if !ok {
				continue // skip invalid items
			}

			delta := qty - current.Quantity
			if delta == 0 {
				continue
			}

			if delta > 0 && delta > current.Stock {
				return fmt.Errorf("cannot increase order by %d; only %d in stock", delta, current.Stock)
			}

			_, err = tx.ExecContext(ctx, `UPDATE order_item SET quantity = $1 WHERE id = $2 AND order_id = $3`,
				qty, item.ID, orderID)
			if err != nil {
				return fmt.Errorf("failed to update item quantity: %w", err)
			}

			_, err = tx.ExecContext(ctx, `
				UPDATE product_storage
				SET quantity = quantity + $1
				WHERE product_id = $2
				  AND storage_id = (SELECT id FROM storage_location WHERE name = 'pending' LIMIT 1)
			`, delta, current.ProductID)
			if err != nil {
				return fmt.Errorf("failed to update pending storage holding: %w", err)
			}
		}
	}

	// 3. Add new items
	if len(input.AddedItems) > 0 {
		var productIDs []string
		for _, item := range input.AddedItems {
			if item.Quantity > 0 {
				productIDs = append(productIDs, item.ProductID)
			}
		}

		type ProductDetails struct {
			ID            string  `db:"id"`
			Name          string  `db:"name"`
			SetName       *string `db:"set_name"`
			FoilTreatment string  `db:"foil_treatment"`
			CardTreatment string  `db:"card_treatment"`
			Condition     *string `db:"condition"`
			Stock         int     `db:"stock"`
		}
		productMap := make(map[string]ProductDetails)

		if len(productIDs) > 0 {
			query, args, err := sqlx.In(`SELECT id, name, set_name, foil_treatment, card_treatment, condition, stock FROM product WHERE id IN (?)`, productIDs)
			if err != nil {
				return fmt.Errorf("failed to build product query: %w", err)
			}
			query = tx.Rebind(query)

			var products []ProductDetails
			err = tx.SelectContext(ctx, &products, query, args...)
			if err != nil {
				return fmt.Errorf("failed to fetch products for new items: %w", err)
			}

			for _, p := range products {
				productMap[p.ID] = p
			}
		}

		var orderItemArgs []interface{}
		var orderItemVals []string

		var storageArgs []interface{}
		var storageVals []string

		valIdxOI := 1
		valIdxPS := 1
		hasAdded := false

		for _, item := range input.AddedItems {
			if item.Quantity <= 0 {
				continue
			}

			product, ok := productMap[item.ProductID]
			if !ok {
				return fmt.Errorf("product not found while adding to order: product %s", item.ProductID)
			}

			if item.Quantity > product.Stock {
				return fmt.Errorf("added quantity %d exceeds available stock %d", item.Quantity, product.Stock)
			}

			orderItemVals = append(orderItemVals, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				valIdxOI, valIdxOI+1, valIdxOI+2, valIdxOI+3, valIdxOI+4, valIdxOI+5, valIdxOI+6, valIdxOI+7, valIdxOI+8))
			orderItemArgs = append(orderItemArgs, orderID, item.ProductID, product.Name, product.SetName, product.FoilTreatment, product.CardTreatment, product.Condition, item.Quantity, item.UnitPriceCOP)
			valIdxOI += 9

			storageVals = append(storageVals, fmt.Sprintf("($%d, (SELECT id FROM storage_location WHERE name = 'pending' LIMIT 1), $%d)", valIdxPS, valIdxPS+1))
			storageArgs = append(storageArgs, item.ProductID, item.Quantity)
			valIdxPS += 2
			hasAdded = true
		}

		if hasAdded {
			orderItemQ := fmt.Sprintf(`
			INSERT INTO order_item (order_id, product_id, product_name, product_set, foil_treatment, card_treatment, condition, quantity, unit_price_cop)
			VALUES %s
			`, strings.Join(orderItemVals, ","))
			_, err = tx.ExecContext(ctx, orderItemQ, orderItemArgs...)
			if err != nil {
				return fmt.Errorf("failed to bulk add new items to order: %w", err)
			}

			storageQ := fmt.Sprintf(`
				INSERT INTO product_storage (product_id, storage_id, quantity)
				VALUES %s
				ON CONFLICT (product_id, storage_id) DO UPDATE SET quantity = product_storage.quantity + EXCLUDED.quantity
			`, strings.Join(storageVals, ","))
			_, err = tx.ExecContext(ctx, storageQ, storageArgs...)
			if err != nil {
				return fmt.Errorf("failed to bulk update pending storage for added items: %w", err)
			}
		}
	}

	// 4. Delete items
	for _, itemID := range input.DeletedIDs {
		var current struct {
			Quantity  int    `db:"quantity"`
			ProductID string `db:"product_id"`
		}
		err = tx.GetContext(ctx, &current, `SELECT quantity, product_id FROM order_item WHERE id = $1 AND order_id = $2`, itemID, orderID)
		if err != nil {
			continue
		}

		_, err = tx.ExecContext(ctx, `DELETE FROM order_item WHERE id = $1 AND order_id = $2`, itemID, orderID)
		if err != nil {
			return fmt.Errorf("failed to delete item from order: %w", err)
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE product_storage 
			SET quantity = GREATEST(0, quantity - $1) 
			WHERE product_id = $2 
			  AND storage_id = (SELECT id FROM storage_location WHERE name = 'pending' LIMIT 1)
		`, current.Quantity, current.ProductID)
		if err != nil {
			return fmt.Errorf("failed to release pending stock for deleted item: %w", err)
		}
	}

	// 5. Recalculate subtotal and total
	var summary struct {
		Subtotal float64 `db:"subtotal"`
		Shipping float64 `db:"shipping"`
		Tax      float64 `db:"tax"`
	}
	err = tx.GetContext(ctx, &summary, `
		SELECT 
			COALESCE((SELECT SUM(unit_price_cop * quantity) FROM order_item WHERE order_id = $1), 0) as subtotal,
			COALESCE(shipping_cop, 0) as shipping,
			COALESCE(tax_cop, 0) as tax
		FROM "order" WHERE id = $1
	`, orderID)

	if err == nil {
		newTotal := summary.Subtotal + summary.Shipping + summary.Tax
		_, err = tx.ExecContext(ctx, `UPDATE "order" SET subtotal_cop = $1, total_cop = $2 WHERE id = $3`,
			summary.Subtotal, newTotal, orderID)
		if err != nil {
			return fmt.Errorf("failed to update order totals: %w", err)
		}
		logger.DebugCtx(ctx, "Order %s recalculated: subtotal=%.2f, total=%.2f", orderID, summary.Subtotal, newTotal)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	if s.Audit != nil {
		s.Audit.LogAction(ctx, "UPDATE_ORDER", "order", orderID, models.JSONB{
			"before": oldOrder,
			"after":  input,
		})
	}

	logger.InfoCtx(ctx, "Order %s updated successfully", orderID)
	return nil
}

func (s *OrderService) ConfirmOrder(ctx context.Context, orderID string, decrements []models.StockDecrement) error {
	jsonData, err := json.Marshal(decrements)
	if err != nil {
		return err
	}
	err = s.Store.ConfirmOrder(ctx, orderID, string(jsonData))
	if err == nil && s.Audit != nil {
		s.Audit.LogAction(ctx, "CONFIRM_ORDER", "order", orderID, models.JSONB{"decrements": decrements})
	}
	return err
}

func (s *OrderService) RestoreStock(ctx context.Context, orderID string, increments []models.StockDecrement) error {
	jsonData, err := json.Marshal(increments)
	if err != nil {
		return err
	}
	err = s.Store.RestoreStock(ctx, orderID, string(jsonData))
	if err == nil && s.Audit != nil {
		s.Audit.LogAction(ctx, "RESTORE_STOCK", "order", orderID, models.JSONB{"increments": increments})
	}
	return err
}

func (s *OrderService) ListOrders(ctx context.Context, whereClause string, args []interface{}, page, pageSize int) ([]models.OrderWithCustomer, int, error) {
	logger.TraceCtx(ctx, "Entering OrderService.ListOrders | Page: %d | PageSize: %d", page, pageSize)
	total, err := s.Store.GetOrderCount(ctx, whereClause, args)
	if err != nil {
		return nil, 0, err
	}

	limit := pageSize
	offset := (page - 1) * pageSize
	orders, err := s.Store.ListWithCustomer(ctx, whereClause, args, limit, offset)
	return orders, total, err
}

func (s *OrderService) ListMe(ctx context.Context, userID string) ([]models.OrderWithItemCount, error) {
	var orders []models.OrderWithItemCount
	err := s.Store.DB.SelectContext(ctx, &orders, `
		SELECT o.*, (SELECT SUM(quantity) FROM order_item WHERE order_id = o.id) as item_count
		FROM "order" o 
		WHERE o.customer_id = $1 
		ORDER BY o.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	if orders == nil {
		orders = []models.OrderWithItemCount{}
	}
	return orders, nil
}

func (s *OrderService) CancelMe(ctx context.Context, orderID, userID string) error {
	logger.TraceCtx(ctx, "Entering OrderService.CancelMe | OrderID: %s | UserID: %s", orderID, userID)
	tx, err := s.Store.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
		UPDATE "order" 
		SET status = 'cancelled' 
		WHERE id = $1 AND customer_id = $2 AND status = 'pending'
	`, orderID, userID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("order cannot be cancelled or not found")
	}

	return tx.Commit()
}

// Helpers
func (s *OrderService) GenerateOrderNumber() string {
	now := time.Now()
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// This should never happen with crypto/rand
		logger.Error("Failed to generate random bytes for order number: %v", err)
	}
	return fmt.Sprintf("EB-%s-%016X", now.Format("20060102"), b)
}

func (s *OrderService) GenerateWhatsAppURL(order models.Order, customer models.Customer) string {
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
