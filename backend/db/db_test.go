package db

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnectResilient_NoURL(t *testing.T) {
	// Clear DATABASE_URL
	oldURL := os.Getenv("DATABASE_URL")
	os.Unsetenv("DATABASE_URL")
	defer os.Setenv("DATABASE_URL", oldURL)

	db, err := ConnectResilient()
	assert.Nil(t, db)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL environment variable is required")
}

func TestConnectResilient_InvalidURL(t *testing.T) {
	// Set an invalid URL that will fail on connection (but not necessarily on parse)
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/dbname?sslmode=disable")
	defer os.Unsetenv("DATABASE_URL")

	// We don't expect it to succeed if no DB is running there
	db, err := ConnectResilient()
	assert.Nil(t, db)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
}
