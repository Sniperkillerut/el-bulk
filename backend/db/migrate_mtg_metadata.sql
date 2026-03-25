-- Migration: Add MTG Metadata columns to products table
-- Run this if you have an existing database and want to enable MTG metadata support.

ALTER TABLE products ADD COLUMN IF NOT EXISTS language TEXT NOT NULL DEFAULT 'en';
ALTER TABLE products ADD COLUMN IF NOT EXISTS color TEXT;
ALTER TABLE products ADD COLUMN IF NOT EXISTS rarity TEXT;
ALTER TABLE products ADD COLUMN IF NOT EXISTS cmc NUMERIC(5, 1);
ALTER TABLE products ADD COLUMN IF NOT EXISTS is_legendary BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE products ADD COLUMN IF NOT EXISTS is_historic BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE products ADD COLUMN IF NOT EXISTS is_land BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE products ADD COLUMN IF NOT EXISTS is_basic_land BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE products ADD COLUMN IF NOT EXISTS art_variation TEXT;
