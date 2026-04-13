package service

import (
	"context"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/el-bulk/backend/utils/logger"
)

type NoticeService struct {
	Store *store.NoticeStore
}

func NewNoticeService(s *store.NoticeStore) *NoticeService {
	return &NoticeService{Store: s}
}

func (s *NoticeService) List(ctx context.Context, isPublished bool, limit int) ([]models.Notice, error) {
	logger.TraceCtx(ctx, "Entering NoticeService.List | isPublished: %v | limit: %d", isPublished, limit)
	return s.Store.List(ctx, isPublished, limit)
}

func (s *NoticeService) GetBySlug(ctx context.Context, slug string) (*models.Notice, error) {
	logger.TraceCtx(ctx, "Entering NoticeService.GetBySlug | Slug: %s", slug)
	return s.Store.GetBySlug(ctx, slug)
}

func (s *NoticeService) Create(ctx context.Context, input models.NoticeInput) (*models.Notice, error) {
	logger.TraceCtx(ctx, "Entering NoticeService.Create | Title: %s", input.Title)
	return s.Store.Create(ctx, input)
}

func (s *NoticeService) Update(ctx context.Context, id string, input models.NoticeInput) (*models.Notice, error) {
	logger.TraceCtx(ctx, "Entering NoticeService.Update | ID: %s", id)
	return s.Store.Update(ctx, id, input)
}

func (s *NoticeService) Delete(ctx context.Context, id string) error {
	logger.TraceCtx(ctx, "Entering NoticeService.Delete | ID: %s", id)
	return s.Store.BaseStore.Delete(ctx, id)
}
