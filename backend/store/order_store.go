package store

import (
	"encoding/json"
	"fmt"

	"github.com/el-bulk/backend/models"
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

func (s *OrderStore) ListWithCustomer(whereClause string, args []interface{}, limit, offset int) ([]models.OrderWithCustomer, error) {
	orders := []models.OrderWithCustomer{}
	query := fmt.Sprintf(`SELECT * FROM view_order_list o %s ORDER BY o.created_at DESC LIMIT %d OFFSET %d`, whereClause, limit, offset)
	err := s.DB.Select(&orders, s.DB.Rebind(query), args...)
	if orders == nil {
		orders = []models.OrderWithCustomer{}
	}
	return orders, err
}

func (s *OrderStore) GetOrderCount(whereClause string, args []interface{}) (int, error) {
	var total int
	query := fmt.Sprintf(`SELECT COUNT(*) FROM view_order_list o %s`, whereClause)
	err := s.DB.Get(&total, s.DB.Rebind(query), args...)
	return total, err
}

func (s *OrderStore) GetEnrichedItems(orderID string) ([]models.OrderItemDetail, error) {
	var rows []struct {
		models.OrderItem
		ImageURL     *string `db:"image_url"`
		Stock        int     `db:"stock"`
		StoredInJSON []byte  `db:"stored_in"`
	}
	
	err := s.DB.Select(&rows, `SELECT * FROM view_order_item_enriched WHERE order_id = $1 ORDER BY product_name`, orderID)
	if err != nil {
		return nil, err
	}

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

func (s *OrderStore) PlaceOrder(customerJSON, itemsJSON, metaJSON string) (string, string, error) {
	var result struct {
		OrderID     string `db:"order_id"`
		OrderNumber string `db:"order_number"`
	}
	err := s.DB.Get(&result, "SELECT order_id, order_number FROM fn_place_order($1, $2, $3)",
		customerJSON, itemsJSON, metaJSON)
	return result.OrderID, result.OrderNumber, err
}

func (s *OrderStore) ConfirmOrder(orderID, decrementsJSON string) error {
	_, err := s.DB.Exec("SELECT fn_confirm_order($1::uuid, $2::jsonb)", orderID, decrementsJSON)
	return err
}

func (s *OrderStore) RestoreStock(orderID, incrementsJSON string) error {
	_, err := s.DB.Exec("SELECT fn_restore_order_stock($1::uuid, $2::jsonb)", orderID, incrementsJSON)
	return err
}

func (s *OrderStore) GetItemsByOrderID(orderID string) ([]models.OrderItem, error) {
	var items []models.OrderItem
	err := s.DB.Select(&items, `SELECT * FROM order_item WHERE order_id = $1`, orderID)
	return items, err
}
