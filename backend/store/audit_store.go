package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
)

type AuditStore struct {
	DB *sqlx.DB
}

func NewAuditStore(db *sqlx.DB) *AuditStore {
	return &AuditStore{DB: db}
}

func (s *AuditStore) Create(ctx context.Context, log models.AuditLog) error {
	query := `
		INSERT INTO admin_audit_log (admin_id, admin_username, action, resource_type, resource_id, details, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := s.DB.ExecContext(ctx, query, log.AdminID, log.AdminUsername, log.Action, log.ResourceType, log.ResourceID, log.Details, log.IPAddress)
	return err
}

func (s *AuditStore) List(ctx context.Context, page, pageSize int, adminID, action, resourceType string) ([]models.AuditLog, int, error) {
	var logs []models.AuditLog
	var total int

	where := []string{}
	args := []interface{}{}
	argIdx := 1

	if adminID != "" {
		where = append(where, fmt.Sprintf("admin_id = $%d", argIdx))
		args = append(args, adminID)
		argIdx++
	}
	if action != "" {
		where = append(where, fmt.Sprintf("action = $%d", argIdx))
		args = append(args, action)
		argIdx++
	}
	if resourceType != "" {
		where = append(where, fmt.Sprintf("resource_type = $%d", argIdx))
		args = append(args, resourceType)
		argIdx++
	}

	whereStr := ""
	if len(where) > 0 {
		whereStr = "WHERE " + strings.Join(where, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM admin_audit_log %s", whereStr)
	if err := s.DB.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	limitArgs := append(args, pageSize, offset)
	listQuery := fmt.Sprintf(`
		SELECT * FROM admin_audit_log 
		%s 
		ORDER BY created_at DESC 
		LIMIT $%d OFFSET $%d
	`, whereStr, argIdx, argIdx+1)

	if err := s.DB.SelectContext(ctx, &logs, listQuery, limitArgs...); err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
