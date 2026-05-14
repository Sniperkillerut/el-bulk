package service

import (
	"context"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/google/uuid"
)

type JobService struct {
	Store *store.JobStore
}

func NewJobService(s *store.JobStore) *JobService {
	return &JobService{Store: s}
}

func (s *JobService) CreateJob(ctx context.Context, jobType string, adminID *string, payload models.JSONB) (*models.Job, error) {
	job := &models.Job{
		ID:       uuid.New().String(),
		Type:     jobType,
		Status:   models.JobPending,
		Progress: 0,
		Payload:  payload,
		AdminID:  adminID,
	}

	if err := s.Store.Create(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *JobService) GetJob(ctx context.Context, id string) (*models.Job, error) {
	return s.Store.Get(ctx, id)
}

func (s *JobService) UpdateStatus(ctx context.Context, id string, status models.JobStatus, progress int, result models.JSONB, err error) error {
	var errStr *string
	if err != nil {
		s := err.Error()
		errStr = &s
	}
	return s.Store.UpdateStatus(ctx, id, status, progress, result, errStr)
}

func (s *JobService) ListRecent(ctx context.Context, limit int) ([]models.Job, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.Store.ListRecent(ctx, limit)
}
