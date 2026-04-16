package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/store"
	"github.com/el-bulk/backend/utils/logger"
)

type RevertService struct {
	AuditStore             *store.AuditStore
	Audit                  Auditer
	ProductStore           *store.ProductStore
	ProductService        *ProductService
	CategoryService       *CategoryService
	StorageLocationService *StorageLocationService
	SettingsService       *SettingsService
}

func NewRevertService(as *store.AuditStore, a Auditer, ps *store.ProductStore, p *ProductService, c *CategoryService, sl *StorageLocationService, s *SettingsService) *RevertService {
	return &RevertService{
		AuditStore:             as,
		Audit:                  a,
		ProductStore:           ps,
		ProductService:        p,
		CategoryService:       c,
		StorageLocationService: sl,
		SettingsService:       s,
	}
}

func (s *RevertService) Undo(ctx context.Context, logID string) error {
	log, err := s.AuditStore.GetByID(ctx, logID)
	if err != nil {
		return fmt.Errorf("audit log not found: %w", err)
	}

	logger.InfoCtx(ctx, "Undoing action: %s on %s (%s)", log.Action, log.ResourceType, log.ResourceID)

	var revertErr error
	var undoAction string

	switch log.ResourceType {
	case "product":
		revertErr = s.undoProductAction(ctx, log)
		undoAction = "UNDO_" + log.Action
	case "category":
		revertErr = s.undoCategoryAction(ctx, log)
		undoAction = "UNDO_" + log.Action
	case "storage":
		revertErr = s.undoStorageAction(ctx, log)
		undoAction = "UNDO_" + log.Action
	case "setting":
		revertErr = s.undoSettingAction(ctx, log)
		undoAction = "UNDO_" + log.Action
	default:
		return fmt.Errorf("undo not supported for resource type: %s", log.ResourceType)
	}

	if revertErr != nil {
		return fmt.Errorf("failed to revert action: %w", revertErr)
	}

	// Log the successful undo as a separate revertible action
	s.Audit.LogAction(ctx, undoAction, log.ResourceType, log.ResourceID, models.JSONB{
		"undone_log_id":   log.ID,
		"original_action": log.Action,
	})

	return nil
}

func (s *RevertService) undoProductAction(ctx context.Context, log *models.AuditLog) error {
	details := log.Details

	switch log.Action {
	case "CREATE_PRODUCT":
		// Undo CREATE -> DELETE
		return s.ProductService.Delete(ctx, log.ResourceID)

	case "UPDATE_PRODUCT", "DELETE_PRODUCT":
		// Both involve restoring a state
		var targetState interface{}
		if log.Action == "UPDATE_PRODUCT" {
			targetState = details["before"]
		} else {
			targetState = details["deleted"]
		}

		if targetState == nil {
			return fmt.Errorf("missing previous state in audit log")
		}

		// Convert historical snapshot to something BulkUpsert understands
		// We explicitly include the ID from the log
		itemMap, err := s.mapProductToRawMap(targetState, log.ResourceID)
		if err != nil {
			return err
		}

		msg, _ := json.Marshal([]interface{}{itemMap})
		_, err = s.ProductStore.BulkUpsert(ctx, string(msg))
		return err
	}

	return fmt.Errorf("undo not implemented for product action: %s", log.Action)
}

func (s *RevertService) undoCategoryAction(ctx context.Context, log *models.AuditLog) error {
	details := log.Details

	switch log.Action {
	case "CREATE_CATEGORY":
		return s.CategoryService.Delete(ctx, log.ResourceID)
	case "UPDATE_CATEGORY":
		before, ok := details["before"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid 'before' state")
		}
		_, err := s.CategoryService.Update(ctx, log.ResourceID, before)
		return err
	case "DELETE_CATEGORY":
		deleted, ok := details["deleted"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid 'deleted' state")
		}
		
		isActive := deleted["is_active"].(bool)
		showBadge := deleted["show_badge"].(bool)
		searchable := deleted["searchable"].(bool)

		input := models.CustomCategoryInput{
			Name:       fmt.Sprintf("%v", deleted["name"]),
			Slug:       fmt.Sprintf("%v", deleted["slug"]),
			IsActive:   &isActive,
			ShowBadge:  &showBadge,
			Searchable: &searchable,
		}
		_, err := s.CategoryService.Create(ctx, input)
		return err
	}
	return nil
}

func (s *RevertService) undoStorageAction(ctx context.Context, log *models.AuditLog) error {
	details := log.Details
	switch log.Action {
	case "CREATE_STORAGE":
		return s.StorageLocationService.Delete(ctx, log.ResourceID)
	case "UPDATE_STORAGE":
		before, ok := details["before"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid 'before' state")
		}
		return s.StorageLocationService.Update(ctx, log.ResourceID, fmt.Sprintf("%v", before["name"]))
	case "DELETE_STORAGE":
		deleted, ok := details["deleted"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid 'deleted' state")
		}
		_, err := s.StorageLocationService.Create(ctx, fmt.Sprintf("%v", deleted["name"]))
		return err
	}
	return nil
}

func (s *RevertService) undoSettingAction(ctx context.Context, log *models.AuditLog) error {
	details := log.Details
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
