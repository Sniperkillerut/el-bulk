-- Add header-specific theme colors for improved accessibility and contrast
ALTER TABLE theme ADD COLUMN IF NOT EXISTS accent_header TEXT NOT NULL DEFAULT '#ffffff';
ALTER TABLE theme ADD COLUMN IF NOT EXISTS status_hp_header TEXT NOT NULL DEFAULT '#ffffff';
