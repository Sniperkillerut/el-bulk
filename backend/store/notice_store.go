package store

import (
	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
)

type NoticeStore struct {
	*BaseStore[models.Notice]
}

func NewNoticeStore(db *sqlx.DB) *NoticeStore {
	return &NoticeStore{
		BaseStore: NewBaseStore[models.Notice](db, "notice"),
	}
}

func (s *NoticeStore) List(isPublished bool, limit int) ([]models.Notice, error) {
	query := "SELECT * FROM notice"
	var args []interface{}
	
	if isPublished {
		query += " WHERE is_published = true"
	}
	
	query += " ORDER BY created_at DESC"
	
	if limit > 0 {
		query += " LIMIT $1"
		args = append(args, limit)
	}

	var notices []models.Notice
	err := s.DB.Select(&notices, s.DB.Rebind(query), args...)
	if err != nil {
		return nil, err
	}
	if notices == nil {
		notices = []models.Notice{}
	}
	return notices, nil
}

func (s *NoticeStore) GetBySlug(slug string) (*models.Notice, error) {
	var notice models.Notice
	err := s.DB.Get(&notice, "SELECT * FROM notice WHERE slug = $1 AND is_published = true", slug)
	if err != nil {
		return nil, err
	}
	return &notice, nil
}

func (s *NoticeStore) Create(input models.NoticeInput) (*models.Notice, error) {
	var notice models.Notice
	query := `
		INSERT INTO notice (title, slug, content_html, featured_image_url, is_published)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING *
	`
	err := s.DB.Get(&notice, query, input.Title, input.Slug, input.ContentHTML, input.FeaturedImageURL, input.IsPublished)
	return &notice, err
}

func (s *NoticeStore) Update(id string, input models.NoticeInput) (*models.Notice, error) {
	var notice models.Notice
	query := `
		UPDATE notice 
		SET title = $1, slug = $2, content_html = $3, featured_image_url = $4, is_published = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING *
	`
	err := s.DB.Get(&notice, query, input.Title, input.Slug, input.ContentHTML, input.FeaturedImageURL, input.IsPublished, id)
	return &notice, err
}
