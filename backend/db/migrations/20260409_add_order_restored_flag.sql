-- Add inventory_restored flag to order table
ALTER TABLE "order" ADD COLUMN IF NOT EXISTS inventory_restored BOOLEAN NOT NULL DEFAULT FALSE;
