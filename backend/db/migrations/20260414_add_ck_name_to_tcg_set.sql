-- Add cardkingdom_name to tcg_set for data-driven normalization
ALTER TABLE tcg_set ADD COLUMN IF NOT EXISTS ck_name TEXT;

-- Update the view to include this if needed (optional, but good for visibility)
-- No changes to view_product_enriched needed yet unless we want to show CK set name in admin.
