package service

import (
	"fmt"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
)

type StorageLocationService struct {
	Store *store.StorageLocationStore
}

func NewStorageLocationService(s *store.StorageLocationStore) *StorageLocationService {
	return &StorageLocationService{Store: s}
}

func (s *StorageLocationService) List() ([]models.StoredIn, error) {
	locs, err := s.Store.List()
	if err != nil {
		return nil, err
	}
	if locs == nil {
		locs = []models.StoredIn{}
	}
	return locs, nil
}

func (s *StorageLocationService) Create(name string) (*models.StoredIn, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	return s.Store.Create(name)
}

func (s *StorageLocationService) Update(id, name string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	return s.Store.Update(id, name)
}

func (s *StorageLocationService) Delete(id string) error {
	return s.Store.Delete(id)
}
