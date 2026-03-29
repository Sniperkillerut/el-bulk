package db

import (
	"fmt"
	"os"
	"path/filepath"

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

	// Run migrations
	if err := Migrate(db); err != nil {
		logger.Error("Migration failure: %v", err)
	}

	return db, nil
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

	// For simplicity in this demo, we'll just run all .sql files.
	// In a production app, we should use a migrations table to track applied ones.
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".sql" {
			path := filepath.Join(migrationsDir, f.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read migration %s: %v", f.Name(), err)
			}
			_, err = db.Exec(string(content))
			if err != nil {
				return fmt.Errorf("failed to execute migration %s: %v", f.Name(), err)
			}
			logger.Info("Applied migration: %s", f.Name())
		}
	}
	return nil
}
