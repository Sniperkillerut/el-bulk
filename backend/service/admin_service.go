package service

import (
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
)

type AdminService struct {
	Store *store.AdminStore
}

func NewAdminService(s *store.AdminStore) *AdminService {
	return &AdminService{Store: s}
}

func (s *AdminService) GetByUsername(username string) (*models.Admin, error) {
	return s.Store.GetByUsername(username)
}
