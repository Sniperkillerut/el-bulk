package store

import (
	"context"
	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
)

type StorageLocationStore struct {
	DB *sqlx.DB
}

func NewStorageLocationStore(db *sqlx.DB) *StorageLocationStore {
	return &StorageLocationStore{DB: db}
}

func (s *StorageLocationStore) List(ctx context.Context) ([]models.StoredIn, error) {
	var locations []models.StoredIn
	err := s.DB.SelectContext(ctx, &locations, `
		SELECT 
			s.id, 
			s.name, 
			COALESCE(SUM(ps.quantity), 0) AS item_count 
		FROM storage_location s 
		LEFT JOIN product_storage ps ON s.id = ps.storage_id 
		GROUP BY s.id, s.name 
		ORDER BY s.name
	`)
	return locations, err
}

func (s *StorageLocationStore) GetByID(ctx context.Context, id string) (*models.StoredIn, error) {
	var loc models.StoredIn
	err := s.DB.GetContext(ctx, &loc, "SELECT id, name FROM storage_location WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &loc, nil
}

func (s *StorageLocationStore) Create(ctx context.Context, name string) (*models.StoredIn, error) {
	var loc models.StoredIn
	loc.Name = name
	err := s.DB.QueryRowContext(ctx, "INSERT INTO storage_location (name) VALUES ($1) RETURNING id", name).Scan(&loc.ID)
	if err != nil {
		return nil, err
	}
	return &loc, nil
}

func (s *StorageLocationStore) Update(ctx context.Context, id, name string) error {
	_, err := s.DB.ExecContext(ctx, "UPDATE storage_location SET name = $1 WHERE id = $2", name, id)
	return err
}

func (s *StorageLocationStore) Delete(ctx context.Context, id string) error {
	_, err := s.DB.ExecContext(ctx, "DELETE FROM storage_location WHERE id = $1", id)
	return err
}
