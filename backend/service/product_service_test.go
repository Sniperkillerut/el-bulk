package service

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock objects for services
type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) LogAction(ctx context.Context, action, resourceType, resourceID string, details models.JSONB) {
	m.Called(ctx, action, resourceType, resourceID, details)
}

func (m *MockAuditService) List(ctx context.Context, page, pageSize int, adminID, action, resourceType string) ([]models.AuditLog, int, error) {
	args := m.Called(ctx, page, pageSize, adminID, action, resourceType)
	return args.Get(0).([]models.AuditLog), args.Int(1), args.Error(2)
}

func TestProductService_Create(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	productStore := store.NewProductStore(sqlxDB)
	tcgStore := store.NewTCGStore(sqlxDB)
	
	mockAudit := new(MockAuditService)
	s := NewProductService(productStore, tcgStore, nil, mockAudit)

	t.Run("Create Product Success", func(t *testing.T) {
		input := models.Product{
			ID:    "p-1",
			Name:  "Test Product",
			Price: 100,
		}

		// Mock the store insert
		sqlMock.ExpectQuery("INSERT INTO product").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price"}).AddRow("p-1", "Test Product", 100))

		// Expect Audit Log
		mockAudit.On("LogAction", mock.Anything, "CREATE_PRODUCT", "product", "p-1", mock.Anything).Return()

		prod, err := s.Create(context.Background(), input)
		assert.NoError(t, err)
		assert.Equal(t, "p-1", prod.ID)
		mockAudit.AssertExpectations(t)
	})
}
