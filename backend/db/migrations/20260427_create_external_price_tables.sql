-- Migration: Create tables for external price mirroring
CREATE TABLE IF NOT EXISTS external_scryfall (
    scryfall_id      UUID PRIMARY KEY,
    name             TEXT NOT NULL,
    set_code         TEXT NOT NULL,
    collector_number TEXT NOT NULL,
    price_usd        NUMERIC(12, 4),
    price_usd_foil   NUMERIC(12, 4),
    price_eur        NUMERIC(12, 4),
    image_url        TEXT,
    updated_at       TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS external_cardkingdom (
    ck_id            INTEGER PRIMARY KEY,
    scryfall_id      UUID,
    name             TEXT NOT NULL,
    edition          TEXT NOT NULL,
    variation        TEXT,
    is_foil          BOOLEAN NOT NULL,
    price_retail     NUMERIC(12, 4),
    updated_at       TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ext_scry_lookup ON external_scryfall (LOWER(name), LOWER(set_code), collector_number);
CREATE INDEX IF NOT EXISTS idx_ext_ck_lookup   ON external_cardkingdom (scryfall_id) WHERE scryfall_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_ext_ck_name     ON external_cardkingdom (LOWER(name));
