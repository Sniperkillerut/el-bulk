package store

import (
	"context"
	"fmt"
	"time"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

type TCGStore struct {
	*BaseStore[models.TCG]
}

func NewTCGStore(db *sqlx.DB) *TCGStore {
	return &TCGStore{
		BaseStore: NewBaseStore[models.TCG](db, "tcg"),
	}
}

func (s *TCGStore) ListWithCount(ctx context.Context, activeOnly bool) ([]models.TCG, error) {
	start := time.Now()
	var tcgs []models.TCG
	
	where := ""
	if activeOnly {
		where = "WHERE t.is_active = true"
	}
	
	query := fmt.Sprintf(`
		SELECT t.*, COUNT(p.id) as item_count 
		FROM tcg t 
		LEFT JOIN product p ON t.id = p.tcg 
		%s
		GROUP BY t.id 
		ORDER BY t.name
	`, where)
	
	logger.TraceCtx(ctx, "[DB] Executing ListWithCount (TCG): %s", query)
	err := s.DB.Unsafe().SelectContext(ctx, &tcgs, query)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] ListWithCount (TCG) failed: %v", err)
		return nil, err
	}
	if tcgs == nil {
		tcgs = []models.TCG{}
	}
	logger.DebugCtx(ctx, "[DB] ListWithCount (TCG) took %v", time.Since(start))
	return tcgs, nil
}

func (s *TCGStore) Create(ctx context.Context, input models.TCGInput) (*models.TCG, error) {
	var tcg models.TCG
	query := `
		INSERT INTO tcg (id, name, is_active)
		VALUES ($1, $2, $3)
		RETURNING *
	`
	logger.TraceCtx(ctx, "[DB] Executing CreateTCG: %s | ID: %s", query, input.ID)
	err := s.DB.QueryRowxContext(ctx, query, input.ID, input.Name, true).StructScan(&tcg)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] CreateTCG failed for %s: %v", input.ID, err)
	}
	return &tcg, err
}

func (s *TCGStore) Update(ctx context.Context, id string, input models.TCGInput) (*models.TCG, error) {
	var tcg models.TCG
	err := s.DB.QueryRowxContext(ctx, `
		UPDATE tcg
		SET name = $1, is_active = $2
		WHERE id = $3
		RETURNING *
	`, input.Name, input.IsActive, id).StructScan(&tcg)
	return &tcg, err
}

func (s *TCGStore) GetProductCount(ctx context.Context, id string) (int, error) {
	start := time.Now()
	var count int
	query := "SELECT COUNT(*) FROM product WHERE tcg = $1"
	logger.TraceCtx(ctx, "[DB] Executing GetProductCount (TCG) for %s: %s", id, query)
	err := s.DB.GetContext(ctx, &count, query, id)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] GetProductCount (TCG) failed for %s: %v", id, err)
	}
	logger.DebugCtx(ctx, "[DB] GetProductCount (TCG) for %s took %v", id, time.Since(start))
	return count, err
}
