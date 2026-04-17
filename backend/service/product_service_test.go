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

func TestProductService_EnrichProducts_Sanitization(t *testing.T) {
	db, _, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "postgres")
	
	productStore := store.NewProductStore(sqlxDB)
	tcgStore := store.NewTCGStore(sqlxDB)
	mockAudit := new(MockAuditService)
	mockSettings := new(MockSettingsService)
	
	s := NewProductService(productStore, tcgStore, mockSettings, mockAudit)
	
	t.Run("Sanitizes fields for non-admin", func(t *testing.T) {
		ref := 10.5
		override := 45000.0
		products := []models.Product{
			{
				ID:               "p-1",
				PriceReference:   &ref,
				PriceCOPOverride: &override,
				PriceSource:      "tcgplayer",
				StoredIn:         []models.StorageLocation{{StorageID: "s1", Name: "Shelf A", Quantity: 5}},
				CartCount:        3,
				Categories: []models.CustomCategory{
					{Name: "ShowMe", ShowBadge: true},
					{Name: "HideMe", ShowBadge: false},
				},
			},
		}

		// CalculatePrices is called inside EnrichProducts
		// IdentifyHotNew is also called
		mockSettings.On("GetSettings", mock.Anything).Return(models.Settings{}, nil)

		// Note: we don't mock Translate or PopulateCartCounts queries here because 
		// they will just fail silently or return empty, which is fine for this unit test 
		// that focuses on the sanitization loop at the end.
		
		err := s.EnrichProducts(context.Background(), products, models.Settings{}, false)
		assert.NoError(t, err)

		p := products[0]
		assert.Nil(t, p.PriceReference, "PriceReference should be sanitized")
		assert.Nil(t, p.PriceCOPOverride, "PriceCOPOverride should be sanitized")
		assert.Equal(t, models.PriceSource(""), p.PriceSource, "PriceSource should be empty")
		assert.Nil(t, p.StoredIn, "StoredIn should be nil")
		assert.Equal(t, 3, p.CartCount, "CartCount should be preserved")
		
		// Phase 2 Metadata Check
		assert.Nil(t, p.CreatedAt, "CreatedAt should be sanitized")
		assert.Nil(t, p.UpdatedAt, "UpdatedAt should be sanitized")

		assert.Len(t, p.Categories, 1)
		cat := p.Categories[0]
		assert.Equal(t, "ShowMe", cat.Name)
		assert.Nil(t, cat.CreatedAt, "Category CreatedAt should be sanitized")
		assert.Equal(t, 0, cat.ItemCount, "Category ItemCount should be 0")
		assert.False(t, cat.IsHot, "Category IsHot should be false")
		assert.False(t, cat.IsNew, "Category IsNew should be false")
	})

	t.Run("Preserves fields for admin", func(t *testing.T) {
		ref := 10.5
		products := []models.Product{
			{
				ID:             "p-1",
				PriceReference: &ref,
				StoredIn:       []models.StorageLocation{{StorageID: "s1", Quantity: 5}},
				Categories:     []models.CustomCategory{{Name: "Both", ShowBadge: false}},
			},
		}

		mockSettings.On("GetSettings", mock.Anything).Return(models.Settings{}, nil)

		err := s.EnrichProducts(context.Background(), products, models.Settings{}, true)
		assert.NoError(t, err)

		p := products[0]
		assert.NotNil(t, p.PriceReference)
		assert.NotNil(t, p.StoredIn)
		assert.Len(t, p.Categories, 1) // Admin sees all regardless of ShowBadge? 
		// Wait, the logic only FILTERS for !isAdmin.
	})
}

