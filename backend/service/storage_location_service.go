package service

import (
	"context"
	"fmt"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
)

type StorageLocationService struct {
	Store *store.StorageLocationStore
	Audit Auditer
}

func NewStorageLocationService(s *store.StorageLocationStore, a Auditer) *StorageLocationService {
	return &StorageLocationService{Store: s, Audit: a}
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

func (s *StorageLocationService) Create(ctx context.Context, name string, id *string) (*models.StoredIn, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	loc, err := s.Store.Create(ctx, name, id)
	if err == nil {
		s.Audit.LogAction(ctx, "CREATE_STORAGE", "storage", loc.ID, models.JSONB{"name": name})
	}
	return loc, err
}

func (s *StorageLocationService) Update(ctx context.Context, id, name string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	before, _ := s.Store.GetByID(ctx, id)
	err := s.Store.Update(ctx, id, name)
	if err == nil {
		s.Audit.LogAction(ctx, "UPDATE_STORAGE", "storage", id, models.JSONB{"before": before, "after": name})
	}
	return err
}

func (s *StorageLocationService) Delete(ctx context.Context, id string) error {
	before, _ := s.Store.GetByID(ctx, id)
	
	// Deep Capture: Get all stock mappings before they are cascaded away
	mappings, _ := s.Store.GetStockMappings(ctx, id)
	
	err := s.Store.Delete(ctx, id)
	if err == nil {
		s.Audit.LogAction(ctx, "DELETE_STORAGE", "storage", id, models.JSONB{
			"deleted": before,
			"stock_mappings": mappings,
		})
	}
	return err
}
