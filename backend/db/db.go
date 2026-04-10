package db

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"time"
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
	start := time.Now()
	logger.Info("Attempting to connect to database...")
	db, err := ConnectResilient()
	if err != nil {
		logger.Error("Database connection failed after %v: %v", time.Since(start), err)
		os.Exit(1)
	}
	logger.Info("Database connection established in %v", time.Since(start))
	return db
}

func ConnectResilient() (*sqlx.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	// 0. Dynamic SSL Provisioning (for GCP/Cloud Run without local cert files)
	rootCert := os.Getenv("DB_SSL_ROOT_CERT")
	if rootCert != "" {
		certDir := "/tmp/elbulk-certs"
		if err := os.MkdirAll(certDir, 0700); err != nil {
			logger.Error("Failed to create cert directory: %v", err)
		} else {
			rootCertPath := filepath.Join(certDir, "root.crt")
			if err := os.WriteFile(rootCertPath, []byte(rootCert), 0600); err != nil {
				logger.Error("Failed to write root cert: %v", err)
			} else {
				// Append SSL parameters to DSN if not already present
				if !strings.Contains(dsn, "sslrootcert") {
					if strings.Contains(dsn, "?") {
						dsn += "&"
					} else {
						dsn += "?"
					}
					dsn += fmt.Sprintf("sslmode=verify-full&sslrootcert=%s", rootCertPath)

					// Optionally add client cert/key if provided
					clientCert := os.Getenv("DB_SSL_CERT")
					clientKey := os.Getenv("DB_SSL_KEY")
					if clientCert != "" && clientKey != "" {
						certPath := filepath.Join(certDir, "client.crt")
						keyPath := filepath.Join(certDir, "client.key")
						_ = os.WriteFile(certPath, []byte(clientCert), 0600)
						_ = os.WriteFile(keyPath, []byte(clientKey), 0600)
						dsn += fmt.Sprintf("&sslcert=%s&sslkey=%s", certPath, keyPath)
					}
					logger.Info("🔒 SSL Provisioning: Detected environment certificates, updated DSN for secure connection.")
				}
			}
		}
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
	start := time.Now()
	schemaDir := filepath.Join("db", "schema")
	initPath := filepath.Join(schemaDir, "init.sql")

	content, err := os.ReadFile(initPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Debug("No init.sql found at %s, skipping initialization", initPath)
			return nil
		}
		return err
	}

	logger.Info("Initializing database schema from %s", initPath)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "\\i ") {
			sqlFile := strings.TrimSpace(strings.TrimPrefix(line, "\\i "))
			if strings.Contains(sqlFile, "..") {
				return fmt.Errorf("invalid path in schema file: %s", sqlFile)
			}
			fileStart := time.Now()
			err := executeSQLFile(db, filepath.Join(schemaDir, sqlFile))
			if err != nil {
				return fmt.Errorf("failed to execute schema file %s: %v", sqlFile, err)
			}
			logger.Trace("Initialized schema component: %s (took %v)", sqlFile, time.Since(fileStart))
		}
	}
	logger.Info("Schema initialization completed in %v", time.Since(start))
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
	start := time.Now()
	migrationsDir := filepath.Join("db", "migrations")
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Debug("No migrations directory found, skipping migrations")
			return nil
		}
		return err
	}

	var applied []string
	err = db.Select(&applied, "SELECT name FROM migration")
	if err != nil {
		return fmt.Errorf("failed to fetch applied migrations: %v", err)
	}

	appliedMap := make(map[string]bool)
	for _, name := range applied {
		appliedMap[name] = true
	}

	count := 0
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".sql" {
			if appliedMap[f.Name()] {
				logger.Trace("Migration already applied: %s", f.Name())
				continue
			}

			path := filepath.Join(migrationsDir, f.Name())
			logger.Info("Applying migration: %s", f.Name())
			migStart := time.Now()

			tx, err := db.Beginx()
			if err != nil {
				return fmt.Errorf("failed to start transaction for %s: %v", f.Name(), err)
			}

			if err := executeSQLFileTx(tx, path); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to execute migration %s: %v", f.Name(), err)
			}

			if _, err := tx.Exec("INSERT INTO migration (name) VALUES ($1)", f.Name()); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to record migration %s: %v", f.Name(), err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit migration %s: %v", f.Name(), err)
			}

			logger.Debug("Successfully applied migration %s in %v", f.Name(), time.Since(migStart))
			count++
		}
	}

	if count > 0 {
		logger.Info("Migration run complete: applied %d new migrations in %v", count, time.Since(start))
	} else {
		logger.Info("No new migrations to apply (took %v)", time.Since(start))
	}
	return nil
}

func executeSQLFileTx(tx *sqlx.Tx, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	_, err = tx.Exec(string(content))
	return err
}
