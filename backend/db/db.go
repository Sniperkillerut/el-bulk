package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	fmt.Println("✓ Database connected")
	return db
}
