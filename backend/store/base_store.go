package store

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// BaseStore provides generic CRUD operations for a given model type T.
type BaseStore[T any] struct {
	DB        *sqlx.DB
	TableName string
}

func NewBaseStore[T any](db *sqlx.DB, tableName string) *BaseStore[T] {
	return &BaseStore[T]{
		DB:        db,
		TableName: tableName,
	}
}

func (s *BaseStore[T]) List(conditions string, args ...interface{}) ([]T, error) {
	var items []T
	query := fmt.Sprintf("SELECT * FROM %s", s.TableName)
	if conditions != "" {
		if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(conditions)), "ORDER BY") && 
		   !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(conditions)), "LIMIT") {
			query += " WHERE " + conditions
		} else {
			query += " " + conditions
		}
	}
	
	err := s.DB.Select(&items, s.DB.Rebind(query), args...)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *BaseStore[T]) GetByID(id string) (*T, error) {
	var item T
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", s.TableName)
	err := s.DB.Get(&item, query, id)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *BaseStore[T]) Create(item *T) error {
	// sqlx can handle struct insertion with NamedQuery if we build the query properly.
	// For now, we'll keep it simple as most entities have specific needs.
	// But a generic NamedExec helper would be:
	// _, err := s.DB.NamedExec(fmt.Sprintf("INSERT INTO %s (...) VALUES (...)", s.TableName), item)
	return fmt.Errorf("generic Create not fully implemented, use specific store methods")
}

func (s *BaseStore[T]) Update(id string, updates map[string]interface{}) (*T, error) {
	if len(updates) == 0 {
		return s.GetByID(id)
	}

	setClauses := []string{}
	args := []interface{}{}
	i := 1
	for col, val := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d RETURNING *", s.TableName, strings.Join(setClauses, ", "), i)
	args = append(args, id)

	var item T
	err := s.DB.QueryRowx(query, args...).StructScan(&item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *BaseStore[T]) Delete(id string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", s.TableName)
	res, err := s.DB.Exec(query, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("no rows deleted")
	}
	return nil
}
