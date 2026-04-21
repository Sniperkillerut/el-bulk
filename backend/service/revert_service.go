package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/el-bulk/backend/utils/logger"
	"strings"
)

type RevertService struct {
	AuditStore             *store.AuditStore
	Audit                  Auditer
	ProductStore           *store.ProductStore
	ProductService        *ProductService
	CategoryService       *CategoryService
	CategoryStore         *store.CategoryStore
	StorageLocationService *StorageLocationService
	StorageStore          *store.StorageLocationStore
	SettingsService       *SettingsService
}

func NewRevertService(as *store.AuditStore, a Auditer, ps *store.ProductStore, p *ProductService, c *CategoryService, cs *store.CategoryStore, sl *StorageLocationService, ss *store.StorageLocationStore, s *SettingsService) *RevertService {
	return &RevertService{
		AuditStore:             as,
		Audit:                  a,
		ProductStore:           ps,
		ProductService:        p,
		CategoryService:       c,
		CategoryStore:         cs,
		StorageLocationService: sl,
		StorageStore:          ss,
		SettingsService:       s,
	}
}

func (s *RevertService) Undo(ctx context.Context, logID string) error {
	log, err := s.AuditStore.GetByID(ctx, logID)
	if err != nil {
		return fmt.Errorf("audit log not found: %w", err)
	}

	// Protect against deep nested undos (only allow 1 level of UNDO_ for Redo)
	if strings.HasPrefix(log.Action, "UNDO_UNDO_") {
		return fmt.Errorf("cannot undo a Redo operation (maximum recursion depth reached)")
	}

	logger.InfoCtx(ctx, "Undoing action: %s on %s (%s)", log.Action, log.ResourceType, log.ResourceID)

	// 1. Capture REAL CURRENT state before revert (High Fidelity Snapshot)
	beforeState := s.getCurrentState(ctx, log.ResourceType, log.ResourceID)

	var revertErr error

	// Handle Redo logic by stripping the prefix for the handler
	baseAction := strings.TrimPrefix(log.Action, "UNDO_")
	isRedo := strings.HasPrefix(log.Action, "UNDO_")

	switch log.ResourceType {
	case "product":
		revertErr = s.undoProductAction(ctx, log, baseAction, isRedo)
	case "category":
		revertErr = s.undoCategoryAction(ctx, log, baseAction, isRedo)
	case "storage":
		revertErr = s.undoStorageAction(ctx, log, baseAction, isRedo)
	case "setting":
		revertErr = s.undoSettingAction(ctx, log, baseAction, isRedo)
	default:
		return fmt.Errorf("undo not supported for resource type: %s", log.ResourceType)
	}

	if revertErr != nil {
		return fmt.Errorf("failed to revert action: %w", revertErr)
	}

	// 2. Capture REAL NEW state after revert (High Fidelity Snapshot)
	afterState := s.getCurrentState(ctx, log.ResourceType, log.ResourceID)

	// 3. Create a NEW audit log entry that is perfectly reversible (Redo)
	// We use the REAL shots we just took, which is much more robust than flipping old log data.
	undoDetails := models.JSONB{
		"undone_log_id":   log.ID,
		"original_action": log.Action,
		"before":          beforeState,
		"after":           afterState,
	}

	// Handle deletion cases specially (if before was nil, it was a creation; if after is nil, it was a deletion)
	if beforeState == nil && afterState != nil {
		// New log created it
	} else if beforeState != nil && afterState == nil {
		// New log deleted it
		undoDetails["deleted"] = beforeState
	}

	undoAction := "UNDO_" + log.Action
	s.Audit.LogAction(ctx, undoAction, log.ResourceType, log.ResourceID, undoDetails)

	return nil
}

