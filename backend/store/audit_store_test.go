package store

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestAuditStore_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	s := NewAuditStore(sqlxDB)

	adminID := "admin-1"
	log := models.AuditLog{
		AdminID:       &adminID,
		AdminUsername: "adminuser",
		Action:        "CREATE",
		ResourceType:  "product",
		ResourceID:    "prod-123",
	}

	mock.ExpectExec("INSERT INTO admin_audit_log").
		WithArgs(log.AdminID, log.AdminUsername, log.Action, log.ResourceType, log.ResourceID, log.Details, log.IPAddress).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = s.Create(context.Background(), log)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuditStore_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	s := NewAuditStore(sqlxDB)

	t.Run("List with Filters", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "action", "resource_type"}).
			AddRow("log-1", "UPDATE", "order").
			AddRow("log-2", "UPDATE", "order")

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM admin_audit_log").
			WithArgs("UPDATE", "order").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		mock.ExpectQuery("SELECT \\* FROM admin_audit_log").
			WithArgs("UPDATE", "order", 20, 0).
			WillReturnRows(rows)

		logs, total, err := s.List(context.Background(), 1, 20, "", "UPDATE", "order")
		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, logs, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
