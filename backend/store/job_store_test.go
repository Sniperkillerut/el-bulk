package store

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestJobStore_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	s := NewJobStore(sqlxDB)

	job := &models.Job{
		ID:     "job-1",
		Type:   "test_job",
		Status: models.JobPending,
	}

	mock.ExpectQuery("INSERT INTO job").
		WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(time.Now()))

	err = s.Create(context.Background(), job)
	assert.NoError(t, err)
	assert.NotNil(t, job.CreatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestJobStore_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	s := NewJobStore(sqlxDB)

	jobID := "job-1"
	rows := sqlmock.NewRows([]string{"id", "type", "status"}).
		AddRow(jobID, "test_job", "pending")

	mock.ExpectQuery("SELECT \\* FROM job WHERE id = \\$1").
		WithArgs(jobID).
		WillReturnRows(rows)

	job, err := s.Get(context.Background(), jobID)
	assert.NoError(t, err)
	assert.Equal(t, jobID, job.ID)
	assert.Equal(t, models.JobPending, job.Status)
}

func TestJobStore_UpdateStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	s := NewJobStore(sqlxDB)

	jobID := "job-1"
	status := models.JobRunning
	progress := 50

	mock.ExpectExec("UPDATE job").
		WithArgs(jobID, status, progress, nil, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = s.UpdateStatus(context.Background(), jobID, status, progress, nil, nil)
	assert.NoError(t, err)
}

func TestJobStore_ListRecent(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	s := NewJobStore(sqlxDB)

	rows := sqlmock.NewRows([]string{"id", "type"}).
		AddRow("job-1", "type-1").
		AddRow("job-2", "type-2")

	mock.ExpectQuery("SELECT \\* FROM job ORDER BY created_at DESC LIMIT \\$1").
		WithArgs(10).
		WillReturnRows(rows)

	jobs, err := s.ListRecent(context.Background(), 10)
	assert.NoError(t, err)
	assert.Len(t, jobs, 2)
}
