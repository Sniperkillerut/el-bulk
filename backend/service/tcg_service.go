package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/el-bulk/backend/external"
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/el-bulk/backend/utils/logger"
)

type TCGService struct {
	Store          *store.TCGStore
	RefreshService *RefreshService
}

func NewTCGService(s *store.TCGStore, refresh *RefreshService) *TCGService {
	return &TCGService{
		Store:          s,
		RefreshService: refresh,
	}
}

func (s *TCGService) ResetCache() {
	s.Store.InvalidateCache()
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

type syncSetDBParams struct {
	TCG        string `db:"tcg"`
	Code       string `db:"code"`
	Name       string `db:"name"`
	ReleasedAt string `db:"released_at"`
	SetType    string `db:"set_type"`
	CKName     string `db:"ck_name"`
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

	// Map and chunk the sets
	batchSize := 1000
	var dbSets []syncSetDBParams
	for _, set := range sets {
		ckGuess := external.NormalizeCKEdition(set.Name)
		dbSets = append(dbSets, syncSetDBParams{
			TCG:        "mtg",
			Code:       set.Code,
			Name:       set.Name,
			ReleasedAt: set.ReleasedAt,
			SetType:    set.SetType,
			CKName:     ckGuess,
		})
	}

	for i := 0; i < len(dbSets); i += batchSize {
		end := i + batchSize
		if end > len(dbSets) {
			end = len(dbSets)
		}
		chunk := dbSets[i:end]

		var placeholders []string
		var args []interface{}

		for j, set := range chunk {
			offset := j * 6
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", offset+1, offset+2, offset+3, offset+4, offset+5, offset+6))
			args = append(args, set.TCG, set.Code, set.Name, set.ReleasedAt, set.SetType, set.CKName)
		}

		query := fmt.Sprintf(`
			INSERT INTO tcg_set (tcg, code, name, released_at, set_type, ck_name)
			VALUES %s
			ON CONFLICT (tcg, code) DO UPDATE SET
				name = EXCLUDED.name,
				released_at = EXCLUDED.released_at,
				set_type = EXCLUDED.set_type,
				-- Only update ck_name if the existing one is NULL
				ck_name = COALESCE(tcg_set.ck_name, EXCLUDED.ck_name)
		`, strings.Join(placeholders, ","))

		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			logger.ErrorCtx(ctx, "Error syncing chunk of sets: %v", err)
			return 0, fmt.Errorf("failed to insert sets: %w", err)
		}
	}

	_, _ = tx.ExecContext(ctx, "INSERT INTO setting (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value", "last_set_sync", time.Now().Format(time.RFC3339))

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	logger.InfoCtx(ctx, "Successfully synced %d sets for %s", len(sets), tcgID)
	return len(sets), nil
}

func (s *TCGService) SyncPrices(ctx context.Context, tcgID string) (int, int, error) {
	logger.TraceCtx(ctx, "Entering TCGService.SyncPrices | TCG: %s", tcgID)

	// Price refresh is currently only implemented for MTG via Scryfall/CardKingdom
	if tcgID != "mtg" {
		return 0, 0, fmt.Errorf("price sync currently only supported for MTG")
	}

	updated, errs := s.RefreshService.RunPriceRefresh(ctx, tcgID, nil)
	return updated, errs, nil
}
