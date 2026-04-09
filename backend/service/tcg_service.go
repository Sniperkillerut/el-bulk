package service

import (
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

func (s *TCGService) List(isAdmin bool) ([]models.TCG, error) {
	// For now, TCGStore.ListWithCount returns all with counts.
	// We might want to filter activeOnly for public view in the future.
	return s.Store.ListWithCount()
}

func (s *TCGService) Create(input models.TCGInput) (*models.TCG, error) {
	return s.Store.Create(input)
}

func (s *TCGService) Update(id string, input models.TCGInput) (*models.TCG, error) {
	return s.Store.Update(id, input)
}

func (s *TCGService) Delete(id string) error {
	count, err := s.Store.GetProductCount(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("cannot delete TCG with %d existing products", count)
	}
	return s.Store.BaseStore.Delete(id)
}

func (s *TCGService) SyncSets(tcgID string) (int, error) {
	if tcgID != "mtg" {
		return 0, fmt.Errorf("sync currently only supported for MTG")
	}

	sets, err := external.FetchSets()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch sets: %w", err)
	}

	tx, err := s.Store.DB.Beginx()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	for _, set := range sets {
		_, err := tx.Exec(`
			INSERT INTO tcg_set (tcg, code, name, released_at, set_type)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (tcg, code) DO UPDATE SET
				name = EXCLUDED.name,
				released_at = EXCLUDED.released_at,
				set_type = EXCLUDED.set_type
		`, "mtg", set.Code, set.Name, set.ReleasedAt, set.SetType)
		if err != nil {
			logger.Error("Error syncing set %s: %v", set.Code, err)
			continue
		}
	}

	_, _ = tx.Exec("INSERT INTO setting (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value", "last_set_sync", time.Now().Format(time.RFC3339))

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return len(sets), nil
}
