package service

import (
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
)

type NoticeService struct {
	Store *store.NoticeStore
}

func NewNoticeService(s *store.NoticeStore) *NoticeService {
	return &NoticeService{Store: s}
}

func (s *NoticeService) List(isPublished bool, limit int) ([]models.Notice, error) {
	return s.Store.List(isPublished, limit)
}

func (s *NoticeService) GetBySlug(slug string) (*models.Notice, error) {
	return s.Store.GetBySlug(slug)
}

func (s *NoticeService) Create(input models.NoticeInput) (*models.Notice, error) {
	return s.Store.Create(input)
}

func (s *NoticeService) Update(id string, input models.NoticeInput) (*models.Notice, error) {
	return s.Store.Update(id, input)
}

func (s *NoticeService) Delete(id string) error {
	return s.Store.BaseStore.Delete(id)
}
