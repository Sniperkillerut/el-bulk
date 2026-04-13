package store

import (
	"context"
	"time"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
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

func (s *NoticeStore) List(ctx context.Context, isPublished bool, limit int) ([]models.Notice, error) {
	start := time.Now()
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

	rebound := s.DB.Rebind(query)
	logger.TraceCtx(ctx, "[DB] Executing List (Notice): %s | Args: %+v", rebound, args)
	var notices []models.Notice
	err := s.DB.SelectContext(ctx, &notices, rebound, args...)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] List (Notice) failed: %v", err)
		return nil, err
	}
	if notices == nil {
		notices = []models.Notice{}
	}
	logger.DebugCtx(ctx, "[DB] List (Notice) took %v", time.Since(start))
	return notices, nil
}

func (s *NoticeStore) GetBySlug(ctx context.Context, slug string) (*models.Notice, error) {
	var notice models.Notice
	err := s.DB.GetContext(ctx, &notice, "SELECT * FROM notice WHERE slug = $1 AND is_published = true", slug)
	if err != nil {
		return nil, err
	}
	return &notice, nil
}

func (s *NoticeStore) Create(ctx context.Context, input models.NoticeInput) (*models.Notice, error) {
	var notice models.Notice
	query := `
		INSERT INTO notice (title, slug, content_html, featured_image_url, is_published)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING *
	`
	logger.TraceCtx(ctx, "[DB] Executing CreateNotice: %s | Title: %s", query, input.Title)
	err := s.DB.GetContext(ctx, &notice, query, input.Title, input.Slug, input.ContentHTML, input.FeaturedImageURL, input.IsPublished)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] CreateNotice failed for %s: %v", input.Title, err)
	}
	return &notice, err
}

func (s *NoticeStore) Update(ctx context.Context, id string, input models.NoticeInput) (*models.Notice, error) {
	var notice models.Notice
	query := `
		UPDATE notice 
		SET title = $1, slug = $2, content_html = $3, featured_image_url = $4, is_published = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING *
	`
	err := s.DB.GetContext(ctx, &notice, query, input.Title, input.Slug, input.ContentHTML, input.FeaturedImageURL, input.IsPublished, id)
	return &notice, err
}
