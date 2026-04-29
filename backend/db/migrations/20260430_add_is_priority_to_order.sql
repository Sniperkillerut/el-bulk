-- Migration: Add is_priority to order table
ALTER TABLE "order" ADD COLUMN IF NOT EXISTS is_priority BOOLEAN NOT NULL DEFAULT false;
