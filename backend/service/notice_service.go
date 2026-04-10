package service

import (
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

func (s *NoticeService) List(isPublished bool, limit int) ([]models.Notice, error) {
	logger.Trace("Entering NoticeService.List | isPublished: %v | limit: %d", isPublished, limit)
	return s.Store.List(isPublished, limit)
}

func (s *NoticeService) GetBySlug(slug string) (*models.Notice, error) {
	logger.Trace("Entering NoticeService.GetBySlug | Slug: %s", slug)
	return s.Store.GetBySlug(slug)
}

func (s *NoticeService) Create(input models.NoticeInput) (*models.Notice, error) {
	logger.Trace("Entering NoticeService.Create | Title: %s", input.Title)
	return s.Store.Create(input)
}

func (s *NoticeService) Update(id string, input models.NoticeInput) (*models.Notice, error) {
	logger.Trace("Entering NoticeService.Update | ID: %s", id)
	return s.Store.Update(id, input)
}

func (s *NoticeService) Delete(id string) error {
	logger.Trace("Entering NoticeService.Delete | ID: %s", id)
	return s.Store.BaseStore.Delete(id)
}