func (s *RevertService) getCurrentState(ctx context.Context, resourceType, resourceID string) interface{} {
	switch resourceType {
	case "product":
		p, err := s.ProductService.GetByID(ctx, resourceID, true)
		if err != nil {
			return nil
		}
		return p
	case "category":
		c, err := s.CategoryStore.GetByID(ctx, resourceID)
		if err != nil {
			return nil
		}
		return c
	case "storage":
		sl, err := s.StorageStore.GetByID(ctx, resourceID)
		if err != nil {
			return nil
		}
		return sl
	case "setting":
		settings, err := s.SettingsService.Store.GetAll(ctx)
		if err != nil {
			return nil
		}
		return settings[resourceID]
	default:
		return nil
	}
}

func (s *RevertService) undoProductAction(ctx context.Context, log *models.AuditLog, baseAction string, isRedo bool) error {
	details := log.Details

	switch baseAction {
	case "CREATE_PRODUCT":
		if isRedo {
			// Redo of Create (after undo deleted it) -> Restore state
			targetState := details["after"]
			if targetState == nil {
				return fmt.Errorf("missing 'after' state for redo create")
			}
			itemMap, err := s.mapProductToRawMap(targetState, log.ResourceID)
			if err != nil {
				return err
			}
			msg, _ := json.Marshal([]interface{}{itemMap})
			_, err = s.ProductStore.BulkUpsert(ctx, string(msg))
			return err
		}
		// Standard Undo CREATE -> DELETE
		return s.ProductService.Delete(ctx, log.ResourceID)

	case "UPDATE_PRODUCT", "DELETE_PRODUCT":
		// Both involve restoring a state
		// We always restore from the "before" state of the current log
		targetState := details["before"]
		if targetState == nil {
			return fmt.Errorf("missing previous state in audit log")
		}

		// Convert historical snapshot to something BulkUpsert understands
		itemMap, err := s.mapProductToRawMap(targetState, log.ResourceID)
		if err != nil {
			return err
		}

		msg, _ := json.Marshal([]interface{}{itemMap})
		_, err = s.ProductStore.BulkUpsert(ctx, string(msg))
		return err
	}

	return fmt.Errorf("undo not implemented for product action: %s", baseAction)
}

func (s *RevertService) undoCategoryAction(ctx context.Context, log *models.AuditLog, baseAction string, isRedo bool) error {
	details := log.Details

	switch baseAction {
	case "CREATE_CATEGORY":
		if !isRedo {
			return s.CategoryService.Delete(ctx, log.ResourceID)
		}
		fallthrough // Redo of create restored from target state
	case "UPDATE_CATEGORY":
		before, ok := details["before"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid 'before' state")
		}
		_, err := s.CategoryService.Update(ctx, log.ResourceID, before)
		return err
	case "DELETE_CATEGORY":
		// Logic same for Undo-Delete or Redo-Create
		// We always use "before" (which contains the snapshot of the category we want to restore)
		state, ok := details["before"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid state for category restoration")
		}
		
		isActive := state["is_active"].(bool)
		showBadge := state["show_badge"].(bool)
		searchable := state["searchable"].(bool)

		input := models.CustomCategoryInput{
			ID:         &log.ResourceID, // Preserve original ID
			Name:       fmt.Sprintf("%v", state["name"]),
			Slug:       fmt.Sprintf("%v", state["slug"]),
			IsActive:   &isActive,
			ShowBadge:  &showBadge,
			Searchable: &searchable,
		}
		
		// Restoration Part 1: Entity
		_, err := s.CategoryService.Create(ctx, input)
		if err != nil {
			return err
		}

		// Restoration Part 2: Deep Undo (Relationships)
		// We should also check for product_mappings in whatever state we are restoring
		mappingsKey := "product_mappings"
		if mappings, ok := details[mappingsKey].([]interface{}); ok && len(mappings) > 0 {
			productIDs := make([]string, 0, len(mappings))
			for _, m := range mappings {
				productIDs = append(productIDs, fmt.Sprintf("%v", m))
			}
			return s.CategoryStore.BatchAddProducts(ctx, log.ResourceID, productIDs)
		}
		return nil
	}
	return nil
}

