-- Add confirmed_at to order table
ALTER TABLE "order" ADD COLUMN IF NOT EXISTS confirmed_at TIMESTAMPTZ;
