package db

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestConnectResilient_Failures(t *testing.T) {
	t.Run("Missing Env", func(t *testing.T) {
		old := os.Getenv("DATABASE_URL")
		os.Setenv("DATABASE_URL", "")
		defer os.Setenv("DATABASE_URL", old)

		_, err := ConnectResilient()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("Connection Fail", func(t *testing.T) {
		os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:9999/db?sslmode=disable")
		_, err := ConnectResilient()
		assert.Error(t, err)
	})
}
