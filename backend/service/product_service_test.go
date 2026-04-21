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

		// Mock the store insert (Smart matching via BulkUpsert)
		sqlMock.ExpectQuery(`SELECT upserted_id FROM fn_bulk_upsert_product`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"upserted_id"}).AddRow("p-1"))

		// Mock GetByID querying product detail
		sqlMock.ExpectQuery(`SELECT fn_get_product_detail`).
			WithArgs("p-1").
			WillReturnRows(sqlmock.NewRows([]string{"fn_get_product_detail"}).AddRow(`{"id":"p-1","name":"Test Product"}`))

		// EnrichProducts mocks:
		// 1. PopulateStorage
		sqlMock.ExpectQuery(`SELECT ps.product_id, s.id as stored_in_id, s.name, ps.quantity`).
			WithArgs("p-1").
			WillReturnRows(sqlmock.NewRows([]string{"product_id", "stored_in_id", "name", "quantity"}))

		// 2. PopulateCategories
		sqlMock.ExpectQuery(`SELECT pc.product_id, c.id, c.name, c.slug, c.show_badge, c.is_active, c.searchable, c.bg_color, c.text_color, c.icon`).
			WithArgs("p-1").
			WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name", "slug", "show_badge", "is_active", "searchable", "bg_color", "text_color", "icon"}))

		// 3. PopulateCartCounts
		sqlMock.ExpectQuery(`SELECT oi.product_id, COUNT\(DISTINCT o.customer_id\) as cart_count`).
			WithArgs("p-1").
			WillReturnRows(sqlmock.NewRows([]string{"product_id", "cart_count"}))

		// 4. IdentifyHotNew -> GetHotProductIDs
		sqlMock.ExpectQuery(`SELECT product_id FROM order_item oi JOIN "order" o ON oi.order_id = o.id WHERE o.created_at > \(now\(\) - interval '7 days'\)`).
			WithArgs("p-1").
			WillReturnRows(sqlmock.NewRows([]string{"product_id"}))

		// Expect Audit Log
		mockAudit.On("LogAction", mock.Anything, "CREATE_PRODUCT", "product", "p-1", mock.Anything).Return()

		prod, err := s.Create(context.Background(), input)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		assert.Equal(t, "p-1", prod.ID)
		mockAudit.AssertExpectations(t)
	})
}

func TestProductService_BulkOperations(t *testing.T) {
	t.Run("BulkCreate Success", func(t *testing.T) {
		db, sqlMock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		productStore := store.NewProductStore(sqlxDB)
		s := NewProductService(productStore, nil, nil, nil)

		inputs := []models.ProductInput{{Name: "P1"}, {Name: "P2"}}
		sqlMock.ExpectQuery(`SELECT upserted_id FROM fn_bulk_upsert_product`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"upserted_id"}).AddRow("p-1").AddRow("p-2"))

		count, err := s.BulkCreate(context.Background(), inputs, nil)
		assert.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("BulkSearch", func(t *testing.T) {
		db, sqlMock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		productStore := store.NewProductStore(sqlxDB)
		mockSettings := new(MockSettingsService)
		s := NewProductService(productStore, nil, mockSettings, nil)

		mockSettings.On("GetSettings", mock.Anything).Return(models.Settings{}, nil)
		
		list := "1x Card A\n2x Card B"
		
		// Search for Card A
		sqlMock.ExpectQuery(`SELECT \* FROM product WHERE \(LOWER\(name\) = LOWER\(\$1\) OR name ILIKE \$1\) AND stock > 0 ORDER BY stock DESC LIMIT 5`).
			WithArgs("Card A").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("p-a", "Card A"))

		// Enrich Card A mocks
		sqlMock.ExpectQuery(`SELECT ps\.product_id, s\.id as stored_in_id, s\.name, ps\.quantity`).WithArgs("p-a").WillReturnRows(sqlmock.NewRows([]string{"product_id", "stored_in_id", "name", "quantity"}))
		sqlMock.ExpectQuery(`SELECT pc\.product_id, c\.id, c\.name`).WithArgs("p-a").WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name"}))
		sqlMock.ExpectQuery(`SELECT oi\.product_id, COUNT`).WithArgs("p-a").WillReturnRows(sqlmock.NewRows([]string{"product_id", "cart_count"}))
		sqlMock.ExpectQuery(`SELECT product_id FROM order_item`).WithArgs("p-a").WillReturnRows(sqlmock.NewRows([]string{"product_id"}))

		// Search for Card B
		sqlMock.ExpectQuery(`SELECT \* FROM product WHERE \(LOWER\(name\) = LOWER\(\$1\) OR name ILIKE \$1\) AND stock > 0 ORDER BY stock DESC LIMIT 5`).
			WithArgs("Card B").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"})) // No matches for B

		results, err := s.BulkSearch(context.Background(), list)
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.True(t, results[0].IsMatched)
		assert.False(t, results[1].IsMatched)
	})

	t.Run("BulkUpdateSource", func(t *testing.T) {
		db, sqlMock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		sqlxDB := sqlx.NewDb(db, "postgres")
		productStore := store.NewProductStore(sqlxDB)
		mockAudit := new(MockAuditService)
		mockSettings := new(MockSettingsService)
		s := NewProductService(productStore, nil, mockSettings, mockAudit)

		ids := []string{"p-1", "p-2"}
		source := models.PriceSourceTCGPlayer
		
		sqlMock.ExpectExec(`UPDATE product SET price_source = \$1`).
			WithArgs(source, "p-1", "p-2").
			WillReturnResult(sqlmock.NewResult(0, 2))

		sqlMock.ExpectQuery(`SELECT p\.id, p\.tcg, p\.name`).
			WithArgs("p-1", "p-2").
			WillReturnRows(sqlmock.NewRows([]string{"id", "tcg", "name", "set_name", "set_code", "collector_number", "foil_treatment", "card_treatment", "price_source", "scryfall_id", "ck_set_name"}).
				AddRow("p-1", "mtg", "Card 1", nil, nil, "", "non_foil", "normal", "tcgplayer", "", nil).
				AddRow("p-2", "mtg", "Card 2", nil, nil, "", "non_foil", "normal", "tcgplayer", "", nil))

		// BulkUpdateMetadata mocks inside BulkUpdateSource
		mockSettings.On("GetSettings", mock.Anything).Return(models.Settings{}, nil)

		// Audit
		mockAudit.On("LogAction", mock.Anything, "BULK_UPDATE_SOURCE", "product", "batch", mock.Anything).Return()

		count, err := s.BulkUpdateSource(context.Background(), ids, source, nil)
		assert.NoError(t, err)
		assert.Equal(t, 2, count)
	})
}

