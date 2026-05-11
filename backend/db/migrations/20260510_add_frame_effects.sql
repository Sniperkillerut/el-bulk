-- Migration: Add frame_effects to product, deck_card, client_request, and bounty
ALTER TABLE product ADD COLUMN IF NOT EXISTS frame_effects JSONB;
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS frame_effects JSONB;
ALTER TABLE client_request ADD COLUMN IF NOT EXISTS frame_effects JSONB;
ALTER TABLE bounty ADD COLUMN IF NOT EXISTS frame_effects JSONB;
