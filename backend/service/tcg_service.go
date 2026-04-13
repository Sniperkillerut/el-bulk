package service

import (
	"context"
	"fmt"
	"time"

	"github.com/el-bulk/backend/external"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/el-bulk/backend/utils/logger"
)

type TCGService struct {
	Store *store.TCGStore
}

func NewTCGService(s *store.TCGStore) *TCGService {
	return &TCGService{Store: s}
}

func (s *TCGService) List(ctx context.Context, isAdmin bool) ([]models.TCG, error) {
	logger.TraceCtx(ctx, "TCGService.List | Admin: %v", isAdmin)
	// If not admin, only show active TCGs
	return s.Store.ListWithCount(ctx, !isAdmin)
}

func (s *TCGService) Create(ctx context.Context, input models.TCGInput) (*models.TCG, error) {
	logger.TraceCtx(ctx, "Entering TCGService.Create | ID: %s", input.ID)
	return s.Store.Create(ctx, input)
}

func (s *TCGService) Update(ctx context.Context, id string, input models.TCGInput) (*models.TCG, error) {
	logger.TraceCtx(ctx, "Entering TCGService.Update | ID: %s", id)
	return s.Store.Update(ctx, id, input)
}

func (s *TCGService) Delete(ctx context.Context, id string) error {
	logger.TraceCtx(ctx, "Entering TCGService.Delete | ID: %s", id)
	count, err := s.Store.GetProductCount(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("cannot delete TCG with %d existing products", count)
	}
	return s.Store.BaseStore.Delete(ctx, id)
}

func (s *TCGService) SyncSets(ctx context.Context, tcgID string) (int, error) {
	logger.TraceCtx(ctx, "Entering TCGService.SyncSets | TCG: %s", tcgID)
	if tcgID != "mtg" {
		return 0, fmt.Errorf("sync currently only supported for MTG")
	}

	sets, err := external.FetchSets(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to fetch sets for %s: %v", tcgID, err)
		return 0, fmt.Errorf("failed to fetch sets: %w", err)
	}
	logger.DebugCtx(ctx, "Fetched %d sets from external API for %s", len(sets), tcgID)

	tx, err := s.Store.DB.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	for _, set := range sets {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO tcg_set (tcg, code, name, released_at, set_type)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (tcg, code) DO UPDATE SET
				name = EXCLUDED.name,
				released_at = EXCLUDED.released_at,
				set_type = EXCLUDED.set_type
		`, "mtg", set.Code, set.Name, set.ReleasedAt, set.SetType)
		if err != nil {
			logger.ErrorCtx(ctx, "Error syncing set %s: %v", set.Code, err)
			continue
		}
	}

	_, _ = tx.ExecContext(ctx, "INSERT INTO setting (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value", "last_set_sync", time.Now().Format(time.RFC3339))

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	logger.InfoCtx(ctx, "Successfully synced %d sets for %s", len(sets), tcgID)
	return len(sets), nil
}
