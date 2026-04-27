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

func (s *TCGStore) ListSets(ctx context.Context, tcgID string) ([]models.TCGSet, error) {
	start := time.Now()
	var sets []models.TCGSet
	query := "SELECT tcg, code, name, released_at, set_type, ck_name FROM tcg_set WHERE tcg = $1 ORDER BY released_at DESC"
	logger.TraceCtx(ctx, "[DB] Executing ListSets for %s: %s", tcgID, query)
	err := s.DB.SelectContext(ctx, &sets, query, tcgID)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] ListSets failed for %s: %v", tcgID, err)
		return nil, err
	}
	if sets == nil {
		sets = []models.TCGSet{}
	}
	logger.DebugCtx(ctx, "[DB] ListSets for %s took %v", tcgID, time.Since(start))
	return sets, nil
}

func (s *TCGStore) GetSetByCode(ctx context.Context, tcgID, code string) (*models.TCGSet, error) {
	start := time.Now()
	var tSet models.TCGSet
	query := "SELECT tcg, code, name, released_at, set_type, ck_name FROM tcg_set WHERE tcg = $1 AND code = $2"
	logger.TraceCtx(ctx, "[DB] Executing GetSetByCode for %s/%s: %s", tcgID, code, query)
	err := s.DB.GetContext(ctx, &tSet, query, tcgID, code)
	if err != nil {
		return nil, err
	}
	logger.DebugCtx(ctx, "[DB] GetSetByCode for %s/%s took %v", tcgID, code, time.Since(start))
	return &tSet, nil
}
