package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

type OrderStore struct {
	*BaseStore[models.Order]
}

func NewOrderStore(db *sqlx.DB) *OrderStore {
	return &OrderStore{
		BaseStore: NewBaseStore[models.Order](db, `"order"`),
	}
}

func (s *OrderStore) ListWithCustomer(ctx context.Context, whereClause string, args []interface{}, limit, offset int) ([]models.OrderWithCustomer, error) {
	start := time.Now()
	orders := []models.OrderWithCustomer{}
	query := fmt.Sprintf(`SELECT * FROM view_order_list o %s ORDER BY o.created_at DESC LIMIT %d OFFSET %d`, whereClause, limit, offset)
	rebound := s.DB.Rebind(query)
	logger.TraceCtx(ctx, "[DB] Executing ListWithCustomer: %s | Args: %+v", rebound, args)
	
	err := s.DB.SelectContext(ctx, &orders, rebound, args...)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] ListWithCustomer failed: %v", err)
	}
	if orders == nil {
		orders = []models.OrderWithCustomer{}
	}
	logger.DebugCtx(ctx, "[DB] ListWithCustomer took %v", time.Since(start))
	return orders, err
}

func (s *OrderStore) GetOrderCount(ctx context.Context, whereClause string, args []interface{}) (int, error) {
	start := time.Now()
	var total int
	query := fmt.Sprintf(`SELECT COUNT(*) FROM view_order_list o %s`, whereClause)
	rebound := s.DB.Rebind(query)
	logger.TraceCtx(ctx, "[DB] Executing GetOrderCount: %s | Args: %+v", rebound, args)
	err := s.DB.GetContext(ctx, &total, rebound, args...)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] GetOrderCount failed: %v", err)
	}
	logger.DebugCtx(ctx, "[DB] GetOrderCount took %v", time.Since(start))
	return total, err
}

func (s *OrderStore) GetEnrichedItems(ctx context.Context, orderID string) ([]models.OrderItemDetail, error) {
	var rows []struct {
		models.OrderItem
		ImageURL     *string `db:"image_url"`
		Stock        int     `db:"stock"`
		StoredInJSON []byte  `db:"stored_in"`
	}
	
	start := time.Now()
	query := `SELECT * FROM view_order_item_enriched WHERE order_id = $1 ORDER BY product_name`
	logger.TraceCtx(ctx, "[DB] Executing GetEnrichedItems for %s: %s", orderID, query)
	err := s.DB.SelectContext(ctx, &rows, query, orderID)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] GetEnrichedItems failed for %s: %v", orderID, err)
		return nil, err
	}
	logger.DebugCtx(ctx, "[DB] GetEnrichedItems for %s took %v", orderID, time.Since(start))

	items := make([]models.OrderItemDetail, len(rows))
	for i, r := range rows {
		items[i] = models.OrderItemDetail{
			OrderItem: r.OrderItem,
			ImageURL:  r.ImageURL,
			Stock:     r.Stock,
			StoredIn:  []models.StorageLocation{},
		}
		if r.StoredInJSON != nil {
			json.Unmarshal(r.StoredInJSON, &items[i].StoredIn)
		}
	}
	return items, nil
}

func (s *OrderStore) PlaceOrder(ctx context.Context, customerJSON, itemsJSON, metaJSON string) (string, string, error) {
	start := time.Now()
	var result struct {
		OrderID     string `db:"order_id"`
		OrderNumber string `db:"order_number"`
	}
	query := "SELECT order_id, order_number FROM fn_place_order($1, $2, $3)"
	logger.TraceCtx(ctx, "[DB] Executing PlaceOrder: %s", query)
	err := s.DB.GetContext(ctx, &result, query, customerJSON, itemsJSON, metaJSON)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] PlaceOrder failed: %v", err)
	}
	logger.DebugCtx(ctx, "[DB] PlaceOrder took %v", time.Since(start))
	return result.OrderID, result.OrderNumber, err
}

func (s *OrderStore) ConfirmOrder(ctx context.Context, orderID, decrementsJSON string) error {
	logger.TraceCtx(ctx, "[DB] Executing ConfirmOrder for %s", orderID)
	_, err := s.DB.ExecContext(ctx, "SELECT fn_confirm_order($1::uuid, $2::jsonb)", orderID, decrementsJSON)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] ConfirmOrder for %s failed: %v", orderID, err)
	}
	return err
}

func (s *OrderStore) RestoreStock(ctx context.Context, orderID, incrementsJSON string) error {
	_, err := s.DB.ExecContext(ctx, "SELECT fn_restore_order_stock($1::uuid, $2::jsonb)", orderID, incrementsJSON)
	return err
}

func (s *OrderStore) GetItemsByOrderID(ctx context.Context, orderID string) ([]models.OrderItem, error) {
	start := time.Now()
	var items []models.OrderItem
	query := `SELECT * FROM order_item WHERE order_id = $1`
	logger.TraceCtx(ctx, "[DB] Executing GetItemsByOrderID for %s: %s", orderID, query)
	err := s.DB.SelectContext(ctx, &items, query, orderID)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] GetItemsByOrderID for %s failed: %v", orderID, err)
	}
	logger.DebugCtx(ctx, "[DB] GetItemsByOrderID for %s took %v", orderID, time.Since(start))
	return items, err
}
