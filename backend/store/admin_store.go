package store

import (
	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
)

type AdminStore struct {
	DB *sqlx.DB
}

func NewAdminStore(db *sqlx.DB) *AdminStore {
	return &AdminStore{DB: db}
}

func (s *AdminStore) GetByUsername(username string) (*models.Admin, error) {
	var admin models.Admin
	err := s.DB.Get(&admin, "SELECT * FROM admin WHERE username = $1", username)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}
