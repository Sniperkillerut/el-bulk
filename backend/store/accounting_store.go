package store

import (
	"context"
	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
)

type AccountingStore struct {
	DB *sqlx.DB
}

func NewAccountingStore(db *sqlx.DB) *AccountingStore {
	return &AccountingStore{DB: db}
}

type AccountingRow struct {
	Date    interface{} `db:"date"`
	Type    string      `db:"type"`
	Ref     string      `db:"ref"`
	Detail  string      `db:"detail"`
	Income  float64     `db:"income"`
	Outcome float64     `db:"outcome"`
	Notes   string      `db:"notes"`
}

func (s *AccountingStore) GetInventoryStats(ctx context.Context) (totalItems int, totalStock int, err error) {
	var stats struct {
		TotalItems int `db:"total_items"`
		TotalStock int `db:"total_stock"`
	}
	err = s.DB.GetContext(ctx, &stats, "SELECT COUNT(*) as total_items, SUM(stock) as total_stock FROM product WHERE stock > 0")
	return stats.TotalItems, stats.TotalStock, err
}

func (s *AccountingStore) GetInventoryValuation(ctx context.Context, usdRate, eurRate, ckRate float64) (*models.InventoryValuation, error) {
	totalItems, totalStock, err := s.GetInventoryStats(ctx)
	if err != nil {
		return nil, err
	}

	valQuery := `
		SELECT 
			SUM(stock * COALESCE(price_cop_override,
				CASE price_source
					WHEN 'tcgplayer' THEN price_reference * $1
					WHEN 'cardmarket' THEN price_reference * $2
					WHEN 'cardkingdom' THEN price_reference * $3
					ELSE 0
				END, 0)) as total_value_cop,
			SUM(stock * cost_basis_cop) as total_cost_basis_cop
		FROM product 
		WHERE stock > 0
	`
	var totals struct {
		Value float64 `db:"total_value_cop"`
		Cost  float64 `db:"total_cost_basis_cop"`
	}
	err = s.DB.GetContext(ctx, &totals, valQuery, usdRate, eurRate, ckRate)
	if err != nil {
		return nil, err
	}

	return &models.InventoryValuation{
		TotalItems:        totalItems,
		TotalStock:        totalStock,
		TotalValueCOP:     totals.Value,
		TotalCostBasisCOP: totals.Cost,
		PotentialProfit:   totals.Value - totals.Cost,
	}, nil
}
