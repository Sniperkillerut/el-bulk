package models

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestGetAllTranslations(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	rows := sqlmock.NewRows([]string{"key", "locale", "value"}).
		AddRow("welcome", "en", "Welcome").
		AddRow("welcome", "es", "Bienvenido")

	mock.ExpectQuery("SELECT key, locale, value FROM translation").WillReturnRows(rows)

	translations, err := GetAllTranslations(context.Background(), sqlxDB)
	assert.NoError(t, err)
	assert.Len(t, translations, 2)
	assert.Equal(t, "Welcome", translations["en"]["welcome"])
	assert.Equal(t, "Bienvenido", translations["es"]["welcome"])
}

func TestGetAllTranslations_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectQuery("SELECT key, locale, value FROM translation").WillReturnError(assert.AnError)

	translations, err := GetAllTranslations(context.Background(), sqlxDB)
	assert.Error(t, err)
	assert.Nil(t, translations)
}

func TestGetTranslationsByLocale(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	rows := sqlmock.NewRows([]string{"key", "value"}).
		AddRow("welcome", "Welcome")

	mock.ExpectQuery("SELECT key, value FROM translation WHERE locale = \\$1").WithArgs("en").WillReturnRows(rows)

	translations, err := GetTranslationsByLocale(context.Background(), sqlxDB, "en")
	assert.NoError(t, err)
	assert.Len(t, translations, 1)
	assert.Equal(t, "Welcome", translations["welcome"])
}

func TestGetTranslationsByLocale_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectQuery("SELECT key, value FROM translation WHERE locale = \\$1").WithArgs("en").WillReturnError(assert.AnError)

	translations, err := GetTranslationsByLocale(context.Background(), sqlxDB, "en")
	assert.Error(t, err)
	assert.Nil(t, translations)
}

func TestUpsertTranslation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	trans := Translation{
		Key:    "welcome",
		Locale: "en",
		Value:  "Welcome",
	}

	mock.ExpectExec("INSERT INTO translation").WithArgs(trans.Key, trans.Locale, trans.Value).WillReturnResult(sqlmock.NewResult(1, 1))

	err = UpsertTranslation(context.Background(), sqlxDB, trans)
	assert.NoError(t, err)
}

func TestDeleteTranslation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectExec("DELETE FROM translation").WithArgs("welcome", "en").WillReturnResult(sqlmock.NewResult(1, 1))

	err = DeleteTranslation(context.Background(), sqlxDB, "welcome", "en")
	assert.NoError(t, err)
}

func TestListAllTranslationKeys(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	now := time.Now()
	rows := sqlmock.NewRows([]string{"key", "locale", "value", "updated_at"}).
		AddRow("welcome", "en", "Welcome", now)

	mock.ExpectQuery("SELECT key, locale, value, updated_at FROM translation").WillReturnRows(rows)

	translations, err := ListAllTranslationKeys(context.Background(), sqlxDB)
	assert.NoError(t, err)
	assert.Len(t, translations, 1)
	assert.Equal(t, "welcome", translations[0].Key)
	assert.Equal(t, "en", translations[0].Locale)
}

func TestListAllTranslationKeys_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectQuery("SELECT key, locale, value, updated_at FROM translation").WillReturnError(assert.AnError)

	translations, err := ListAllTranslationKeys(context.Background(), sqlxDB)
	assert.Error(t, err)
	assert.Nil(t, translations)
}
