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

// MockSettingsService
type MockSettingsService struct {
	mock.Mock
}

func (m *MockSettingsService) GetSettings(ctx context.Context) (models.Settings, error) {
	args := m.Called(ctx)
	return args.Get(0).(models.Settings), args.Error(1)
}

func (m *MockSettingsService) Upsert(ctx context.Context, key, value string) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockSettingsService) InvalidateCache() {
	m.Called()
}

func TestProductService_Create(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	productStore := store.NewProductStore(sqlxDB)
	tcgStore := store.NewTCGStore(sqlxDB)
	
	mockAudit := new(MockAuditService)
	mockSettings := new(MockSettingsService)
	s := NewProductService(productStore, tcgStore, mockSettings, mockAudit)

	t.Run("Create Product Success", func(t *testing.T) {
		mockSettings.On("GetSettings", mock.Anything).Return(models.Settings{}, nil)

		input := models.ProductInput{
			Name:  "Test Product",
		}
		
		input.PriceCOPOverride = new(float64)
		*input.PriceCOPOverride = 100

		// Mock the store insert (assuming the store takes ProductInput)
		sqlMock.ExpectQuery("INSERT INTO product").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("p-1", "Test Product"))

		// Expect Audit Log
		mockAudit.On("LogAction", mock.Anything, "CREATE_PRODUCT", "product", "p-1", mock.Anything).Return()

		prod, err := s.Create(context.Background(), input)
		assert.NoError(t, err)
		assert.Equal(t, "p-1", prod.ID)
		mockAudit.AssertExpectations(t)
	})
}
