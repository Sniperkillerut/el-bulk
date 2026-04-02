-- Multi-language translation strings
CREATE TABLE IF NOT EXISTS translation (
  key        TEXT NOT NULL,
  locale     TEXT NOT NULL,
  value      TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (key, locale)
);

-- Index for faster key-based lookups
CREATE INDEX IF NOT EXISTS idx_translation_key ON translation (key);

-- Index for faster key-based lookups
CREATE INDEX IF NOT EXISTS idx_translation_key ON translation (key);
