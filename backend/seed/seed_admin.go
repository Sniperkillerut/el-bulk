package main

import (
	"os"

	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

func seedAdmin(db *sqlx.DB) string {
	user := os.Getenv("ADMIN_USERNAME")
	pass := os.Getenv("ADMIN_PASSWORD")
	if user == "" {
		user = "admin"
	}
	if pass == "" {
		pass = "elbulk2024!"
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	var id string
	db.QueryRow(`
		INSERT INTO admin (username, password_hash)
		VALUES ($1, $2)
		ON CONFLICT (username) DO UPDATE SET password_hash = EXCLUDED.password_hash
		RETURNING id
	`, user, string(hash)).Scan(&id)
	logger.Info("Admin user '%s' created/updated", user)
	return id
}
