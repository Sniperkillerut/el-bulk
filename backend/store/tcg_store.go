package store

import (
	"github.com/el-bulk/backend/models"
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

func (s *TCGStore) ListWithCount() ([]models.TCG, error) {
	var tcgs []models.TCG
	err := s.DB.Select(&tcgs, `
		SELECT t.*, COUNT(p.id) as item_count 
		FROM tcg t 
		LEFT JOIN product p ON t.id = p.tcg 
		GROUP BY t.id 
		ORDER BY t.name
	`)
	if err != nil {
		return nil, err
	}
	if tcgs == nil {
		tcgs = []models.TCG{}
	}
	return tcgs, nil
}

func (s *TCGStore) Create(input models.TCGInput) (*models.TCG, error) {
	var tcg models.TCG
	err := s.DB.QueryRowx(`
		INSERT INTO tcg (id, name, is_active)
		VALUES ($1, $2, $3)
		RETURNING *
	`, input.ID, input.Name, true).StructScan(&tcg)
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
	var count int
	err := s.DB.Get(&count, "SELECT COUNT(*) FROM product WHERE tcg = $1", id)
	return count, err
}
