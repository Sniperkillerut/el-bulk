package service

import (
	"context"

	"github.com/el-bulk/backend/middleware"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/el-bulk/backend/utils/logger"
)

type AuditService struct {
	Store      *store.AuditStore
	AdminStore *store.AdminStore
}

func NewAuditService(s *store.AuditStore, as *store.AdminStore) *AuditService {
	return &AuditService{
		Store:      s,
		AdminStore: as,
	}
}

// LogAction records an administrative action.
// It extracts the adminID from the context (provided by AdminAuth middleware).
func (s *AuditService) LogAction(ctx context.Context, action, resourceType, resourceID string, details models.JSONB) {
	if s == nil || s.Store == nil {
		return
	}
	adminID, _ := ctx.Value(middleware.AdminContextKey).(string)
	
	username := "system"
	if adminID != "" && s.AdminStore != nil {
		admin, err := s.AdminStore.GetByID(ctx, adminID)
		if err == nil && admin != nil {
			username = admin.Username
		}
	}

	log := models.AuditLog{
		AdminUsername: username,
		Action:        action,
		ResourceType:  resourceType,
		ResourceID:    resourceID,
		Details:       details,
	}
	
	if adminID != "" {
		log.AdminID = &adminID
	}

	if err := s.Store.Create(ctx, log); err != nil {
		logger.ErrorCtx(ctx, "Failed to record audit log: %v", err)
	}
}

func (s *AuditService) List(ctx context.Context, page, pageSize int, adminID, action, resourceType string) ([]models.AuditLog, int, error) {
	if s == nil || s.Store == nil {
		return nil, 0, nil
	}
	return s.Store.List(ctx, page, pageSize, adminID, action, resourceType)
}