func TestProductService_EnrichProducts_Sanitization(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
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

		// Mock side-data calls inside EnrichProducts
		sqlMock.ExpectQuery(`SELECT ps\.product_id, s\.id as stored_in_id, s\.name, ps\.quantity`).WithArgs("p-1").WillReturnRows(sqlmock.NewRows([]string{"product_id", "stored_in_id", "name", "quantity"}))
		sqlMock.ExpectQuery(`SELECT pc\.product_id, c\.id, c\.name, c\.slug, c\.show_badge, c\.is_active, c\.searchable, c\.bg_color, c\.text_color, c\.icon`).
			WithArgs("p-1").
			WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name", "slug", "show_badge", "is_active", "searchable", "bg_color", "text_color", "icon"}).
				AddRow("p-1", "c1", "ShowMe", "showme", true, true, true, nil, nil, nil).
				AddRow("p-1", "c2", "HideMe", "hideme", false, true, true, nil, nil, nil))
		sqlMock.ExpectQuery(`SELECT oi\.product_id, COUNT\(DISTINCT o\.customer_id\) as cart_count`).
			WithArgs("p-1").
			WillReturnRows(sqlmock.NewRows([]string{"product_id", "cart_count"}).AddRow("p-1", 3))
		sqlMock.ExpectQuery(`SELECT product_id FROM order_item oi JOIN "order" o ON oi\.order_id = o\.id WHERE o\.created_at > \(now\(\) - interval '7 days'\)`).
			WithArgs("p-1").
			WillReturnRows(sqlmock.NewRows([]string{"product_id"}))

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

		// Mock side-data calls inside EnrichProducts
		sqlMock.ExpectQuery(`SELECT ps\.product_id, s\.id as stored_in_id, s\.name, ps\.quantity`).WithArgs("p-1").WillReturnRows(sqlmock.NewRows([]string{"product_id", "stored_in_id", "name", "quantity"}))
		sqlMock.ExpectQuery(`SELECT pc\.product_id, c\.id, c\.name, c\.slug, c\.show_badge, c\.is_active, c\.searchable, c\.bg_color, c\.text_color, c\.icon`).
			WithArgs("p-1").
			WillReturnRows(sqlmock.NewRows([]string{"product_id", "id", "name", "slug", "show_badge", "is_active", "searchable", "bg_color", "text_color", "icon"}).
				AddRow("p-1", "c1", "Both", "both", false, true, true, nil, nil, nil))
		sqlMock.ExpectQuery(`SELECT oi\.product_id, COUNT\(DISTINCT o\.customer_id\) as cart_count`).
			WithArgs("p-1").
			WillReturnRows(sqlmock.NewRows([]string{"product_id", "cart_count"}))
		sqlMock.ExpectQuery(`SELECT product_id FROM order_item oi JOIN "order" o ON oi\.order_id = o\.id WHERE o\.created_at > \(now\(\) - interval '7 days'\)`).
			WithArgs("p-1").
			WillReturnRows(sqlmock.NewRows([]string{"product_id"}))

		err := s.EnrichProducts(context.Background(), products, models.Settings{}, true)
		assert.NoError(t, err)

		p := products[0]
		assert.NotNil(t, p.PriceReference)
		assert.NotNil(t, p.StoredIn)
		assert.Len(t, p.Categories, 1) // Admin sees all regardless of ShowBadge? 
		// Wait, the logic only FILTERS for !isAdmin.
	})
}

