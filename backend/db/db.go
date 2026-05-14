package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"time"

	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	// Cloud SQL Connector imports
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "cloud.google.com/go/cloudsqlconn/postgres/pgxv5"
	"golang.org/x/oauth2/google"
)

func init() {
	// Walk up from cwd until we find a .env file (works regardless of where the
	// binary is invoked from). Silently skip if not found — Docker supplies vars
	// via docker-compose environment instead.
	dir, _ := os.Getwd()
	for {
		localCandidate := filepath.Join(dir, ".env.local")
		if _, err := os.Stat(localCandidate); err == nil {
			_ = godotenv.Load(localCandidate)
			// We don't break here, we might still want to load base .env for defaults
		}

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
	ctx := context.Background()
	start := time.Now()
	logger.InfoCtx(ctx, "Attempting to connect to database...")
	db, err := ConnectResilient()
	if err != nil {
		logger.ErrorCtx(ctx, "Database connection failed after %v: %v", time.Since(start), err)
		os.Exit(1)
	}
	logger.InfoCtx(ctx, "Database connection established in %v", time.Since(start))
	return db
}

func ConnectResilient() (*sqlx.DB, error) {
	ctx := context.Background()
	dsn := os.Getenv("DATABASE_URL")
	instanceName := os.Getenv("INSTANCE_CONNECTION_NAME")

	logger.InfoCtx(ctx, "🔍 [DB] Diagnostics: INSTANCE_CONNECTION_NAME=%q | DATABASE_URL_SET=%v (len=%d) | DB_IAM_AUTH=%q",
		instanceName, dsn != "", len(dsn), os.Getenv("DB_IAM_AUTH"))

	// 0. Cloud SQL Connector (Recommended for local/complex setups)
	// OR Unix Domain Sockets (Recommended for Cloud Run)
	if instanceName != "" {
		logger.InfoCtx(ctx, "🔌 Cloud SQL: Using native connection via Unix Socket: /cloudsql/%s", instanceName)

		// Extract DB name from DSN or fallback to elbulk
		dbName := "elbulk"
		if dsn != "" && strings.Contains(dsn, "/") {
			parts := strings.Split(dsn, "/")
			dbName = strings.Split(parts[len(parts)-1], "?")[0]
		}

		user := "elbulk"
		pass := ""

		// Auto-detect IAM user from environment or DSN
		iamUser := os.Getenv("DB_IAM_USER")
		if iamUser == "" && dsn != "" && strings.Contains(dsn, "@") {
			// Extract user from postgres://user:pass@host/db
			uPart := strings.Split(strings.TrimPrefix(dsn, "postgres://"), ":")[0]
			if strings.Contains(uPart, "@") {
				iamUser = uPart
			}
		}

		if os.Getenv("DB_IAM_AUTH") == "true" || (iamUser != "" && strings.Contains(iamUser, "@")) {
			if iamUser != "" {
				user = strings.TrimSuffix(iamUser, ".gserviceaccount.com")
			}
			
			token, err := getIAMToken()
			if err != nil {
				logger.ErrorCtx(ctx, "❌ [DB] Failed to fetch IAM token: %v", err)
			} else {
				pass = token
				logger.InfoCtx(ctx, "🔐 Cloud SQL: Using IAM-based authentication for user: %s", user)
			}
		}

		// Build Unix Socket DSN
		// format: host=/cloudsql/INSTANCE user=USER password=PASS dbname=DB sslmode=disable
		// Important: Password (token) can contain special chars, so we use the key=value format which is safer
		connectorDsn := fmt.Sprintf("host='/cloudsql/%s' user='%s' dbname='%s' sslmode='disable'", instanceName, user, dbName)
		if pass != "" {
			connectorDsn += fmt.Sprintf(" password='%s'", pass)
		}

		db, err := sqlx.Open("postgres", connectorDsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open database via Unix socket: %v", err)
		}

		maxOpen := getEnvInt("DB_MAX_OPEN_CONNS", 10)
		maxIdle := getEnvInt("DB_MAX_IDLE_CONNS", 2)
		maxLifetime := getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute)
		
		db.SetMaxOpenConns(maxOpen)
		db.SetMaxIdleConns(maxIdle)
		db.SetConnMaxLifetime(maxLifetime)

		logger.InfoCtx(ctx, "⚙️ DB Pooling (Unix): MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v", maxOpen, maxIdle, maxLifetime)

		if err := Initialize(db); err != nil {
			logger.ErrorCtx(ctx, "Schema initialization failure: %v", err)
		}
		if err := Migrate(db); err != nil {
			logger.ErrorCtx(ctx, "Migration failure: %v", err)
		}

		return db, nil
	}

	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	// 0. Dynamic SSL Provisioning (for GCP/Cloud Run without local cert files)
	rootCert := os.Getenv("DB_SSL_ROOT_CERT")
	if rootCert != "" {
		certDir := "/tmp/elbulk-certs"
		if err := os.MkdirAll(certDir, 0700); err != nil {
			logger.ErrorCtx(ctx, "Failed to create cert directory: %v", err)
		} else {
			rootCertPath := filepath.Join(certDir, "root.crt")
			if err := os.WriteFile(rootCertPath, []byte(rootCert), 0600); err != nil {
				logger.ErrorCtx(ctx, "Failed to write root cert: %v", err)
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
					logger.InfoCtx(ctx, "🔒 SSL Provisioning: Detected environment certificates, updated DSN for secure connection.")
				}
			}
		}
	}

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	maxOpen := getEnvInt("DB_MAX_OPEN_CONNS", 10)
	maxIdle := getEnvInt("DB_MAX_IDLE_CONNS", 2)
	maxLifetime := getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute)
	
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(maxLifetime)

	logger.InfoCtx(ctx, "⚙️ DB Pooling (DSN): MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v", maxOpen, maxIdle, maxLifetime)

	logger.InfoCtx(ctx, "Database connected successfully")

	// 1. Initialize core schema (Go-native)
	if err := Initialize(db); err != nil {
		logger.ErrorCtx(ctx, "Schema initialization failure: %v", err)
	}

	// 2. Run incremental migrations (if any left)
	if err := Migrate(db); err != nil {
		logger.ErrorCtx(ctx, "Migration failure: %v", err)
	}

	return db, nil
}

