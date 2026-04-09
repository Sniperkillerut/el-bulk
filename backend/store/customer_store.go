package store

import (
	"github.com/el-bulk/backend/models"
	"github.com/jmoiron/sqlx"
)

type CustomerStore struct {
	*BaseStore[models.Customer]
}

func NewCustomerStore(db *sqlx.DB) *CustomerStore {
	return &CustomerStore{
		BaseStore: NewBaseStore[models.Customer](db, "customer"),
	}
}
