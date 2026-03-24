-- Migration: Add completed_at to orders, add 'completed' to status enum
-- Run against a live DB that already has the orders table

-- 1. Drop the old CHECK constraint and add new one with 'completed'
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
ALTER TABLE orders ADD CONSTRAINT orders_status_check
  CHECK (status IN ('pending', 'confirmed', 'completed', 'cancelled'));

-- 2. Add completed_at column
ALTER TABLE orders ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ;
