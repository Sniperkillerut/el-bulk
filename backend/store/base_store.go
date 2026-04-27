package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/el-bulk/backend/utils/logger"
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

func (s *BaseStore[T]) List(ctx context.Context, conditions string, args ...interface{}) ([]T, error) {
	items := make([]T, 0)
	query := fmt.Sprintf("SELECT * FROM %s", s.TableName)
	if conditions != "" {
		if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(conditions)), "ORDER BY") && 
		   !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(conditions)), "LIMIT") {
			query += " WHERE " + conditions
		} else {
			query += " " + conditions
		}
	}
	
	start := time.Now()
	rebound := s.DB.Rebind(query)
	logger.TraceCtx(ctx, "[DB] Executing List on %s: %s | Args: %+v", s.TableName, rebound, args)
	
	err := s.DB.Unsafe().SelectContext(ctx, &items, rebound, args...)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] List on %s failed: %v", s.TableName, err)
		return nil, err
	}
	logger.DebugCtx(ctx, "[DB] List on %s took %v", s.TableName, time.Since(start))
	return items, nil
}

func (s *BaseStore[T]) GetByID(ctx context.Context, id string) (*T, error) {
	var item T
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", s.TableName)
	
	start := time.Now()
	logger.TraceCtx(ctx, "[DB] Executing GetByID on %s: %s | ID: %s", s.TableName, query, id)
	
	err := s.DB.Unsafe().GetContext(ctx, &item, query, id)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] GetByID on %s (%s) failed: %v", s.TableName, id, err)
		return nil, err
	}
	logger.DebugCtx(ctx, "[DB] GetByID on %s took %v", s.TableName, time.Since(start))
	return &item, nil
}

func (s *BaseStore[T]) Create(item *T) error {
	// sqlx can handle struct insertion with NamedQuery if we build the query properly.
	// For now, we'll keep it simple as most entities have specific needs.
	// But a generic NamedExec helper would be:
	// _, err := s.DB.NamedExec(fmt.Sprintf("INSERT INTO %s (...) VALUES (...)", s.TableName), item)
	return fmt.Errorf("generic Create not fully implemented, use specific store methods")
}

func (s *BaseStore[T]) Update(ctx context.Context, id string, updates map[string]interface{}) (*T, error) {
	if len(updates) == 0 {
		return s.GetByID(ctx, id)
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

	start := time.Now()
	logger.TraceCtx(ctx, "[DB] Executing Update on %s: %s | Args: %+v", s.TableName, query, args)

	var item T
	err := s.DB.QueryRowxContext(ctx, query, args...).StructScan(&item)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] Update on %s (%s) failed: %v", s.TableName, id, err)
		return nil, err
	}
	logger.DebugCtx(ctx, "[DB] Update on %s took %v", s.TableName, time.Since(start))
	return &item, nil
}

func (s *BaseStore[T]) Delete(ctx context.Context, id string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", s.TableName)
	
	start := time.Now()
	logger.TraceCtx(ctx, "[DB] Executing Delete on %s | ID: %s", s.TableName, id)
	
	res, err := s.DB.ExecContext(ctx, query, id)
	if err != nil {
		logger.ErrorCtx(ctx, "[DB] Delete on %s (%s) failed: %v", s.TableName, id, err)
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("no rows deleted")
	}
	logger.DebugCtx(ctx, "[DB] Delete on %s took %v", s.TableName, time.Since(start))
	return nil
}

func (s *BaseStore[T]) WithTransaction(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := s.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
