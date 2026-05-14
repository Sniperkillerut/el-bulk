package service

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestWorkerPool(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	js := NewJobService(store.NewJobStore(sqlxDB))
	pool := NewWorkerPool(js, 2)

	var processedCount int32
	pool.RegisterHandler("test_job", func(ctx context.Context, job *models.Job, updateProgress func(int)) (models.JSONB, error) {
		updateProgress(50)
		atomic.AddInt32(&processedCount, 1)
		return models.JSONB{"processed": true}, nil
	})

	pool.Start()
	defer pool.Stop()

	job := &models.Job{
		ID:   "job-1",
		Type: "test_job",
	}

	// Expectations for processJob:
	// 1. Update to 'running'
	mock.ExpectExec("UPDATE job").WithArgs("job-1", models.JobRunning, 0, nil, nil).WillReturnResult(sqlmock.NewResult(1, 1))
	// 2. Update progress (50)
	mock.ExpectExec("UPDATE job SET progress = \\$2 WHERE id = \\$1").WithArgs("job-1", 50).WillReturnResult(sqlmock.NewResult(1, 1))
	// 3. Update to 'completed'
	mock.ExpectExec("UPDATE job").WithArgs("job-1", models.JobCompleted, 100, sqlmock.AnyArg(), nil).WillReturnResult(sqlmock.NewResult(1, 1))

	pool.Submit(job)

	// Wait for processing
	for i := 0; i < 10; i++ {
		if atomic.LoadInt32(&processedCount) > 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	assert.Equal(t, int32(1), atomic.LoadInt32(&processedCount))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkerPool_NoHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	js := NewJobService(store.NewJobStore(sqlxDB))
	pool := NewWorkerPool(js, 1)

	pool.Start()
	defer pool.Stop()

	job := &models.Job{
		ID:   "job-fail",
		Type: "unknown",
	}

	// Expectations:
	// 1. Update to 'failed' because no handler
	mock.ExpectExec("UPDATE job").WithArgs("job-fail", models.JobFailed, 0, nil, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))

	pool.Submit(job)

	// Give it a moment to fail
	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkerPool_HandlerError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	js := NewJobService(store.NewJobStore(sqlxDB))
	pool := NewWorkerPool(js, 1)

	pool.RegisterHandler("error_job", func(ctx context.Context, job *models.Job, updateProgress func(int)) (models.JSONB, error) {
		return nil, fmt.Errorf("boom")
	})

	pool.Start()
	defer pool.Stop()

	job := &models.Job{
		ID:   "job-err",
		Type: "error_job",
	}

	mock.ExpectExec("UPDATE job").WithArgs("job-err", models.JobRunning, 0, nil, nil).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE job").WithArgs("job-err", models.JobFailed, 0, nil, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))

	pool.Submit(job)

	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, mock.ExpectationsWereMet())
}
