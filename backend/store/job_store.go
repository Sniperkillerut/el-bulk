package store

import (
	"context"
	"fmt"

	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
)

type JobStore struct {
	DB *sqlx.DB
}

func NewJobStore(db *sqlx.DB) *JobStore {
	return &JobStore{DB: db}
}

func (s *JobStore) Create(ctx context.Context, job *models.Job) error {
	query := `
		INSERT INTO job (id, type, status, progress, payload, admin_id)
		VALUES (:id, :type, :status, :progress, :payload, :admin_id)
		RETURNING created_at
	`
	rows, err := s.DB.NamedQueryContext(ctx, query, job)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&job.CreatedAt); err != nil {
			return err
		}
	}
	return nil
}

func (s *JobStore) Get(ctx context.Context, id string) (*models.Job, error) {
	var job models.Job
	err := s.DB.GetContext(ctx, &job, "SELECT * FROM job WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *JobStore) UpdateStatus(ctx context.Context, id string, status models.JobStatus, progress int, result models.JSONB, errStr *string) error {
	query := `
		UPDATE job 
		SET status = $2, progress = $3, result = $4, error = $5,
		    started_at = CASE WHEN $2 = 'running' AND started_at IS NULL THEN now() ELSE started_at END,
		    completed_at = CASE WHEN $2 IN ('completed', 'failed') THEN now() ELSE completed_at END
		WHERE id = $1
	`
	_, err := s.DB.ExecContext(ctx, query, id, status, progress, result, errStr)
	return err
}

func (s *JobStore) UpdateProgress(ctx context.Context, id string, progress int) error {
	_, err := s.DB.ExecContext(ctx, "UPDATE job SET progress = $2 WHERE id = $1", id, progress)
	return err
}

func (s *JobStore) ListRecent(ctx context.Context, limit int) ([]models.Job, error) {
	var jobs []models.Job
	err := s.DB.SelectContext(ctx, &jobs, "SELECT * FROM job ORDER BY created_at DESC LIMIT $1", limit)
	return jobs, err
}
