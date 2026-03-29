-- Add price_source and price_reference to bounty table
ALTER TABLE bounty ADD COLUMN price_source TEXT NOT NULL DEFAULT 'manual';
ALTER TABLE bounty ADD COLUMN price_reference NUMERIC(12, 2);
