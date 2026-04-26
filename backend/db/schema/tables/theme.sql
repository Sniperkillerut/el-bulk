-- Dynamic Themes Table
CREATE TABLE IF NOT EXISTS theme (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name                TEXT NOT NULL,
  is_system           BOOLEAN NOT NULL DEFAULT false,
  
  -- Semantic Colors
  bg_page             TEXT NOT NULL,
  bg_header           TEXT NOT NULL,
  bg_surface          TEXT NOT NULL,
  
  text_main           TEXT NOT NULL,
  text_secondary      TEXT NOT NULL,
  text_muted          TEXT NOT NULL,
  text_on_accent      TEXT NOT NULL,
  
  accent_primary      TEXT NOT NULL,
  accent_primary_hover TEXT NOT NULL,
  border_main         TEXT NOT NULL,
  
  -- Status Colors
  status_nm           TEXT NOT NULL,
  status_lp           TEXT NOT NULL,
  status_mp           TEXT NOT NULL,
  status_hp           TEXT NOT NULL,
  status_dmg          TEXT NOT NULL DEFAULT '#455a64',
  
  -- Specialized Colors
  bg_card             TEXT NOT NULL DEFAULT '#ffffff',
  text_on_header      TEXT NOT NULL DEFAULT '#ffffff',
  border_focus        TEXT NOT NULL DEFAULT '#3b3127',
  
  -- Context-Specific Header Colors
  accent_header       TEXT NOT NULL DEFAULT '#ffffff',
  status_hp_header    TEXT NOT NULL DEFAULT '#ffffff',
  
  -- Interactive Elements
  btn_primary_bg      TEXT NOT NULL DEFAULT '#1a1f2e',
  btn_primary_text    TEXT NOT NULL DEFAULT '#ffffff',
  btn_secondary_bg    TEXT NOT NULL DEFAULT 'transparent',
  btn_secondary_text  TEXT NOT NULL DEFAULT '#3b3127',
  
  checkbox_border     TEXT NOT NULL DEFAULT '#8b795c',
  checkbox_checked    TEXT NOT NULL DEFAULT '#d4af37',
  
  -- Layout & Geometry
  radius_base         TEXT NOT NULL DEFAULT '8px',
  padding_card        TEXT NOT NULL DEFAULT '12px',
  gap_grid            TEXT NOT NULL DEFAULT '24px',
  
  -- Advanced Branding Extensions
  bg_image_url        TEXT,
  font_heading        TEXT,
  font_body           TEXT,
  accent_secondary    TEXT,
  accent_rose         TEXT,
  
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
