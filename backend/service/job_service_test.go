package service

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestJobService_CreateJob(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	js := NewJobService(store.NewJobStore(sqlxDB))

	mock.ExpectQuery("INSERT INTO job").
		WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(time.Now()))

	job, err := js.CreateJob(context.Background(), "test_type", nil, models.JSONB{"key": "value"})
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "test_type", job.Type)
	assert.Equal(t, models.JobPending, job.Status)
}

func TestJobService_UpdateStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	js := NewJobService(store.NewJobStore(sqlxDB))

	jobID := "job-1"
	mock.ExpectExec("UPDATE job").
		WithArgs(jobID, models.JobCompleted, 100, sqlmock.AnyArg(), nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = js.UpdateStatus(context.Background(), jobID, models.JobCompleted, 100, models.JSONB{"result": "ok"}, nil)
	assert.NoError(t, err)
}
