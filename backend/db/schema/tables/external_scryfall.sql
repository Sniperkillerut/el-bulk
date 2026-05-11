-- External Scryfall Cache Table
CREATE TABLE IF NOT EXISTS external_scryfall (
    scryfall_id      UUID PRIMARY KEY,
    name             TEXT NOT NULL,
    set_code         TEXT NOT NULL,
    collector_number TEXT NOT NULL,
    price_usd        NUMERIC(12, 4),
    price_usd_foil   NUMERIC(12, 4),
    price_eur        NUMERIC(12, 4),
    image_url        TEXT,
    frame_effects     JSONB,
    updated_at       TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ext_scry_lookup ON external_scryfall (LOWER(name), LOWER(set_code), collector_number);
