-- TCG set metadata for ordering and filtering
CREATE TABLE IF NOT EXISTS tcg_set (
    tcg         TEXT NOT NULL,
    code        TEXT NOT NULL,
    name        TEXT NOT NULL,
    released_at DATE,
    set_type    TEXT,
    ck_name     TEXT,
    PRIMARY KEY (tcg, code)
);

CREATE INDEX IF NOT EXISTS idx_tcg_set_name ON tcg_set(LOWER(name));