// Initialize runs the core schema defined in db/schema/init.sql
func Initialize(db *sqlx.DB) error {
	ctx := context.Background()
	start := time.Now()
	schemaDir := filepath.Join("db", "schema")
	initPath := filepath.Join(schemaDir, "init.sql")

	content, err := os.ReadFile(initPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.DebugCtx(ctx, "No init.sql found at %s, skipping initialization", initPath)
			return nil
		}
		return err
	}

	logger.InfoCtx(ctx, "Initializing database schema from %s (with lock)", initPath)

	// Start a transaction for the entire initialization to acquire an advisory lock
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start initialization transaction: %v", err)
	}
	defer tx.Rollback()

	// Acquire a transaction-level advisory lock (arbitrary ID 1111)
	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock($1)", 1111); err != nil {
		return fmt.Errorf("failed to acquire schema initialization lock: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "\\i ") {
			sqlFile := strings.TrimSpace(strings.TrimPrefix(line, "\\i "))
			if strings.Contains(sqlFile, "..") {
				return fmt.Errorf("invalid path in schema file: %s", sqlFile)
			}
			fileStart := time.Now()
			err := executeSQLFileTx(tx, filepath.Join(schemaDir, sqlFile))
			if err != nil {
				return fmt.Errorf("failed to execute schema file %s: %v", sqlFile, err)
			}
			logger.TraceCtx(ctx, "Initialized schema component: %s (took %v)", sqlFile, time.Since(fileStart))
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit initialization transaction: %v", err)
	}

	logger.InfoCtx(ctx, "Schema initialization completed in %v", time.Since(start))
	return nil
}


func Migrate(db *sqlx.DB) error {
	ctx := context.Background()
	start := time.Now()
	migrationsDir := filepath.Join("db", "migrations")
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			logger.DebugCtx(ctx, "No migrations directory found, skipping migrations")
			return nil
		}
		return err
	}

	// Start a transaction to acquire a global advisory lock for the migration process
	lockTx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start migration lock transaction: %v", err)
	}
	defer lockTx.Rollback()

	// Acquire a transaction-level advisory lock (arbitrary ID 1112)
	if _, err := lockTx.ExecContext(ctx, "SELECT pg_advisory_xact_lock($1)", 1112); err != nil {
		return fmt.Errorf("failed to acquire migration lock: %v", err)
	}

	var applied []string
	err = lockTx.SelectContext(ctx, &applied, "SELECT name FROM migration")
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
				logger.TraceCtx(ctx, "Migration already applied: %s", f.Name())
				continue
			}

			path := filepath.Join(migrationsDir, f.Name())
			logger.InfoCtx(ctx, "Applying migration: %s", f.Name())
			migStart := time.Now()

			tx, err := db.Beginx()
			if err != nil {
				return fmt.Errorf("failed to start transaction for %s: %v", f.Name(), err)
			}

			if err := executeSQLFileTx(tx, path); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to execute migration %s: %v", f.Name(), err)
			}

			if _, err := tx.ExecContext(ctx, "INSERT INTO migration (name) VALUES ($1)", f.Name()); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to record migration %s: %v", f.Name(), err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit migration %s: %v", f.Name(), err)
			}

			logger.DebugCtx(ctx, "Successfully applied migration %s in %v", f.Name(), time.Since(migStart))
			count++
		}
	}

	// Release the global lock by committing the lock transaction
	if err := lockTx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration lock transaction: %v", err)
	}

	if count > 0 {
		logger.InfoCtx(ctx, "Migration run complete: applied %d new migrations in %v", count, time.Since(start))
	} else {
		logger.InfoCtx(ctx, "No new migrations to apply (took %v)", time.Since(start))
	}
	return nil
}

func executeSQLFileTx(tx *sqlx.Tx, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(context.Background(), string(content))
	return err
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return fallback
}

// getIAMToken fetches a fresh Google OAuth2 access token for IAM database authentication.
func getIAMToken() (string, error) {
	ctx := context.Background()
	// Scopes required for Cloud SQL IAM Auth
	scopes := []string{
		"https://www.googleapis.com/auth/sqlservice.admin",
		"https://www.googleapis.com/auth/sqlservice.login",
		"https://www.googleapis.com/auth/cloud-platform",
	}
	creds, err := google.FindDefaultCredentials(ctx, scopes...)
	if err != nil {
		return "", err
	}
	token, err := creds.TokenSource.Token()
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}
