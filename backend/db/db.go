package db

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/el-bulk/backend/utils/logger"
)

func init() {
	// Walk up from cwd until we find a .env file (works regardless of where the
	// binary is invoked from). Silently skip if not found — Docker supplies vars
	// via docker-compose environment instead.
	dir, _ := os.Getwd()
	for {
		candidate := filepath.Join(dir, ".env")
		if _, err := os.Stat(candidate); err == nil {
			_ = godotenv.Load(candidate)
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached filesystem root, no .env found
		}
		dir = parent
	}
}

func Connect() *sqlx.DB {
	db, err := ConnectResilient()
	if err != nil {
		logger.Error("Database connection failed: %v", err)
		os.Exit(1)
	}
	return db
}

func ConnectResilient() (*sqlx.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	logger.Info("Database connected successfully")

	// 1. Initialize core schema (Go-native)
	if err := Initialize(db); err != nil {
		logger.Error("Schema initialization failure: %v", err)
	}

	// 2. Run incremental migrations (if any left)
	if err := Migrate(db); err != nil {
		logger.Error("Migration failure: %v", err)
	}

	return db, nil
}

// Initialize runs the core schema defined in db/schema/init.sql
func Initialize(db *sqlx.DB) error {
	schemaDir := filepath.Join("db", "schema")
	initPath := filepath.Join(schemaDir, "init.sql")

	content, err := os.ReadFile(initPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No initial schema defined
		}
		return err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Basic parser for \i commands
		if strings.HasPrefix(line, "\\i ") {
			sqlFile := strings.TrimSpace(strings.TrimPrefix(line, "\\i "))
			err := executeSQLFile(db, filepath.Join(schemaDir, sqlFile))
			if err != nil {
				return fmt.Errorf("failed to execute schema file %s: %v", sqlFile, err)
			}
			logger.Info("Initialized schema component: %s", sqlFile)
		}
	}
	return nil
}

func executeSQLFile(db *sqlx.DB, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	_, err = db.Exec(string(content))
	return err
}

func Migrate(db *sqlx.DB) error {
	migrationsDir := filepath.Join("db", "migrations")
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		// If directory doesn't exist, skip silently
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	count := 0
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".sql" {
			path := filepath.Join(migrationsDir, f.Name())
			if err := executeSQLFile(db, path); err != nil {
				return fmt.Errorf("failed to execute migration %s: %v", f.Name(), err)
			}
			logger.Info("Applied migration: %s", f.Name())
			count++
		}
	}

	if count > 0 {
		logger.Info("Successfully applied %d migrations", count)
	}
	return nil
}
