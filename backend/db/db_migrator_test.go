package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestMigrate(t *testing.T) {
	// Create a temporary directory structure: db/migrations
	tempDir := t.TempDir()
	migrationsDir := filepath.Join(tempDir, "db", "migrations")
	err := os.MkdirAll(migrationsDir, 0755)
	assert.NoError(t, err)

	// Create a dummy migration file
	migrationFile := "20260101_test.sql"
	migrationPath := filepath.Join(migrationsDir, migrationFile)
	err = os.WriteFile(migrationPath, []byte("CREATE TABLE test (id int)"), 0644)
	assert.NoError(t, err)

	// Change current working directory to tempDir
	oldCwd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldCwd)

	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// Expect query for applied migrations
	mock.ExpectQuery("SELECT name FROM migration").WillReturnRows(sqlmock.NewRows([]string{"name"}))

	// Expect migration transaction
	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE test").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO migration").WithArgs(migrationFile).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = Migrate(sqlxDB)
	assert.NoError(t, err)
}

func TestMigrate_DBError(t *testing.T) {
	tempDir := t.TempDir()
	migrationsDir := filepath.Join(tempDir, "db", "migrations")
	os.MkdirAll(migrationsDir, 0755)

	oldCwd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldCwd)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectQuery("SELECT name FROM migration").WillReturnError(assert.AnError)

	err = Migrate(sqlxDB)
	assert.Error(t, err)
}

func TestMigrate_NoMigrationsDir(t *testing.T) {
	tempDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldCwd)

	db, _, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	err := Migrate(sqlxDB)
	assert.NoError(t, err) // Should skip silently
}

func TestInitialize(t *testing.T) {
	// Create a temporary directory structure: db/schema
	tempDir := t.TempDir()
	schemaDir := filepath.Join(tempDir, "db", "schema")
	err := os.MkdirAll(schemaDir, 0755)
	assert.NoError(t, err)

	// Create init.sql and component.sql
	initFile := "init.sql"
	compFile := "component.sql"
	err = os.WriteFile(filepath.Join(schemaDir, initFile), []byte("\\i component.sql"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(schemaDir, compFile), []byte("CREATE TABLE schema_test (id int)"), 0644)
	assert.NoError(t, err)

	// Change current working directory to tempDir
	oldCwd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldCwd)

	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// Expect schema execution
	mock.ExpectExec("CREATE TABLE schema_test").WillReturnResult(sqlmock.NewResult(1, 1))

	err = Initialize(sqlxDB)
	assert.NoError(t, err)
}

func TestInitialize_NoSchemaDir(t *testing.T) {
	tempDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldCwd)

	db, _, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	err := Initialize(sqlxDB)
	assert.NoError(t, err) // Should skip silently
}

func TestInitialize_InvalidPath(t *testing.T) {
	tempDir := t.TempDir()
	schemaDir := filepath.Join(tempDir, "db", "schema")
	os.MkdirAll(schemaDir, 0755)
	os.WriteFile(filepath.Join(schemaDir, "init.sql"), []byte("\\i ../outside.sql"), 0644)

	oldCwd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldCwd)

	db, _, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	err := Initialize(sqlxDB)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid path")
}
