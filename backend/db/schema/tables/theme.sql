-- Dynamic Themes Table
CREATE TABLE IF NOT EXISTS theme (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name                TEXT NOT NULL,
  is_system           BOOLEAN NOT NULL DEFAULT false,
  
  -- Semantic Colors (Hex or HSL)
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
  
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Default Theme Base Data (Cardboard)
INSERT INTO theme (
  id, name, is_system, 
  bg_page, bg_header, bg_surface, bg_card,
  text_main, text_secondary, text_muted, text_on_accent, text_on_header,
  accent_primary, accent_primary_hover, border_main, border_focus,
  status_nm, status_lp, status_mp, status_hp, status_dmg,
  btn_primary_bg, btn_primary_text, btn_secondary_bg, btn_secondary_text,
  checkbox_border, checkbox_checked,
  radius_base, padding_card, gap_grid,
  accent_header, status_hp_header
)
VALUES (
  '00000000-0000-0000-0000-000000000001', 'Cardboard', true, 
  '#e6dac3', '#1a1f2e', '#fdfbf7', '#ffffff',
  '#3b3127', '#5c4e3d', '#8b795c', '#2c251d', '#ffffff',
  '#d4af37', '#b8961e', '#d4c5ab', '#3b3127',
  '#2e7d32', '#558b2f', '#ef6c00', '#c62828', '#455a64',
  '#1a1f2e', '#ffffff', 'transparent', '#3b3127',
  '#8b795c', '#d4af37',
  '8px', '12px', '16px',
  '#d4af37', '#c62828'
) ON CONFLICT DO NOTHING;

-- Obsidiana (Professional Dark)
INSERT INTO theme (
  id, name, is_system, 
  bg_page, bg_header, bg_surface, bg_card,
  text_main, text_secondary, text_muted, text_on_accent, text_on_header,
  accent_primary, accent_primary_hover, border_main, border_focus,
  status_nm, status_lp, status_mp, status_hp, status_dmg,
  accent_header, status_hp_header
  btn_primary_bg, btn_primary_text, btn_secondary_bg, btn_secondary_text,
  checkbox_border, checkbox_checked,
  radius_base, padding_card, gap_grid
)
VALUES (
  '00000000-0000-0000-0000-000000000002', 'Obsidiana', true, 
  '#0a0a0a', '#121212', '#1a1a1a', '#1a1a1a',
  '#f8fafc', '#94a3b8', '#475569', '#ffffff', '#ffffff',
  '#3b82f6', '#2563eb', '#334155', '#3b82f6',
  '#10b981', '#fbbf24', '#f59e0b', '#ef4444', '#64748b',
  '#60a5fa', '#ef4444',
  '#3b82f6', '#ffffff', 'transparent', '#f8fafc',
  '#475569', '#3b82f6',
  '2px', '14px', '20px'
) ON CONFLICT DO NOTHING;

-- Yule (Christmas)
INSERT INTO theme (
  id, name, is_system, 
  bg_page, bg_header, bg_surface, bg_card,
  text_main, text_secondary, text_muted, text_on_accent, text_on_header,
  accent_primary, accent_primary_hover, border_main, border_focus,
  status_nm, status_lp, status_mp, status_hp, status_dmg,
  accent_header, status_hp_header
  btn_primary_bg, btn_primary_text, btn_secondary_bg, btn_secondary_text,
  checkbox_border, checkbox_checked,
  radius_base, padding_card, gap_grid
)
VALUES (
  '00000000-0000-0000-0000-000000000003', 'Yule', true, 
  '#052e16', '#991b1b', '#064e3b', '#064e3b',
  '#f0fdf4', '#bbf7d0', '#166534', '#ffffff', '#ffffff',
  '#fbbf24', '#f59e0b', '#14532d', '#f59e0b',
  '#4ade80', '#fbbf24', '#f97316', '#ef4444', '#991b1b',
  '#fde68a', '#f87171',
  '#991b1b', '#ffffff', 'transparent', '#f0fdf4',
  '#8b795c', '#d4af37',
  '8px', '12px', '16px'
) ON CONFLICT DO NOTHING;

-- Spring Egg (Easter)
INSERT INTO theme (
  id, name, is_system, 
  bg_page, bg_header, bg_surface, bg_card,
  text_main, text_secondary, text_muted, text_on_accent, text_on_header,
  accent_primary, accent_primary_hover, border_main, border_focus,
  status_nm, status_lp, status_mp, status_hp, status_dmg,
  accent_header, status_hp_header
  btn_primary_bg, btn_primary_text, btn_secondary_bg, btn_secondary_text,
  checkbox_border, checkbox_checked,
  radius_base, padding_card, gap_grid
)
VALUES (
  '00000000-0000-0000-0000-000000000004', 'Spring Egg', true, 
  '#fffbea', '#f5f3ff', '#ffffff', '#ffffff',
  '#4c1d95', '#7c3aed', '#a78bfa', '#ffffff', '#4c1d95',
  '#8b5cf6', '#a78bfa', '#f3f4f6', '#a78bfa',
  '#10b981', '#fbbf24', '#f59e0b', '#ef4444', '#94a3b8',
  '#5b21b6', '#dc2626',
  '#8b5cf6', '#ffffff', 'transparent', '#4c1d95',
  '#a78bfa', '#8b5cf6',
  '24px', '16px', '24px'
) ON CONFLICT DO NOTHING;

-- Neon Flux (New Release)
INSERT INTO theme (
  id, name, is_system, 
  bg_page, bg_header, bg_surface, bg_card,
  text_main, text_secondary, text_muted, text_on_accent, text_on_header,
  accent_primary, accent_primary_hover, border_main, border_focus,
  status_nm, status_lp, status_mp, status_hp, status_dmg,
  accent_header, status_hp_header
  btn_primary_bg, btn_primary_text, btn_secondary_bg, btn_secondary_text,
  checkbox_border, checkbox_checked,
  radius_base, padding_card, gap_grid
)
VALUES (
  '00000000-0000-0000-0000-000000000005', 'Neon Flux', true, 
  '#020617', '#0f172a', '#020617', '#020617',
  '#f8fafc', '#64748b', '#334155', '#000000', '#ffffff',
  '#22c55e', '#4ade80', '#1e293b', '#4ade80',
  '#22c55e', '#eab308', '#f97316', '#ef4444', '#334155',
  '#4ade80', '#ef4444',
  '#22c55e', '#000000', 'transparent', '#f8fafc',
  '#334155', '#22c55e',
  '0px', '10px', '12px'
) ON CONFLICT DO NOTHING;

-- Arena (Tournament)
INSERT INTO theme (
  id, name, is_system, 
  bg_page, bg_header, bg_surface, bg_card,
  text_main, text_secondary, text_muted, text_on_accent, text_on_header,
  accent_primary, accent_primary_hover, border_main, border_focus,
  status_nm, status_lp, status_mp, status_hp, status_dmg,
  accent_header, status_hp_header
  btn_primary_bg, btn_primary_text, btn_secondary_bg, btn_secondary_text,
  checkbox_border, checkbox_checked,
  radius_base, padding_card, gap_grid
)
VALUES (
  '00000000-0000-0000-0000-000000000006', 'Arena', true, 
  '#171717', '#262626', '#1c1c1c', '#1c1c1c',
  '#ffffff', '#a3a3a3', '#525252', '#ffffff', '#ffffff',
  '#ea580c', '#f97316', '#404040', '#f97316',
  '#22c55e', '#eab308', '#f97316', '#dc2626', '#404040',
  '#f87171', '#fca5a5',
  '#ea580c', '#ffffff', 'transparent', '#ffffff',
  '#525252', '#ea580c',
  '4px', '14px', '16px'
) ON CONFLICT DO NOTHING;

-- Celebrate (Birthday)
INSERT INTO theme (
  id, name, is_system, 
  bg_page, bg_header, bg_surface, bg_card,
  text_main, text_secondary, text_muted, text_on_accent, text_on_header,
  accent_primary, accent_primary_hover, border_main, border_focus,
  status_nm, status_lp, status_mp, status_hp, status_dmg,
  accent_header, status_hp_header
  btn_primary_bg, btn_primary_text, btn_secondary_bg, btn_secondary_text,
  checkbox_border, checkbox_checked,
  radius_base, padding_card, gap_grid
)
VALUES (
  '00000000-0000-0000-0000-000000000007', 'Celebrate', true, 
  '#fdf2f8', '#be185d', '#ffffff', '#ffffff',
  '#831843', '#db2777', '#f472b6', '#ffffff', '#ffffff',
  '#db2777', '#f472b6', '#fce7f3', '#f472b6',
  '#10b981', '#fbbf24', '#f59e0b', '#ef4444', '#db2777',
  '#fdf2f8', '#fce7f3',
  '#be185d', '#ffffff', 'transparent', '#831843',
  '#f472b6', '#be185d',
  '16px', '14px', '20px'
) ON CONFLICT DO NOTHING;
