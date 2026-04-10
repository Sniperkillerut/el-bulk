package store

import (
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
	"time"
)

type TCGStore struct {
	*BaseStore[models.TCG]
}

func NewTCGStore(db *sqlx.DB) *TCGStore {
	return &TCGStore{
		BaseStore: NewBaseStore[models.TCG](db, "tcg"),
	}
}

func (s *TCGStore) ListWithCount() ([]models.TCG, error) {
	start := time.Now()
	var tcgs []models.TCG
	query := `
		SELECT t.*, COUNT(p.id) as item_count 
		FROM tcg t 
		LEFT JOIN product p ON t.id = p.tcg 
		GROUP BY t.id 
		ORDER BY t.name
	`
	logger.Trace("[DB] Executing ListWithCount (TCG): %s", query)
	err := s.DB.Select(&tcgs, query)
	if err != nil {
		logger.Error("[DB] ListWithCount (TCG) failed: %v", err)
		return nil, err
	}
	if tcgs == nil {
		tcgs = []models.TCG{}
	}
	logger.Debug("[DB] ListWithCount (TCG) took %v", time.Since(start))
	return tcgs, nil
}

func (s *TCGStore) Create(input models.TCGInput) (*models.TCG, error) {
	var tcg models.TCG
	query := `
		INSERT INTO tcg (id, name, is_active)
		VALUES ($1, $2, $3)
		RETURNING *
	`
	logger.Trace("[DB] Executing CreateTCG: %s | ID: %s", query, input.ID)
	err := s.DB.QueryRowx(query, input.ID, input.Name, true).StructScan(&tcg)
	if err != nil {
		logger.Error("[DB] CreateTCG failed for %s: %v", input.ID, err)
	}
	return &tcg, err
}

func (s *TCGStore) Update(id string, input models.TCGInput) (*models.TCG, error) {
	var tcg models.TCG
	err := s.DB.QueryRowx(`
		UPDATE tcg
		SET name = $1, is_active = $2
		WHERE id = $3
		RETURNING *
	`, input.Name, input.IsActive, id).StructScan(&tcg)
	return &tcg, err
}

func (s *TCGStore) GetProductCount(id string) (int, error) {
	start := time.Now()
	var count int
	query := "SELECT COUNT(*) FROM product WHERE tcg = $1"
	logger.Trace("[DB] Executing GetProductCount (TCG) for %s: %s", id, query)
	err := s.DB.Get(&count, query, id)
	if err != nil {
		logger.Error("[DB] GetProductCount (TCG) failed for %s: %v", id, err)
	}
	logger.Debug("[DB] GetProductCount (TCG) for %s took %v", id, time.Since(start))
	return count, err
}
