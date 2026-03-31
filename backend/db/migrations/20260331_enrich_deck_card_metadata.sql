-- Enroll enriched card metadata to deck_card
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS foil_treatment TEXT;
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS card_treatment TEXT;
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS rarity         TEXT;
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS art_variation  TEXT;
