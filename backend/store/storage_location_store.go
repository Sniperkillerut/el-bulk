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

func (s *StorageLocationStore) Create(ctx context.Context, name string, id *string) (*models.StoredIn, error) {
	var loc models.StoredIn
	loc.Name = name

	query := "INSERT INTO storage_location (id, name) VALUES (COALESCE($1, gen_random_uuid()), $2) RETURNING id"
	err := s.DB.QueryRowContext(ctx, query, id, name).Scan(&loc.ID)
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

func (s *StorageLocationStore) GetStockMappings(ctx context.Context, storageID string) ([]models.ProductStorage, error) {
	var mappings []models.ProductStorage
	query := "SELECT product_id, storage_id, quantity FROM product_storage WHERE storage_id = $1"
	err := s.DB.SelectContext(ctx, &mappings, query, storageID)
	return mappings, err
}

func (s *StorageLocationStore) BatchRestoreStock(ctx context.Context, storageID string, mappings []models.ProductStorage) error {
	if len(mappings) == 0 {
		return nil
	}

	tx, err := s.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := "INSERT INTO product_storage (product_id, storage_id, quantity) VALUES ($1, $2, $3) ON CONFLICT (product_id, storage_id) DO UPDATE SET quantity = EXCLUDED.quantity"
	stmt, err := tx.PreparexContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, m := range mappings {
		if _, err := stmt.ExecContext(ctx, m.ProductID, storageID, m.Quantity); err != nil {
			return err
		}
	}

	return tx.Commit()
}