func (s *RevertService) undoStorageAction(ctx context.Context, log *models.AuditLog, baseAction string, isRedo bool) error {
	details := log.Details
	switch baseAction {
	case "CREATE_STORAGE":
		if !isRedo {
			return s.StorageLocationService.Delete(ctx, log.ResourceID)
		}
		fallthrough
	case "UPDATE_STORAGE":
		before, ok := details["before"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid 'before' state")
		}
		return s.StorageLocationService.Update(ctx, log.ResourceID, fmt.Sprintf("%v", before["name"]))
	case "DELETE_STORAGE":
		state, ok := details["before"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid state for storage restoration")
		}
		
		// Restoration Part 1: Entity
		_, err := s.StorageLocationService.Create(ctx, fmt.Sprintf("%v", state["name"]), &log.ResourceID)
		if err != nil {
			return err
		}

		// Restoration Part 2: Deep Undo (Inventory)
		if rawMappings, ok := details["stock_mappings"].([]interface{}); ok && len(rawMappings) > 0 {
			data, _ := json.Marshal(rawMappings)
			var mappings []models.ProductStorage
			if err := json.Unmarshal(data, &mappings); err == nil {
				return s.StorageStore.BatchRestoreStock(ctx, log.ResourceID, mappings)
			}
		}
		return nil
	}
	return nil
}

func (s *RevertService) undoSettingAction(ctx context.Context, log *models.AuditLog, _ string, _ bool) error {
	details := log.Details
	// UNDO or REDO of setting is always just restoring "before"
	before, ok := details["before"].(string)
	if !ok {
		return fmt.Errorf("undo failed: 'before' value not captured for this setting")
	}
	return s.SettingsService.Upsert(ctx, log.ResourceID, before)
}

// mapProductToRawMap converts a historical Product object into a map suitable for fn_bulk_upsert_product
func (s *RevertService) mapProductToRawMap(raw interface{}, id string) (map[string]interface{}, error) {
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	var p models.Product
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}

	// Construct a map that exactly matches the JSON keys expected by fn_bulk_upsert_product
	item := map[string]interface{}{
		"id":                 id,
		"name":               p.Name,
		"tcg":                p.TCG,
		"category":           p.Category,
		"set_name":           p.SetName,
		"set_code":           p.SetCode,
		"collector_number":    p.CollectorNumber,
		"condition":          p.Condition,
		"foil_treatment":     p.FoilTreatment,
		"card_treatment":     p.CardTreatment,
		"promo_type":         p.PromoType,
		"price_reference":    p.PriceReference,
		"price_source":       p.PriceSource,
		"price_cop_override": p.PriceCOPOverride,
		"image_url":          p.ImageURL,
		"description":        p.Description,
		"language":           p.Language,
		"color_identity":     p.ColorIdentity,
		"rarity":             p.Rarity,
		"cmc":                p.CMC,
		"is_legendary":       p.IsLegendary,
		"is_historic":        p.IsHistoric,
		"is_land":            p.IsLand,
		"is_basic_land":      p.IsBasicLand,
		"art_variation":      p.ArtVariation,
		"oracle_text":        p.OracleText,
		"artist":             p.Artist,
		"type_line":          p.TypeLine,
		"border_color":       p.BorderColor,
		"frame":              p.Frame,
		"full_art":           p.FullArt,
		"textless":           p.Textless,
		"scryfall_id":        p.ScryfallID,
		"legalities":         p.Legalities,
		"stock":              p.Stock,
	}

	// Categories
	catIDs := make([]string, 0, len(p.Categories))
	for _, c := range p.Categories {
		catIDs = append(catIDs, c.ID)
	}
	item["category_ids"] = catIDs

	// Storage
	storageItems := make([]models.StorageLocation, 0, len(p.StoredIn))
	for _, st := range p.StoredIn {
		storageItems = append(storageItems, models.StorageLocation{
			StorageID: st.StorageID,
			Quantity:  st.Quantity,
		})
	}
	item["storage_items"] = storageItems

	return item, nil
}
