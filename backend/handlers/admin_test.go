package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestAdminHandler_Login(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := NewAdminHandler(sqlxDB)

	password := "password123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	tests := []struct {
		name           string
		reqBody        models.LoginRequest
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "Successful Login",
			reqBody: models.LoginRequest{
				Username: "admin",
				Password: password,
			},
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "password_hash"}).
					AddRow("admin-uuid", "admin", string(hash))
				mock.ExpectQuery("SELECT \\* FROM admin WHERE username = \\$1").
					WithArgs("admin").
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "User Not Found",
			reqBody: models.LoginRequest{
				Username: "nonexistent",
				Password: password,
			},
			mockSetup: func() {
				mock.ExpectQuery("SELECT \\* FROM admin WHERE username = \\$1").
					WithArgs("nonexistent").
					WillReturnError(http.ErrNoLocation) // Any error to simulate not found
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Wrong Password",
			reqBody: models.LoginRequest{
				Username: "admin",
				Password: "wrongpassword",
			},
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "password_hash"}).
					AddRow("admin-uuid", "admin", string(hash))
				mock.ExpectQuery("SELECT \\* FROM admin WHERE username = \\$1").
					WithArgs("admin").
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Invalid JSON",
			reqBody: models.LoginRequest{}, // This won't trigger invalid JSON unless the body is actually malformed
			mockSetup: func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var body []byte
			if tt.name == "Invalid JSON" {
				body = []byte("{invalid-json}")
			} else {
				body, _ = json.Marshal(tt.reqBody)
			}

			req, _ := http.NewRequest("POST", "/api/admin/login", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			h.Login(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			
			if tt.expectedStatus == http.StatusOK {
				var res models.LoginResponse
				json.NewDecoder(rr.Body).Decode(&res)
				assert.NotEmpty(t, res.Token)
			}
		})
	}
}

func TestAdminHandler_Login_TokenFail(t *testing.T) {
	// This is hard to trigger unless we force a secret that causes SignedString to fail,
	// which is unlikely with HS256 and a string secret.
	// But let's check default secret path.
	os.Unsetenv("JWT_SECRET")
	
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	h := NewAdminHandler(sqlxDB)

	password := "password123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	rows := sqlmock.NewRows([]string{"id", "username", "password_hash"}).
		AddRow("admin-uuid", "admin", string(hash))
	mock.ExpectQuery("SELECT \\* FROM admin WHERE username = \\$1").
		WithArgs("admin").
		WillReturnRows(rows)

	reqBody := models.LoginRequest{Username: "admin", Password: password}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/admin/login", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	h.Login(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}
