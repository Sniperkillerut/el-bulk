-- Enrich deck_card with full MTG metadata
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS set_name          TEXT;
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS language          TEXT NOT NULL DEFAULT 'en';
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS color_identity    TEXT;
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS cmc               NUMERIC(5, 1);
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS is_legendary      BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS is_historic       BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS is_land           BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS is_basic_land     BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS oracle_text       TEXT;
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS artist            TEXT;
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS border_color      TEXT;
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS frame             TEXT;
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS full_art          BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS textless          BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE deck_card ADD COLUMN IF NOT EXISTS promo_type        TEXT;
