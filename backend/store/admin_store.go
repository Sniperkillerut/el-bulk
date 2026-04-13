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
	err := s.DB.Unsafe().Get(&admin, "SELECT * FROM admin WHERE username = $1", username)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (s *AdminStore) GetByID(id string) (*models.Admin, error) {
	var admin models.Admin
	err := s.DB.Unsafe().Get(&admin, "SELECT * FROM admin WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (s *AdminStore) GetByEmail(email string) (*models.Admin, error) {
	var admin models.Admin
	err := s.DB.Unsafe().Get(&admin, "SELECT * FROM admin WHERE email = $1", email)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (s *AdminStore) Create(admin models.Admin) (*models.Admin, error) {
	query := `
		INSERT INTO admin (username, email, password_hash, avatar_url)
		VALUES ($1, $2, $3, $4)
		RETURNING *
	`
	var newAdmin models.Admin
	err := s.DB.Unsafe().Get(&newAdmin, query, admin.Username, admin.Email, admin.PasswordHash, admin.AvatarURL)
	if err != nil {
		return nil, err
	}
	return &newAdmin, nil
}
