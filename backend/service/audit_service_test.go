package service

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestAuditService_LogAction(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	auditStore := store.NewAuditStore(sqlxDB)
	adminStore := store.NewAdminStore(sqlxDB)
	s := NewAuditService(auditStore, adminStore)

	t.Run("Record with Admin Context", func(t *testing.T) {
		adminID := "admin-123"
		ctx := context.WithValue(context.Background(), middleware.AdminContextKey, adminID)

		// Mock Admin lookup
		mock.ExpectQuery("SELECT \\* FROM admin WHERE id = \\$1").
			WithArgs(adminID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(adminID, "test-admin"))

		// Mock Audit Creation
		mock.ExpectExec("INSERT INTO admin_audit_log").
			WithArgs(adminID, "test-admin", "TEST_ACTION", "test_resource", "res-456", models.JSONB{"foo": "bar"}, nil).
			WillReturnResult(sqlmock.NewResult(1, 1))

		s.LogAction(ctx, "TEST_ACTION", "test_resource", "res-456", models.JSONB{"foo": "bar"})
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Record without Admin Context (System Fallback)", func(t *testing.T) {
		ctx := context.Background()

		// Mock Audit Creation (no admin lookup)
		mock.ExpectExec("INSERT INTO admin_audit_log").
			WithArgs(nil, "system", "SYSTEM_ACTION", "system_res", "res-000", nil, nil).
			WillReturnResult(sqlmock.NewResult(1, 1))

		s.LogAction(ctx, "SYSTEM_ACTION", "system_res", "res-000", nil)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Handle Admin Lookup Error", func(t *testing.T) {
		adminID := "admin-err"
		ctx := context.WithValue(context.Background(), middleware.AdminContextKey, adminID)

		// Mock Admin lookup failure
		mock.ExpectQuery("SELECT \\* FROM admin WHERE id = \\$1").
			WithArgs(adminID).
			WillReturnError(errors.New("db error"))

		// Mock Audit Creation (fallback to system)
		mock.ExpectExec("INSERT INTO admin_audit_log").
			WithArgs(adminID, "system", "ERROR_ACTION", "err_res", "res-err", nil, nil).
			WillReturnResult(sqlmock.NewResult(1, 1))

		s.LogAction(ctx, "ERROR_ACTION", "err_res", "res-err", nil)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
