-- External CardKingdom Cache Table
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

CREATE INDEX IF NOT EXISTS idx_ext_ck_lookup   ON external_cardkingdom (scryfall_id) WHERE scryfall_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_ext_ck_name     ON external_cardkingdom (LOWER(name));
