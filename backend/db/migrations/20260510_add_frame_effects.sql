-- Migration: Add frame_effects to product, deck_card, client_request, and bounty
ALTER TABLE product ADD COLUMN frame_effects JSONB;
ALTER TABLE deck_card ADD COLUMN frame_effects JSONB;
ALTER TABLE client_request ADD COLUMN frame_effects JSONB;
ALTER TABLE bounty ADD COLUMN frame_effects JSONB;
