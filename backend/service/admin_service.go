package service

import (
	"context"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
)

type AdminService struct {
	Store *store.AdminStore
}

func NewAdminService(s *store.AdminStore) *AdminService {
	return &AdminService{Store: s}
}

func (s *AdminService) GetByUsername(ctx context.Context, username string) (*models.Admin, error) {
	return s.Store.GetByUsername(ctx, username)
}

func (s *AdminService) GetByID(ctx context.Context, id string) (*models.Admin, error) {
	return s.Store.GetByID(ctx, id)
}

func (s *AdminService) GetByEmail(ctx context.Context, email string) (*models.Admin, error) {
	return s.Store.GetByEmail(ctx, email)
}

func (s *AdminService) Create(ctx context.Context, admin models.Admin) (*models.Admin, error) {
	return s.Store.Create(ctx, admin)
}
