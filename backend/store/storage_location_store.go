package store

import (
	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
)

type StorageLocationStore struct {
	DB *sqlx.DB
}

func NewStorageLocationStore(db *sqlx.DB) *StorageLocationStore {
	return &StorageLocationStore{DB: db}
}

func (s *StorageLocationStore) List() ([]models.StoredIn, error) {
	var locations []models.StoredIn
	err := s.DB.Select(&locations, `
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

func (s *StorageLocationStore) Create(name string) (*models.StoredIn, error) {
	var loc models.StoredIn
	loc.Name = name
	err := s.DB.QueryRow("INSERT INTO storage_location (name) VALUES ($1) RETURNING id", name).Scan(&loc.ID)
	if err != nil {
		return nil, err
	}
	return &loc, nil
}

func (s *StorageLocationStore) Update(id, name string) error {
	_, err := s.DB.Exec("UPDATE storage_location SET name = $1 WHERE id = $2", name, id)
	return err
}

func (s *StorageLocationStore) Delete(id string) error {
	_, err := s.DB.Exec("DELETE FROM storage_location WHERE id = $1", id)
	return err
}
