package store

import (
	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
	"time"
)

type ThemeStore struct {
	*BaseStore[models.Theme]
}

func NewThemeStore(db *sqlx.DB) *ThemeStore {
	return &ThemeStore{
		BaseStore: NewBaseStore[models.Theme](db, "theme"),
	}
}

func (s *ThemeStore) List() ([]models.Theme, error) {
	return s.BaseStore.List("ORDER BY is_system DESC, name ASC")
}

func (s *ThemeStore) Create(input models.ThemeInput) (*models.Theme, error) {
	var theme models.Theme
	query := `
		INSERT INTO theme (
			name, bg_page, bg_header, bg_surface, bg_card, text_main, text_secondary, text_muted, text_on_accent, text_on_header,
			accent_primary, accent_primary_hover, border_main, border_focus, status_nm, status_lp, status_mp, status_hp, status_dmg,
			btn_primary_bg, btn_primary_text, btn_secondary_bg, btn_secondary_text, checkbox_border, checkbox_checked,
			bg_image_url, font_heading, font_body, accent_secondary
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25,
			$26, $27, $28, $29
		) RETURNING *
	`
	logger.Trace("[DB] Executing CreateTheme: %s | Name: %s", query, input.Name)
	err := s.DB.Get(&theme, query,
		input.Name, input.BgPage, input.BgHeader, input.BgSurface, input.BgCard,
		input.TextMain, input.TextSecondary, input.TextMuted, input.TextOnAccent, input.TextOnHeader,
		input.AccentPrimary, input.AccentPrimaryHover, input.BorderMain, input.BorderFocus,
		input.StatusNM, input.StatusLP, input.StatusMP, input.StatusHP, input.StatusDMG,
		input.BtnPrimaryBg, input.BtnPrimaryText, input.BtnSecondaryBg, input.BtnSecondaryText,
		input.CheckboxBorder, input.CheckboxChecked,
		input.BgImageURL, input.FontHeading, input.FontBody, input.AccentSecondary,
	)
	if err != nil {
		logger.Error("[DB] CreateTheme failed for %s: %v", input.Name, err)
	}
	return &theme, err
}

func (s *ThemeStore) Update(id string, input models.ThemeInput) (*models.Theme, error) {
	var theme models.Theme
	query := `
		UPDATE theme SET
			name = $1, bg_page = $2, bg_header = $3, bg_surface = $4, bg_card = $5,
			text_main = $6, text_secondary = $7, text_muted = $8, text_on_accent = $9, text_on_header = $10,
			accent_primary = $11, accent_primary_hover = $12, border_main = $13, border_focus = $14,
			status_nm = $15, status_lp = $16, status_mp = $17, status_hp = $18, status_dmg = $19,
			btn_primary_bg = $20, btn_primary_text = $21, btn_secondary_bg = $22, btn_secondary_text = $23,
			checkbox_border = $24, checkbox_checked = $25,
			bg_image_url = $26, font_heading = $27, font_body = $28, accent_secondary = $29,
			updated_at = now()
		WHERE id = $30
		RETURNING *
	`
	err := s.DB.Get(&theme, query,
		input.Name, input.BgPage, input.BgHeader, input.BgSurface, input.BgCard,
		input.TextMain, input.TextSecondary, input.TextMuted, input.TextOnAccent, input.TextOnHeader,
		input.AccentPrimary, input.AccentPrimaryHover, input.BorderMain, input.BorderFocus,
		input.StatusNM, input.StatusLP, input.StatusMP, input.StatusHP, input.StatusDMG,
		input.BtnPrimaryBg, input.BtnPrimaryText, input.BtnSecondaryBg, input.BtnSecondaryText,
		input.CheckboxBorder, input.CheckboxChecked,
		input.BgImageURL, input.FontHeading, input.FontBody, input.AccentSecondary,
		id,
	)
	return &theme, err
}

func (s *ThemeStore) IsSystemTheme(id string) (bool, error) {
	start := time.Now()
	var isSystem bool
	query := "SELECT is_system FROM theme WHERE id = $1"
	logger.Trace("[DB] Executing IsSystemTheme for %s: %s", id, query)
	err := s.DB.Get(&isSystem, query, id)
	if err != nil {
		logger.Error("[DB] IsSystemTheme failed for %s: %v", id, err)
	}
	logger.Debug("[DB] IsSystemTheme for %s took %v", id, time.Since(start))
	return isSystem, err
}
