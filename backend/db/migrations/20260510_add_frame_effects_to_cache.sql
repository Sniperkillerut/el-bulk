-- Migration: Add frame_effects to external_scryfall cache
ALTER TABLE external_scryfall ADD COLUMN IF NOT EXISTS frame_effects JSONB;
