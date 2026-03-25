-- Migration to drop flavor_text column
-- Approved on 2026-03-24

ALTER TABLE products DROP COLUMN IF EXISTS flavor_text;
