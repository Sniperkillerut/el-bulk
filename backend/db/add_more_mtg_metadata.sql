-- Add more MTG-specific metadata columns to products table
-- Expanded on 2026-03-24

ALTER TABLE products 
  ADD COLUMN IF NOT EXISTS oracle_text   TEXT,
  ADD COLUMN IF NOT EXISTS flavor_text   TEXT,
  ADD COLUMN IF NOT EXISTS artist        TEXT,
  ADD COLUMN IF NOT EXISTS type_line     TEXT,
  ADD COLUMN IF NOT EXISTS border_color  TEXT,
  ADD COLUMN IF NOT EXISTS frame         TEXT,
  ADD COLUMN IF NOT EXISTS full_art      BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS textless      BOOLEAN NOT NULL DEFAULT false;

-- Create an index for type_line to allow faster filtering by card types in the future
CREATE INDEX IF NOT EXISTS idx_products_type_line ON products(type_line);
