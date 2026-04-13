package service

import (
	"context"
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

func (s *StorageLocationService) List(ctx context.Context) ([]models.StoredIn, error) {
	locs, err := s.Store.List(ctx)
	if err != nil {
		return nil, err
	}
	if locs == nil {
		locs = []models.StoredIn{}
	}
	return locs, nil
}

func (s *StorageLocationService) Create(ctx context.Context, name string) (*models.StoredIn, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	return s.Store.Create(ctx, name)
}

func (s *StorageLocationService) Update(ctx context.Context, id, name string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	return s.Store.Update(ctx, id, name)
}

func (s *StorageLocationService) Delete(ctx context.Context, id string) error {
	return s.Store.Delete(ctx, id)
}
