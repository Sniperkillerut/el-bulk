package handlers

import (
	"testing"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestConstructors(t *testing.T) {
	db := &sqlx.DB{}
	assert.NotNil(t, NewAdminHandler(db))
	assert.NotNil(t, NewCategoriesHandler(db))
	assert.NotNil(t, NewOrderHandler(db))
	assert.NotNil(t, NewProductHandler(db))
	assert.NotNil(t, NewRefreshHandler(db))
	assert.NotNil(t, NewSettingsHandler(db))
	assert.NotNil(t, NewStorageHandler(db))
	assert.NotNil(t, NewTCGHandler(db))
}
